package checker

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/scanner"
)

// Lab instrument (TSGO_FLOW_MEMO). It memoizes the flow type of an identifier or 'this'
// reference at condition, switch clause and branch label nodes across flow analysis
// invocations, keyed like the flow loop cache. With "shadow" every hit is recomputed
// and compared with the memoized type, so results are unchanged; with "on" hits are
// returned. A result is stored only when its computation was complete, ran outside loop
// analysis, reduce labels and aliased-condition inlining, did not trip the depth limit,
// took no decision that depends on the reference node itself, hit no type resolution
// cycle and reported no diagnostic, and only when its type would be the same object if
// recomputed (see isFlowMemoStableType).
var flowMemoMode = os.Getenv("TSGO_FLOW_MEMO")

// flowMemoTiming times flow analysis and recomputed hits; only shadow mode pays for it.
var flowMemoTiming = flowMemoMode == "shadow"

type flowMemoCount int

const (
	flowMemoInvocations flowMemoCount = iota
	flowMemoFlowNanos
	flowMemoVisits
	flowMemoMaxDepth
	flowMemoDepthTrips
	flowMemoJunctions
	flowMemoIneligibleReference
	flowMemoIneligibleLoop
	flowMemoIneligibleReduce
	flowMemoIneligibleInline
	flowMemoIneligibleDisabled
	flowMemoIneligibleKey
	flowMemoHits
	flowMemoHitsWithIncompleteShared
	flowMemoOutermostHits
	flowMemoSavedVisits
	flowMemoSavedNanos
	flowMemoMismatchType
	flowMemoMismatchIncomplete
	flowMemoStores
	flowMemoBlockedIncomplete
	flowMemoBlockedDisabled
	flowMemoBlockedReference
	flowMemoBlockedCycle
	flowMemoBlockedDiagnostic
	flowMemoBlockedFresh
	flowMemoDepthRefusedHits
	flowMemoMaxBound
	flowMemoCountLen
)

var flowMemoCountNames = [flowMemoCountLen]string{
	"invocations", "flow_ns", "visits", "max_depth", "depth_trips", "junctions",
	"ineligible_reference", "ineligible_loop", "ineligible_reduce", "ineligible_inline", "ineligible_disabled", "ineligible_key",
	"hits", "hits_with_incomplete_shared", "outermost_hits", "saved_visits", "saved_ns",
	"mismatch_type", "mismatch_incomplete", "stores",
	"blocked_incomplete", "blocked_disabled", "blocked_reference", "blocked_cycle", "blocked_diagnostic", "blocked_fresh",
	"depth_refused_hits", "max_bound",
}

// flowMemoStats is the per-checker state of the instrument; it is nil when the instrument is off.
// flowMemoEntry is a memoized flow type and the deepest additional recursion its computation
// needed. Taking the entry skips that recursion, so it is taken only when the recursion could
// not have reached the depth limit: the limit then trips exactly where it would without the
// memo (the memo only ever grows, so a later computation cannot recurse deeper than the bound).
type flowMemoEntry struct {
	t          *Type
	depthBound int32
}

type flowMemoStats struct {
	memo            map[FlowLoopKey]flowMemoEntry
	shadowDepth     int // > 0 while a hit is being recomputed in shadow mode
	invocationDepth int // nesting of getFlowTypeOfReferenceEx
	cycles          int // type resolution cycles found
	diagnostics     int // diagnostics and suggestions reported
	counts          *flowMemoCounts
}

type flowMemoCounts [flowMemoCountLen]int64

var (
	flowMemoStatsMu   sync.Mutex
	flowMemoAllCounts []*flowMemoCounts // only the counters: the memo dies with its checker
	flowMemoReported  atomic.Int32
	flowMemoReportMax = int32(20)
)

func newFlowMemoStats() *flowMemoStats {
	if flowMemoMode == "" {
		return nil
	}
	s := &flowMemoStats{memo: make(map[FlowLoopKey]flowMemoEntry), counts: new(flowMemoCounts)}
	flowMemoStatsMu.Lock()
	flowMemoAllCounts = append(flowMemoAllCounts, s.counts)
	flowMemoStatsMu.Unlock()
	return s
}

// WriteFlowMemoCensus prints the instrument's counters summed over all checkers.
func WriteFlowMemoCensus(w io.Writer) {
	if flowMemoMode == "" {
		return
	}
	flowMemoStatsMu.Lock()
	defer flowMemoStatsMu.Unlock()
	var total flowMemoCounts
	for _, counts := range flowMemoAllCounts {
		for i, n := range counts {
			if flowMemoCount(i) == flowMemoMaxDepth || flowMemoCount(i) == flowMemoMaxBound {
				total[i] = max(total[i], n)
			} else {
				total[i] += n
			}
		}
	}
	fmt.Fprintf(w, "flow-memo\tmode\t%s\n", flowMemoMode)
	fmt.Fprintf(w, "flow-memo\tcheckers\t%d\n", len(flowMemoAllCounts))
	for i, n := range total {
		fmt.Fprintf(w, "flow-memo\t%s\t%d\n", flowMemoCountNames[i], n)
	}
}

func isFlowMemoReference(reference *ast.Node) bool {
	return reference.Kind == ast.KindThisKeyword || ast.IsIdentifier(reference) && !ast.IsThisInTypeQuery(reference)
}

// isReferenceDependentNarrowableType reports whether getNarrowableTypeForReference may
// substitute constraints in t depending on the position of the reference.
func (c *Checker) isReferenceDependentNarrowableType(t *Type) bool {
	if c.isNoInferType(t) {
		t = t.AsSubstitutionType().baseType
	}
	return someType(t, c.isGenericTypeWithUnionConstraint)
}

// A type that originates in an object, array or JSX literal is created afresh by every
// evaluation of that expression, so two evaluations yield structurally equal but distinct
// types. A memo must not hand one evaluation's copy to another reference: the copies print
// the same but are not identical, and a union of both would print its member twice.
const flowMemoUnstableObjectFlags = ObjectFlagsObjectLiteral | ObjectFlagsArrayLiteral | ObjectFlagsJsxAttributes

// isFlowMemoStableType reports whether recomputing a flow type is known to yield t itself
// rather than a structurally equal copy: t carries no literal-expression origin, and either
// existed before the flow analysis invocation began (its id is at most typeCount; types made
// during the invocation can reach the result through the shared flow cache) or is a literal
// type or a union or intersection of stable types, all of which are interned.
func isFlowMemoStableType(t *Type, typeCount uint32) bool {
	if t.objectFlags&flowMemoUnstableObjectFlags != 0 {
		return false
	}
	if t.flags&TypeFlagsUnionOrIntersection != 0 {
		for _, constituent := range t.Types() {
			if !isFlowMemoStableType(constituent, typeCount) {
				return false
			}
		}
		if t.flags&TypeFlagsUnion != 0 {
			if origin := t.AsUnionType().origin; origin != nil && !isFlowMemoStableType(origin, typeCount) {
				return false
			}
		}
	} else if uint32(t.id) > typeCount && t.flags&TypeFlagsLiteral == 0 {
		return false
	}
	if alias := t.alias; alias != nil {
		for _, argument := range alias.typeArguments {
			if !isFlowMemoStableType(argument, typeCount) {
				return false
			}
		}
	}
	return true
}

// computeTypeAtFlowJunctionBounded computes the junction and returns the deepest additional
// recursion it needed.
func (c *Checker) computeTypeAtFlowJunctionBounded(f *FlowState, flow *ast.FlowNode, antecedents *ast.FlowList) (FlowType, int32) {
	outerMax := f.maxDepth
	f.maxDepth = f.depth
	result := c.computeTypeAtFlowJunction(f, flow, antecedents)
	bound := int32(f.maxDepth - f.depth)
	if outerMax > f.maxDepth {
		f.maxDepth = outerMax
	}
	return result, bound
}

func (c *Checker) computeTypeAtFlowJunction(f *FlowState, flow *ast.FlowNode, antecedents *ast.FlowList) FlowType {
	switch {
	case flow.Flags&ast.FlowFlagsCondition != 0:
		return c.getTypeAtFlowCondition(f, flow)
	case flow.Flags&ast.FlowFlagsSwitchClause != 0:
		return c.getTypeAtSwitchClause(f, flow)
	default:
		return c.getTypeAtFlowBranchLabel(f, flow, antecedents)
	}
}

// getTypeAtFlowJunction returns the flow type at a condition, switch clause or branch label node.
func (c *Checker) getTypeAtFlowJunction(f *FlowState, flow *ast.FlowNode, antecedents *ast.FlowList) FlowType {
	s := c.flowMemoStats
	if s == nil {
		return c.computeTypeAtFlowJunction(f, flow, antecedents)
	}
	s.counts[flowMemoJunctions]++
	switch {
	case !isFlowMemoReference(f.reference):
		s.counts[flowMemoIneligibleReference]++
		return c.computeTypeAtFlowJunction(f, flow, antecedents)
	case len(c.flowLoopStack) != 0:
		s.counts[flowMemoIneligibleLoop]++
		return c.computeTypeAtFlowJunction(f, flow, antecedents)
	case len(f.reduceLabels) != 0:
		s.counts[flowMemoIneligibleReduce]++
		return c.computeTypeAtFlowJunction(f, flow, antecedents)
	case c.inlineLevel != 0:
		s.counts[flowMemoIneligibleInline]++
		return c.computeTypeAtFlowJunction(f, flow, antecedents)
	case c.flowAnalysisDisabled:
		s.counts[flowMemoIneligibleDisabled]++
		return c.computeTypeAtFlowJunction(f, flow, antecedents)
	}
	if f.refKey.IsZero() {
		f.refKey = c.getFlowReferenceKey(f)
	}
	if f.refKey == nonDottedNameCacheKey {
		s.counts[flowMemoIneligibleKey]++
		return c.computeTypeAtFlowJunction(f, flow, antecedents)
	}
	key := FlowLoopKey{flowNode: flow, refKey: f.refKey}
	entry, hit := s.memo[key]
	if hit {
		s.counts[flowMemoHits]++
		switch {
		case f.incompleteShared != 0:
			// A recomputation could reach an incomplete shared flow type of this invocation.
			s.counts[flowMemoHitsWithIncompleteShared]++
		case f.depth+int(entry.depthBound) >= flowDepthLimit:
			// Taking the entry here would skip recursion that reaches the depth limit.
			s.counts[flowMemoDepthRefusedHits]++
		case flowMemoMode == "on":
			return FlowType{t: entry.t}
		}
		outermost := s.shadowDepth == 0
		visitsBefore := s.counts[flowMemoVisits]
		var start time.Time
		if outermost && flowMemoTiming {
			start = time.Now()
		}
		s.shadowDepth++
		result, _ := c.computeTypeAtFlowJunctionBounded(f, flow, antecedents)
		s.shadowDepth--
		if outermost {
			s.counts[flowMemoOutermostHits]++
			s.counts[flowMemoSavedVisits] += s.counts[flowMemoVisits] - visitsBefore
			if flowMemoTiming {
				s.counts[flowMemoSavedNanos] += int64(time.Since(start))
			}
		}
		if result.t != entry.t {
			s.counts[flowMemoMismatchType]++
			c.reportFlowMemoMismatch(f, flow, entry.t, result)
		} else if result.incomplete {
			s.counts[flowMemoMismatchIncomplete]++
			c.reportFlowMemoMismatch(f, flow, entry.t, result)
		}
		return result
	}
	referenceDependent := f.referenceDependent
	cycles := s.cycles
	diagnostics := s.diagnostics
	result, bound := c.computeTypeAtFlowJunctionBounded(f, flow, antecedents)
	if int64(bound) > s.counts[flowMemoMaxBound] {
		s.counts[flowMemoMaxBound] = int64(bound)
	}
	switch {
	case result.incomplete:
		s.counts[flowMemoBlockedIncomplete]++
	case c.flowAnalysisDisabled:
		s.counts[flowMemoBlockedDisabled]++
	case f.referenceDependent != referenceDependent:
		s.counts[flowMemoBlockedReference]++
	case s.cycles != cycles:
		s.counts[flowMemoBlockedCycle]++
	case s.diagnostics != diagnostics:
		s.counts[flowMemoBlockedDiagnostic]++
	case !isFlowMemoStableType(result.t, f.typeCount):
		s.counts[flowMemoBlockedFresh]++
	default:
		s.memo[key] = flowMemoEntry{t: result.t, depthBound: bound}
		s.counts[flowMemoStores]++
	}
	return result
}

func (c *Checker) reportFlowMemoMismatch(f *FlowState, flow *ast.FlowNode, cached *Type, result FlowType) {
	if flowMemoReported.Add(1) > flowMemoReportMax {
		return
	}
	file := ast.GetSourceFileOfNode(f.reference)
	line, character := scanner.GetECMALineAndUTF16CharacterOfPosition(file, scanner.GetTokenPosOfNode(f.reference, file, false))
	fmt.Fprintf(os.Stderr, "flow-memo\tmismatch\t%s:%d:%d\t%s\tflags=%#x\tincomplete=%t\tmemo=%s\tcomputed=%s\n",
		file.FileName(), line+1, int(character)+1, scanner.GetTextOfNode(f.reference), uint32(flow.Flags), result.incomplete,
		c.TypeToString(cached), c.TypeToString(result.t))
}
