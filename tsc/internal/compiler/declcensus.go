package compiler

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/microsoft/TypeScript/tsc/internal/checker"
	"github.com/microsoft/TypeScript/tsc/internal/json"
)

// Declaration census: the cross-checker join of the frames that checker.DeclCensus records
// (research instrumentation for the G1 gate, research/agents/20260913T150256-fable-5.1.md
// sections 1 and 2; enabled by TSGO_DECL_CENSUS=<dir>). All checkers of a program run in this
// process, so once forEachCheckerGroupDo has waited for every checker goroutine the frames are
// joined here by key: how much of the work is repeated across checkers and how much of the
// repetition is removable, when the repeated results would have been available, the work and
// span of the dependency DAG with list-scheduling makespans, the coverage of the first sound
// subsets S0 and S1, guard sensitivity and the constructor-category counts behind
// materialisation. The join is written as <dir>/census.json, every checker's own accounting as
// <dir>/checker-<i>.json (its exclusive frame time next to its measured group-task span, so the
// attribution can be checked), and one decl-census line with the headline numbers goes to
// stderr for the benchmark harness. Nothing is recorded when the variable is unset.

// declCensusPhase is the measured span of one checker's group tasks, for reconciling the
// exclusive frame times with the check phase. Times are nanoseconds since the census epoch.
type declCensusPhase struct {
	StartNs int64 `json:"startNs"` // start of the first group task
	EndNs   int64 `json:"endNs"`   // end of the last group task
	SpanNs  int64 `json:"spanNs"`  // sum of the group task spans
}

// declCensusRun is the census of one program: the epoch every checker's times are relative to,
// the per-checker censuses and their measured phases. A nil run records nothing.
type declCensusRun struct {
	dir      string
	epoch    time.Time
	censuses []*checker.DeclCensus
	phases   []declCensusPhase
}

func newDeclCensusRun(checkers []*checker.Checker) *declCensusRun {
	dir := os.Getenv("TSGO_DECL_CENSUS")
	if dir == "" {
		return nil
	}
	run := &declCensusRun{dir: dir, epoch: time.Now(), censuses: make([]*checker.DeclCensus, len(checkers)), phases: make([]declCensusPhase, len(checkers))}
	for i, c := range checkers {
		run.censuses[i] = c.EnableDeclCensus(run.epoch)
	}
	return run
}

func (r *declCensusRun) begin() int64 {
	if r == nil {
		return 0
	}
	return time.Since(r.epoch).Nanoseconds()
}

// end records one group task of a checker; only that checker's goroutine touches its phase.
func (r *declCensusRun) end(checkerIdx int, began int64) {
	if r == nil {
		return
	}
	now := time.Since(r.epoch).Nanoseconds()
	phase := &r.phases[checkerIdx]
	if phase.EndNs == 0 {
		phase.StartNs = began
	}
	phase.EndNs = now
	phase.SpanNs += now - began
}

// report joins the censuses and writes the files once every checker goroutine has finished. It
// runs after every group, so the files always hold everything recorded so far.
func (r *declCensusRun) report(program string, w io.Writer) {
	if r == nil {
		return
	}
	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		fmt.Fprintf(w, "decl-census error: %v\n", err)
		return
	}
	report := declCensusReport{Program: program}
	for i, cs := range r.censuses {
		checkerReport := newDeclCensusCheckerReport(program, i, cs, r.phases[i])
		report.Checkers = append(report.Checkers, checkerReport)
		if err := writeDeclCensusJSON(filepath.Join(r.dir, fmt.Sprintf("checker-%d.json", i)), checkerReport); err != nil {
			fmt.Fprintf(w, "decl-census error: %v\n", err)
			return
		}
	}
	report.analyse(joinDeclCensus(r.censuses))
	file := filepath.Join(r.dir, "census.json")
	if err := writeDeclCensusJSON(file, report); err != nil {
		fmt.Fprintf(w, "decl-census error: %v\n", err)
		return
	}
	line, err := json.Marshal(declCensusHeadline{Program: program, SummaryOnly: true, File: file, Summary: report.Summary})
	if err != nil {
		fmt.Fprintf(w, "decl-census error: %v\n", err)
		return
	}
	fmt.Fprintf(w, "decl-census %s\n", line)
}

func writeDeclCensusJSON(path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Output.

type declCensusReport struct {
	Program         string                    `json:"program"`
	Summary         declCensusSummary         `json:"summary"`
	Checkers        []declCensusCheckerReport `json:"checkers"`
	Repetition      declCensusRepetition      `json:"repetition"`
	DAG             declCensusDAG             `json:"dag"`
	Coverage        declCensusCoverage        `json:"coverage"`
	Guard           declCensusGuard           `json:"guard"`
	Materialisation []declCensusFrameTotals   `json:"materialisation"` // per kind of S1 frame (and "all"): the objects the frames created next to their exclusive time; a measured per-object cost turns the counts into the share of the time a recipe store could not save
	Heaviest        []declCensusHeavyKey      `json:"heaviest"`
}

// declCensusHeadline is the stderr line: the summary without the tables.
type declCensusHeadline struct {
	Program     string            `json:"program"`
	SummaryOnly bool              `json:"summaryOnly"`
	File        string            `json:"file"`
	Summary     declCensusSummary `json:"summary"`
}

// declCensusSummary holds the headline numbers; the shares are of the exclusive frame time
// unless named otherwise.
type declCensusSummary struct {
	Checkers              int     `json:"checkers"`
	Frames                int     `json:"frames"`
	Truncated             bool    `json:"truncated"`
	PhaseNs               int64   `json:"phaseNs"`               // first group task start to last group task end over all checkers
	GroupSpanNs           int64   `json:"groupSpanNs"`           // sum of the measured group task spans of all checkers
	ExclusiveNs           int64   `json:"exclusiveNs"`           // sum of the exclusive times of all frames
	ExclusiveShareOfGroup float64 `json:"exclusiveShareOfGroup"` // ExclusiveNs / GroupSpanNs: what the frames account for
	RemovableNs           int64   `json:"removableNs"`
	RemovableShare        float64 `json:"removableShare"`
	ClosedRemovableNs     int64   `json:"closedRemovableNs"`
	ClosedRemovableShare  float64 `json:"closedRemovableShare"` // abandon criterion A1
	AvailableShare        float64 `json:"availableShare"`       // of the repeated time: the producer had finished before the consumer started
	UniqueWorkNs          int64   `json:"uniqueWorkNs"`
	SpanNs                int64   `json:"spanNs"`
	WorkSpanRatio         float64 `json:"workSpanRatio"` // A2
	Makespan4Ns           int64   `json:"makespan4Ns"`
	Makespan8Ns           int64   `json:"makespan8Ns"`
	S0Share               float64 `json:"s0Share"`             // of the removable time
	S1Share               float64 `json:"s1Share"`             // A4
	GuardSensitiveShare   float64 `json:"guardSensitiveShare"` // A5
}

// declCensusCheckerReport is one checker's own accounting, from its frames before the join.
type declCensusCheckerReport struct {
	Program               string                  `json:"program"`
	Checker               int                     `json:"checker"`
	Frames                int                     `json:"frames"`
	Truncated             bool                    `json:"truncated"`
	Phase                 declCensusPhase         `json:"phase"`
	ExclusiveNs           int64                   `json:"exclusiveNs"`           // sum of the exclusive times of all frames
	ExclusiveInPhaseNs    int64                   `json:"exclusiveInPhaseNs"`    // frames that ran inside the group tasks
	ExclusiveShareOfPhase float64                 `json:"exclusiveShareOfPhase"` // ExclusiveInPhaseNs / Phase.SpanNs
	ByKind                []declCensusFrameTotals `json:"byKind"`
	ByClass               []declCensusFrameTotals `json:"byClass"`
}

// declCensusFrameTotals sums frames: exclusive time and the objects they created.
type declCensusFrameTotals struct {
	Name           string `json:"name"`
	Frames         int    `json:"frames"`
	ExclNs         int64  `json:"exclNs"`
	Types          uint64 `json:"types"`
	Symbols        uint64 `json:"symbols"`
	Signatures     uint64 `json:"signatures"`
	Instantiations uint64 `json:"instantiations"`
}

func (t *declCensusFrameTotals) add(f *checker.DeclCensusFrame) {
	t.Frames++
	t.ExclNs += f.ExclNs
	t.Types += uint64(f.Types)
	t.Symbols += uint64(f.Symbols)
	t.Signatures += uint64(f.Signatures)
	t.Instantiations += uint64(f.Instantiations)
}

// declCensusCost sums keys: the cost over all checkers that computed the key and the removable
// part of it, the sum minus the cheapest checker's cost.
type declCensusCost struct {
	Keys                    int    `json:"keys"`
	SumNs                   int64  `json:"sumNs"`
	RemovableNs             int64  `json:"removableNs"`
	SumTypes                uint64 `json:"sumTypes"`
	RemovableTypes          uint64 `json:"removableTypes"`
	SumSymbols              uint64 `json:"sumSymbols"`
	RemovableSymbols        uint64 `json:"removableSymbols"`
	SumSignatures           uint64 `json:"sumSignatures"`
	RemovableSignatures     uint64 `json:"removableSignatures"`
	SumInstantiations       uint64 `json:"sumInstantiations"`
	RemovableInstantiations uint64 `json:"removableInstantiations"`
}

func (c *declCensusCost) add(k *declCensusKeyCost) {
	c.Keys++
	c.SumNs += k.sumNs
	c.RemovableNs += k.removableNs()
	c.SumTypes += k.sum[0]
	c.RemovableTypes += k.sum[0] - k.min[0]
	c.SumSymbols += k.sum[1]
	c.RemovableSymbols += k.sum[1] - k.min[1]
	c.SumSignatures += k.sum[2]
	c.RemovableSignatures += k.sum[2] - k.min[2]
	c.SumInstantiations += k.sum[3]
	c.RemovableInstantiations += k.sum[3] - k.min[3]
}

type declCensusNamedCost struct {
	Name string `json:"name"`
	declCensusCost
}

type declCensusRepetition struct {
	Total              declCensusCost           `json:"total"`
	ByMultiplicity     []declCensusMultiplicity `json:"byMultiplicity"`
	Availability       declCensusAvailability   `json:"availability"`       // all repeated keys
	ClosedAvailability declCensusAvailability   `json:"closedAvailability"` // repeated keys without poison flags
}

// declCensusMultiplicity is the cost of the keys computed by exactly Checkers checkers, by the
// class of their flags and by kind.
type declCensusMultiplicity struct {
	Checkers int                   `json:"checkers"`
	Total    declCensusCost        `json:"total"`
	ByClass  []declCensusNamedCost `json:"byClass"`
	ByKind   []declCensusNamedCost `json:"byKind"`
}

// declCensusAvailability splits the exclusive time of the consumers of repeated keys, the
// checkers other than the one that finished the key first (the producer), by when the producer
// finished relative to each consumer: what a store filled by the first finisher could have
// saved, and when.
type declCensusAvailability struct {
	RepeatedNs      int64 `json:"repeatedNs"`      // exclusive time of all consumers; at most the removable time, less when the producer was not the cheapest checker
	AvailableNs     int64 `json:"availableNs"`     // the producer had finished before the consumer started
	InProgressNs    int64 `json:"inProgressNs"`    // the producer was still computing when the consumer started
	ConsumerFirstNs int64 `json:"consumerFirstNs"` // the consumer started before the producer
}

type declCensusDAG struct {
	Nodes           int     `json:"nodes"`
	Edges           int     `json:"edges"`
	CollapsedCycles int     `json:"collapsedCycles"` // strongly connected components of more than one key
	UniqueWorkNs    int64   `json:"uniqueWorkNs"`    // sum of the cheapest checker's exclusive time per key
	SpanNs          int64   `json:"spanNs"`          // longest dependency path by the same weights
	WorkSpanRatio   float64 `json:"workSpanRatio"`
	Makespan4Ns     int64   `json:"makespan4Ns"` // list scheduling over the DAG alone
	Makespan8Ns     int64   `json:"makespan8Ns"`
}

type declCensusCoverage struct {
	RemovableNs       int64            `json:"removableNs"`
	ClosedRemovableNs int64            `json:"closedRemovableNs"`
	S0                declCensusSubset `json:"s0"` // declaration-closed declared types, members, base types, signatures, return types of non-generic declarations, and variances
	S1                declCensusSubset `json:"s1"` // S0 plus declaration-closed instantiations
	S2                declCensusSubset `json:"s2"` // every closed key
}

type declCensusSubset struct {
	Keys                   int     `json:"keys"`
	RemovableNs            int64   `json:"removableNs"`
	ShareOfRemovable       float64 `json:"shareOfRemovable"`
	ShareOfClosedRemovable float64 `json:"shareOfClosedRemovable"`
}

// declCensusGuard reports the S1 keys whose instantiation count or depth differed enough between
// checkers, or was large enough, for an instantiation guard to fire in one checker and not in
// another.
type declCensusGuard struct {
	RepeatedS1Keys        int     `json:"repeatedS1Keys"`
	SensitiveKeys         int     `json:"sensitiveKeys"`
	RepeatedS1RemovableNs int64   `json:"repeatedS1RemovableNs"`
	SensitiveRemovableNs  int64   `json:"sensitiveRemovableNs"`
	Share                 float64 `json:"share"`
}

type declCensusHeavyKey struct {
	Key         string                `json:"key"`
	Kind        string                `json:"kind"`
	Flags       []string              `json:"flags"`
	Checkers    int                   `json:"checkers"`
	Description string                `json:"description"`
	RemovableNs int64                 `json:"removableNs"`
	PerChecker  []declCensusFrameCost `json:"perChecker"`
}

type declCensusFrameCost struct {
	Checker            int    `json:"checker"`
	ExclNs             int64  `json:"exclNs"`
	Types              uint32 `json:"types"`
	Symbols            uint32 `json:"symbols"`
	Signatures         uint32 `json:"signatures"`
	Instantiations     uint32 `json:"instantiations"`
	InclInstantiations uint32 `json:"inclInstantiations"`
	Depth              uint32 `json:"depth"`
	StartNs            int64  `json:"startNs"`
	EndNs              int64  `json:"endNs"`
}

// Flag classes, in the precedence of the report's table: a frame with several poison flags is
// counted under the first.
var declCensusClassNames = [...]string{"closed", "srcExpr", "diag", "guard", "cycle", "unkeyable"}

func declCensusClass(flags checker.DeclCensusFlags) int {
	switch {
	case flags&checker.DeclCensusFlagsSrcExpr != 0:
		return 1
	case flags&checker.DeclCensusFlagsDiag != 0:
		return 2
	case flags&checker.DeclCensusFlagsGuard != 0:
		return 3
	case flags&checker.DeclCensusFlagsCycle != 0:
		return 4
	case flags&checker.DeclCensusFlagsUnkeyable != 0:
		return 5
	}
	return 0
}

// declCensusSubsets says whether a key belongs to S0, S1 and S2 (section 2c of the report).
// S0 is the declaration-level store: declaration-closed frames of the declaration kinds for
// non-generic declarations, and variances, which only exist for generic declarations but are
// keyed by the declaration alone. S1 adds the declaration-closed instantiation frames, whose
// key material (target, arguments, alias) is then entirely made of declaration-file nodes and
// keyable types. S2 is every closed key.
func declCensusSubsets(kind checker.DeclCensusKind, flags checker.DeclCensusFlags) (s0, s1, s2 bool) {
	if flags&checker.DeclCensusPoisonFlags != 0 {
		return false, false, false
	}
	declarationClosed := flags&checker.DeclCensusFlagsSourceKey == 0
	switch kind {
	case checker.DeclCensusKindDeclaredType, checker.DeclCensusKindDeclaredMembers, checker.DeclCensusKindBaseTypes,
		checker.DeclCensusKindResolvedSignature, checker.DeclCensusKindResolvedReturnType:
		s0 = declarationClosed && flags&checker.DeclCensusFlagsGeneric == 0
	case checker.DeclCensusKindVariances:
		s0 = declarationClosed
	case checker.DeclCensusKindObjectInstantiation, checker.DeclCensusKindConditionalInstantiation, checker.DeclCensusKindAliasInstantiation,
		checker.DeclCensusKindInstantiatedType, checker.DeclCensusKindReferenceMembers:
		s1 = declarationClosed
	}
	return s0, s0 || s1, true
}

// Per-checker accounting.

func newDeclCensusCheckerReport(program string, idx int, cs *checker.DeclCensus, phase declCensusPhase) declCensusCheckerReport {
	report := declCensusCheckerReport{Program: program, Checker: idx, Frames: len(cs.Frames), Truncated: cs.Truncated, Phase: phase}
	byKind := make([]declCensusFrameTotals, checker.DeclCensusKindCount)
	byClass := make([]declCensusFrameTotals, len(declCensusClassNames))
	for i := range cs.Frames {
		f := &cs.Frames[i]
		report.ExclusiveNs += f.ExclNs
		if f.Start >= phase.StartNs && f.End <= phase.EndNs {
			report.ExclusiveInPhaseNs += f.ExclNs
		}
		byKind[f.Kind].add(f)
		byClass[declCensusClass(f.Flags)].add(f)
	}
	report.ExclusiveShareOfPhase = declCensusShare(report.ExclusiveInPhaseNs, phase.SpanNs)
	for kind := range byKind {
		byKind[kind].Name = checker.DeclCensusKind(kind).String()
	}
	for class := range byClass {
		byClass[class].Name = declCensusClassNames[class]
	}
	report.ByKind = byKind
	report.ByClass = byClass
	return report
}

func declCensusShare(part, whole int64) float64 {
	if whole == 0 {
		return 0
	}
	return float64(part) / float64(whole)
}

// Join.

// declCensusJoin indexes the frames of all checkers by key: node i is the i-th distinct key and
// frames[i*checkers+c] is the frame that checker c recorded for it, or -1.
type declCensusJoin struct {
	censuses []*checker.DeclCensus
	index    map[checker.CacheHashKey]int32
	kinds    []checker.DeclCensusKind
	flags    []checker.DeclCensusFlags // union over the checkers
	frames   []int32
}

func joinDeclCensus(censuses []*checker.DeclCensus) *declCensusJoin {
	total := 0
	for _, cs := range censuses {
		total += len(cs.Frames)
	}
	j := &declCensusJoin{censuses: censuses, index: make(map[checker.CacheHashKey]int32, total/2)}
	k := len(censuses)
	for ci, cs := range censuses {
		for fi := range cs.Frames {
			f := &cs.Frames[fi]
			node, ok := j.index[f.Key]
			if !ok {
				node = int32(len(j.kinds))
				j.index[f.Key] = node
				j.kinds = append(j.kinds, f.Kind)
				j.flags = append(j.flags, 0)
				for range k {
					j.frames = append(j.frames, -1)
				}
			}
			j.flags[node] |= f.Flags
			slot := &j.frames[int(node)*k+ci]
			if *slot < 0 {
				*slot = int32(fi)
			} else {
				foldDeclCensusFrame(&cs.Frames[*slot], f)
			}
		}
	}
	return j
}

// foldDeclCensusFrame adds a later frame of the same key in the same checker (a slot that was
// not written, or a computation restarted after a cycle) into the first one, so that every
// (checker, key) pair has one cost.
func foldDeclCensusFrame(into, f *checker.DeclCensusFrame) {
	into.ExclNs += f.ExclNs
	into.Types += f.Types
	into.Symbols += f.Symbols
	into.Signatures += f.Signatures
	into.Instantiations += f.Instantiations
	into.InclInstantiations = max(into.InclInstantiations, f.InclInstantiations)
	into.Depth = max(into.Depth, f.Depth)
	into.End = max(into.End, f.End)
	into.Flags |= f.Flags
}

func (j *declCensusJoin) nodes() int {
	return len(j.kinds)
}

// frame returns the frame checker c recorded for node, or nil.
func (j *declCensusJoin) frame(node int32, c int) *checker.DeclCensusFrame {
	fi := j.frames[int(node)*len(j.censuses)+c]
	if fi < 0 {
		return nil
	}
	return &j.censuses[c].Frames[fi]
}

// declCensusKeyCost is the cost of one key over the checkers that computed it: the sum and the
// minimum of the exclusive time and of the created types, symbols, signatures and
// instantiations.
type declCensusKeyCost struct {
	checkers int
	sumNs    int64
	minNs    int64
	sum      [4]uint64
	min      [4]uint64
}

func (k *declCensusKeyCost) removableNs() int64 {
	return k.sumNs - k.minNs
}

func (j *declCensusJoin) keyCost(node int32) declCensusKeyCost {
	var k declCensusKeyCost
	for c := range j.censuses {
		f := j.frame(node, c)
		if f == nil {
			continue
		}
		counters := [4]uint64{uint64(f.Types), uint64(f.Symbols), uint64(f.Signatures), uint64(f.Instantiations)}
		if k.checkers == 0 {
			k.minNs = f.ExclNs
			k.min = counters
		} else {
			k.minNs = min(k.minNs, f.ExclNs)
			for i, v := range counters {
				k.min[i] = min(k.min[i], v)
			}
		}
		k.checkers++
		k.sumNs += f.ExclNs
		for i, v := range counters {
			k.sum[i] += v
		}
	}
	return k
}

// Analysis.

func (r *declCensusReport) analyse(j *declCensusJoin) {
	k := len(j.censuses)
	summary := &r.Summary
	summary.Checkers = k
	var phaseStart, phaseEnd int64
	for i, cs := range j.censuses {
		summary.Frames += len(cs.Frames)
		summary.Truncated = summary.Truncated || cs.Truncated
		summary.ExclusiveNs += r.Checkers[i].ExclusiveNs
		phase := r.Checkers[i].Phase
		summary.GroupSpanNs += phase.SpanNs
		if phase.EndNs != 0 {
			if phaseStart == 0 || phase.StartNs < phaseStart {
				phaseStart = phase.StartNs
			}
			phaseEnd = max(phaseEnd, phase.EndNs)
		}
	}
	summary.PhaseNs = phaseEnd - phaseStart
	summary.ExclusiveShareOfGroup = declCensusShare(summary.ExclusiveNs, summary.GroupSpanNs)

	repetition := &r.Repetition
	repetition.ByMultiplicity = make([]declCensusMultiplicity, k)
	for m := range repetition.ByMultiplicity {
		row := &repetition.ByMultiplicity[m]
		row.Checkers = m + 1
		for _, name := range declCensusClassNames {
			row.ByClass = append(row.ByClass, declCensusNamedCost{Name: name})
		}
		for kind := range checker.DeclCensusKindCount {
			row.ByKind = append(row.ByKind, declCensusNamedCost{Name: kind.String()})
		}
	}
	coverage := &r.Coverage
	guard := &r.Guard
	materialisation := make([]declCensusFrameTotals, checker.DeclCensusKindCount+1)
	materialisation[0].Name = "all"
	for kind := range checker.DeclCensusKindCount {
		materialisation[kind+1].Name = kind.String()
	}
	weights := make([]int64, j.nodes())
	var ranked []declCensusRankedKey
	for node := range int32(j.nodes()) {
		cost := j.keyCost(node)
		kind := j.kinds[node]
		flags := j.flags[node]
		weights[node] = cost.minNs
		repetition.Total.add(&cost)
		row := &repetition.ByMultiplicity[cost.checkers-1]
		row.Total.add(&cost)
		row.ByClass[declCensusClass(flags)].add(&cost)
		row.ByKind[kind].add(&cost)
		removable := cost.removableNs()
		closed := flags&checker.DeclCensusPoisonFlags == 0
		if closed {
			coverage.ClosedRemovableNs += removable
		}
		if cost.checkers > 1 {
			j.classifyAvailability(node, &repetition.Availability)
			if closed {
				j.classifyAvailability(node, &repetition.ClosedAvailability)
			}
			if removable > 0 {
				ranked = append(ranked, declCensusRankedKey{node: node, removableNs: removable})
			}
		}
		s0, s1, s2 := declCensusSubsets(kind, flags)
		if s0 {
			coverage.S0.Keys++
			coverage.S0.RemovableNs += removable
		}
		if s1 {
			coverage.S1.Keys++
			coverage.S1.RemovableNs += removable
			for c := range k {
				if f := j.frame(node, c); f != nil {
					materialisation[0].add(f)
					materialisation[kind+1].add(f)
				}
			}
			if cost.checkers > 1 {
				guard.RepeatedS1Keys++
				guard.RepeatedS1RemovableNs += removable
				if j.guardSensitive(node) {
					guard.SensitiveKeys++
					guard.SensitiveRemovableNs += removable
				}
			}
		}
		if s2 {
			coverage.S2.Keys++
			coverage.S2.RemovableNs += removable
		}
	}
	coverage.RemovableNs = repetition.Total.RemovableNs
	for _, subset := range []*declCensusSubset{&coverage.S0, &coverage.S1, &coverage.S2} {
		subset.ShareOfRemovable = declCensusShare(subset.RemovableNs, coverage.RemovableNs)
		subset.ShareOfClosedRemovable = declCensusShare(subset.RemovableNs, coverage.ClosedRemovableNs)
	}
	guard.Share = declCensusShare(guard.SensitiveRemovableNs, guard.RepeatedS1RemovableNs)
	r.Materialisation = materialisation
	r.DAG = j.dag(weights)
	r.Heaviest = j.heaviest(ranked, 40)

	summary.RemovableNs = repetition.Total.RemovableNs
	summary.RemovableShare = declCensusShare(summary.RemovableNs, summary.ExclusiveNs)
	summary.ClosedRemovableNs = coverage.ClosedRemovableNs
	summary.ClosedRemovableShare = declCensusShare(summary.ClosedRemovableNs, summary.ExclusiveNs)
	summary.AvailableShare = declCensusShare(repetition.Availability.AvailableNs, repetition.Availability.RepeatedNs)
	summary.UniqueWorkNs = r.DAG.UniqueWorkNs
	summary.SpanNs = r.DAG.SpanNs
	summary.WorkSpanRatio = r.DAG.WorkSpanRatio
	summary.Makespan4Ns = r.DAG.Makespan4Ns
	summary.Makespan8Ns = r.DAG.Makespan8Ns
	summary.S0Share = coverage.S0.ShareOfRemovable
	summary.S1Share = coverage.S1.ShareOfRemovable
	summary.GuardSensitiveShare = guard.Share
}

// classifyAvailability adds the exclusive time of every consumer of a repeated key to the
// availability class of its timing relative to the producer, the checker that finished the key
// first.
func (j *declCensusJoin) classifyAvailability(node int32, a *declCensusAvailability) {
	var producer *checker.DeclCensusFrame
	for c := range j.censuses {
		if f := j.frame(node, c); f != nil && (producer == nil || f.End < producer.End) {
			producer = f
		}
	}
	for c := range j.censuses {
		f := j.frame(node, c)
		if f == nil || f == producer {
			continue
		}
		a.RepeatedNs += f.ExclNs
		switch {
		case producer.End <= f.Start:
			a.AvailableNs += f.ExclNs
		case producer.Start <= f.Start:
			a.InProgressNs += f.ExclNs
		default:
			a.ConsumerFirstNs += f.ExclNs
		}
	}
}

// guardSensitive is the rule of section 2d: the instantiations or the depth of a key exceed
// what a guard tolerates in some checker, or differ by more than a factor of two between
// checkers. Differences below a few instantiations or levels are noise rather than
// cache-state dependence and are ignored.
func (j *declCensusJoin) guardSensitive(node int32) bool {
	const (
		instantiationLimit, depthLimit = 100_000, 50
		instantiationFloor, depthFloor = 16, 4
	)
	var minInst, maxInst, minDepth, maxDepth uint32
	first := true
	for c := range j.censuses {
		f := j.frame(node, c)
		if f == nil {
			continue
		}
		if f.InclInstantiations > instantiationLimit || f.Depth > depthLimit {
			return true
		}
		if first {
			minInst, maxInst, minDepth, maxDepth = f.InclInstantiations, f.InclInstantiations, f.Depth, f.Depth
			first = false
			continue
		}
		minInst, maxInst = min(minInst, f.InclInstantiations), max(maxInst, f.InclInstantiations)
		minDepth, maxDepth = min(minDepth, f.Depth), max(maxDepth, f.Depth)
	}
	return maxInst >= instantiationFloor && maxInst > 2*minInst || maxDepth >= depthFloor && maxDepth > 2*minDepth
}

type declCensusRankedKey struct {
	node        int32
	removableNs int64
}

func (j *declCensusJoin) heaviest(ranked []declCensusRankedKey, n int) []declCensusHeavyKey {
	slices.SortFunc(ranked, func(a, b declCensusRankedKey) int {
		return cmp.Or(cmp.Compare(b.removableNs, a.removableNs), cmp.Compare(a.node, b.node))
	})
	heaviest := make([]declCensusHeavyKey, 0, min(n, len(ranked)))
	for _, r := range ranked[:min(n, len(ranked))] {
		key := declCensusHeavyKey{Kind: j.kinds[r.node].String(), Flags: j.flags[r.node].Names(), RemovableNs: r.removableNs}
		for c := range j.censuses {
			f := j.frame(r.node, c)
			if f == nil {
				continue
			}
			if key.Key == "" {
				key.Key = fmt.Sprintf("%016x%016x", f.Key.Hi, f.Key.Lo)
				key.Description = f.Description()
			}
			key.Checkers++
			key.PerChecker = append(key.PerChecker, declCensusFrameCost{Checker: c, ExclNs: f.ExclNs, Types: f.Types, Symbols: f.Symbols, Signatures: f.Signatures,
				Instantiations: f.Instantiations, InclInstantiations: f.InclInstantiations, Depth: f.Depth, StartNs: f.Start, EndNs: f.End})
		}
		heaviest = append(heaviest, key)
	}
	return heaviest
}

// Dependency DAG.

// declCensusCSR is an adjacency list in compressed sparse row form.
type declCensusCSR struct {
	offsets []int32
	targets []int32
}

func (g *declCensusCSR) of(node int32) []int32 {
	return g.targets[g.offsets[node]:g.offsets[node+1]]
}

// newDeclCensusCSR builds the adjacency of n nodes from edges packed as from<<32 | to, which it
// sorts and deduplicates.
func newDeclCensusCSR(n int, edges []uint64) declCensusCSR {
	slices.Sort(edges)
	edges = slices.Compact(edges)
	g := declCensusCSR{offsets: make([]int32, n+1), targets: make([]int32, len(edges))}
	for i, e := range edges {
		g.offsets[int32(e>>32)+1]++
		g.targets[i] = int32(uint32(e))
	}
	for i := range n {
		g.offsets[i+1] += g.offsets[i]
	}
	return g
}

// transpose returns the reversed adjacency, sources in increasing order.
func (g *declCensusCSR) transpose() declCensusCSR {
	n := len(g.offsets) - 1
	t := declCensusCSR{offsets: make([]int32, n+1), targets: make([]int32, len(g.targets))}
	for _, v := range g.targets {
		t.offsets[v+1]++
	}
	for i := range n {
		t.offsets[i+1] += t.offsets[i]
	}
	next := slices.Clone(t.offsets[:n])
	for u := range int32(n) {
		for _, v := range g.of(u) {
			t.targets[next[v]] = u
			next[v]++
		}
	}
	return t
}

// dag computes the work/span numbers of section 2b over the unique keys: an edge u -> v says
// that computing u needed the result of v (a miss or a hit in some checker). Cycles, possible
// across checkers when two keys needed each other in different orders, are collapsed.
func (j *declCensusJoin) dag(weights []int64) declCensusDAG {
	n := j.nodes()
	var edges []uint64
	for _, cs := range j.censuses {
		for fi := range cs.Frames {
			f := &cs.Frames[fi]
			if f.Parent >= 0 {
				edges = j.appendEdge(edges, cs.Frames[f.Parent].Key, f.Key)
			}
		}
		for _, hit := range cs.HitEdges() {
			edges = j.appendEdge(edges, cs.Frames[hit[0]].Key, cs.Frames[hit[1]].Key)
		}
	}
	graph := newDeclCensusCSR(n, edges)
	component, components := declCensusComponents(n, &graph)

	// Condense: component ids are in completion order, so a dependency's component has a
	// smaller id than the component that depends on it.
	weight := make([]int64, components)
	var condensed []uint64
	for u := range int32(n) {
		weight[component[u]] += weights[u]
		for _, v := range graph.of(u) {
			if component[u] != component[v] {
				condensed = append(condensed, uint64(component[u])<<32|uint64(component[v]))
			}
		}
	}
	deps := newDeclCensusCSR(components, condensed)
	dependents := deps.transpose()

	result := declCensusDAG{Nodes: n, Edges: len(graph.targets), CollapsedCycles: n - components}
	// Longest path ending at each component, dependencies first.
	level := make([]int64, components)
	for c := range int32(components) {
		var longest int64
		for _, d := range deps.of(c) {
			longest = max(longest, level[d])
		}
		level[c] = weight[c] + longest
		result.SpanNs = max(result.SpanNs, level[c])
		result.UniqueWorkNs += weight[c]
	}
	// Longest path from each component to the end, dependents first: the list-scheduling priority.
	priority := make([]int64, components)
	for c := int32(components) - 1; c >= 0; c-- {
		var longest int64
		for _, d := range dependents.of(c) {
			longest = max(longest, priority[d])
		}
		priority[c] = weight[c] + longest
	}
	result.WorkSpanRatio = declCensusShare(result.UniqueWorkNs, result.SpanNs)
	result.Makespan4Ns = declCensusMakespan(4, weight, &deps, &dependents, priority)
	result.Makespan8Ns = declCensusMakespan(8, weight, &deps, &dependents, priority)
	return result
}

func (j *declCensusJoin) appendEdge(edges []uint64, from, to checker.CacheHashKey) []uint64 {
	u, v := j.index[from], j.index[to]
	if u == v {
		return edges
	}
	return append(edges, uint64(u)<<32|uint64(v))
}

// declCensusComponents is Tarjan's algorithm, iterative because dependency chains can be deep.
// It numbers the strongly connected components in completion order.
func declCensusComponents(n int, g *declCensusCSR) ([]int32, int) {
	type visit struct {
		node int32
		next int32 // next edge of node to follow
	}
	index := make([]int32, n) // 1-based visit order, 0 = unvisited
	low := make([]int32, n)
	component := make([]int32, n)
	onStack := make([]bool, n)
	var stack []int32
	var path []visit
	var next, components int32
	for root := range int32(n) {
		if index[root] != 0 {
			continue
		}
		next++
		index[root], low[root] = next, next
		stack = append(stack, root)
		onStack[root] = true
		path = append(path, visit{node: root, next: g.offsets[root]})
		for len(path) > 0 {
			top := &path[len(path)-1]
			if top.next < g.offsets[top.node+1] {
				w := g.targets[top.next]
				top.next++
				if index[w] == 0 {
					next++
					index[w], low[w] = next, next
					stack = append(stack, w)
					onStack[w] = true
					path = append(path, visit{node: w, next: g.offsets[w]})
				} else if onStack[w] {
					low[top.node] = min(low[top.node], index[w])
				}
				continue
			}
			v := top.node
			path = path[:len(path)-1]
			if len(path) > 0 {
				u := path[len(path)-1].node
				low[u] = min(low[u], low[v])
			}
			if low[v] == index[v] {
				for {
					w := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					onStack[w] = false
					component[w] = components
					if w == v {
						break
					}
				}
				components++
			}
		}
	}
	return component, int(components)
}

// declCensusMakespan simulates list scheduling of the condensed DAG on the given number of
// workers: whenever a worker is free it takes the ready component with the longest path to the
// end.
func declCensusMakespan(workers int, weight []int64, deps, dependents *declCensusCSR, priority []int64) int64 {
	type running struct {
		finish int64
		node   int32
	}
	ready := declCensusHeap[int32]{less: func(a, b int32) bool { return priority[a] > priority[b] }}
	busy := declCensusHeap[running]{less: func(a, b running) bool { return a.finish < b.finish }}
	pending := make([]int32, len(weight))
	for c := range int32(len(weight)) {
		pending[c] = int32(len(deps.of(c)))
		if pending[c] == 0 {
			ready.push(c)
		}
	}
	var now int64
	for ready.len() > 0 || busy.len() > 0 {
		for busy.len() < workers && ready.len() > 0 {
			c := ready.pop()
			busy.push(running{finish: now + weight[c], node: c})
		}
		done := busy.pop()
		now = done.finish
		for _, d := range dependents.of(done.node) {
			pending[d]--
			if pending[d] == 0 {
				ready.push(d)
			}
		}
	}
	return now
}

// declCensusHeap is a binary heap ordered by less.
type declCensusHeap[T any] struct {
	items []T
	less  func(a, b T) bool
}

func (h *declCensusHeap[T]) len() int {
	return len(h.items)
}

func (h *declCensusHeap[T]) push(item T) {
	h.items = append(h.items, item)
	for i := len(h.items) - 1; i > 0; {
		parent := (i - 1) / 2
		if !h.less(h.items[i], h.items[parent]) {
			break
		}
		h.items[i], h.items[parent] = h.items[parent], h.items[i]
		i = parent
	}
}

func (h *declCensusHeap[T]) pop() T {
	top := h.items[0]
	last := len(h.items) - 1
	h.items[0] = h.items[last]
	h.items = h.items[:last]
	for i := 0; ; {
		smallest := i
		for _, child := range [2]int{2*i + 1, 2*i + 2} {
			if child < last && h.less(h.items[child], h.items[smallest]) {
				smallest = child
			}
		}
		if smallest == i {
			break
		}
		h.items[i], h.items[smallest] = h.items[smallest], h.items[i]
		i = smallest
	}
	return top
}
