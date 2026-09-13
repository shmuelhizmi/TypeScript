package compiler

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/json"
)

// checkerTimeline records, for one forEachCheckerGroupDo call, when every checker
// starts and finishes its file group and how long each file takes, and prints the
// result as one JSON line on stderr. Research instrumentation, enabled by setting
// the environment variable TSGO_CHECKER_TIMELINE; a nil timeline records nothing.
type checkerTimeline struct {
	program  string
	started  time.Time
	weights  map[*ast.SourceFile]int
	summary  checkerTimelineWeights
	checkers []checkerTimelineChecker
}

// checkerTimelineWeights are the totals the association policy was derived from.
type checkerTimelineWeights struct {
	TotalBase                  int  `json:"totalBaseWeight"`
	DeclarationBase            int  `json:"declarationBaseWeight"`
	DeclarationFiles           int  `json:"declarationFiles"`
	SourceFileWeightMultiplier int  `json:"sourceFileWeightMultiplier"`
	BalancePenaltyMultiplier   int  `json:"balancePenaltyMultiplier"`
	PrioritizeSourceFiles      bool `json:"prioritizeSourceFiles"`
}

// checkerTimelineChecker is written by the goroutine of its checker only.
type checkerTimelineChecker struct {
	Checker int                   `json:"checker"`
	StartMs float64               `json:"startMs"`
	EndMs   float64               `json:"endMs"`
	BusyMs  float64               `json:"busyMs"`
	Weight  int                   `json:"weight"`
	Files   int                   `json:"files"`
	files   []checkerTimelineFile `json:"-"`
}

type checkerTimelineFile struct {
	Path        string  `json:"path"`
	Checker     int     `json:"checker"`
	Ms          float64 `json:"ms"`
	Weight      int     `json:"weight"`
	Declaration bool    `json:"declaration"`
}

type checkerTimelineReport struct {
	Program      string                   `json:"program"`
	Files        int                      `json:"files"`
	PhaseMs      float64                  `json:"phaseMs"`
	BusyMs       float64                  `json:"busyMs"`
	IdealMs      float64                  `json:"idealMs"`
	Weights      checkerTimelineWeights   `json:"weights"`
	Checkers     []checkerTimelineChecker `json:"checkers"`
	SlowestFiles []checkerTimelineFile    `json:"slowestFiles"`
}

func checkerTimelineEnabled() bool {
	return os.Getenv("TSGO_CHECKER_TIMELINE") != ""
}

func newCheckerTimeline(program string, checkerCount int, weights map[*ast.SourceFile]int, summary checkerTimelineWeights) *checkerTimeline {
	if !checkerTimelineEnabled() {
		return nil
	}
	t := &checkerTimeline{program: program, started: time.Now(), weights: weights, summary: summary, checkers: make([]checkerTimelineChecker, checkerCount)}
	for i := range t.checkers {
		t.checkers[i].Checker = i
	}
	return t
}

func (t *checkerTimeline) elapsedMs() float64 {
	return float64(time.Since(t.started).Microseconds()) / 1000
}

func (t *checkerTimeline) start(idx int) {
	if t != nil {
		t.checkers[idx].StartMs = t.elapsedMs()
	}
}

func (t *checkerTimeline) begin() float64 {
	if t == nil {
		return 0
	}
	return t.elapsedMs()
}

func (t *checkerTimeline) end(idx int, file *ast.SourceFile, began float64) {
	if t == nil {
		return
	}
	ms := t.elapsedMs() - began
	entry := &t.checkers[idx]
	entry.BusyMs += ms
	entry.Weight += t.weights[file]
	entry.Files++
	entry.files = append(entry.files, checkerTimelineFile{Path: file.FileName(), Checker: idx, Ms: ms, Weight: t.weights[file], Declaration: file.IsDeclarationFile})
}

func (t *checkerTimeline) finish(idx int) {
	if t != nil {
		t.checkers[idx].EndMs = t.elapsedMs()
	}
}

// report prints the timeline once every checker goroutine has finished. idealMs is the
// lower bound of the phase for the same per-file costs and a perfect assignment:
// the larger of the average load and the slowest file.
func (t *checkerTimeline) report(w io.Writer) {
	if t == nil {
		return
	}
	report := checkerTimelineReport{Program: t.program, Weights: t.summary, Checkers: t.checkers}
	var files []checkerTimelineFile
	for _, c := range t.checkers {
		report.PhaseMs = max(report.PhaseMs, c.EndMs)
		report.BusyMs += c.BusyMs
		files = append(files, c.files...)
	}
	report.Files = len(files)
	slices.SortFunc(files, func(a, b checkerTimelineFile) int { return cmp.Compare(b.Ms, a.Ms) })
	report.SlowestFiles = files[:min(len(files), 40)]
	report.IdealMs = report.BusyMs / float64(len(t.checkers))
	if len(files) > 0 {
		report.IdealMs = max(report.IdealMs, files[0].Ms)
	}
	line, err := json.Marshal(report)
	if err != nil {
		fmt.Fprintf(w, "checker-timeline error: %v\n", err)
		return
	}
	fmt.Fprintf(w, "checker-timeline %s\n", line)
}
