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
// cycle and reported no diagnostic.
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
	flowMemoCountLen
)

var flowMemoCountNames = [flowMemoCountLen]string{
	"invocations", "flow_ns", "visits", "max_depth", "depth_trips", "junctions",
	"ineligible_reference", "ineligible_loop", "ineligible_reduce", "ineligible_inline", "ineligible_disabled", "ineligible_key",
	"hits", "hits_with_incomplete_shared", "outermost_hits", "saved_visits", "saved_ns",
	"mismatch_type", "mismatch_incomplete", "stores",
	"blocked_incomplete", "blocked_disabled", "blocked_reference", "blocked_cycle", "blocked_diagnostic",
}

// flowMemoStats is the per-checker state of the instrument; it is nil when the instrument is off.
type flowMemoStats struct {
	memo            map[FlowLoopKey]*Type
	shadowDepth     int // > 0 while a hit is being recomputed in shadow mode
	invocationDepth int // nesting of getFlowTypeOfReferenceEx
	cycles          int // type resolution cycles found
	diagnostics     int // diagnostics and suggestions reported
	counts          [flowMemoCountLen]int64
}

var (
	flowMemoStatsMu   sync.Mutex
	flowMemoAllStats  []*flowMemoStats
	flowMemoReported  atomic.Int32
	flowMemoReportMax = int32(20)
)

func newFlowMemoStats() *flowMemoStats {
	if flowMemoMode == "" {
		return nil
	}
	s := &flowMemoStats{memo: make(map[FlowLoopKey]*Type)}
	flowMemoStatsMu.Lock()
	flowMemoAllStats = append(flowMemoAllStats, s)
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
	var total [flowMemoCountLen]int64
	for _, s := range flowMemoAllStats {
		for i, n := range s.counts {
			if flowMemoCount(i) == flowMemoMaxDepth {
				total[i] = max(total[i], n)
			} else {
				total[i] += n
			}
		}
	}
	fmt.Fprintf(w, "flow-memo\tmode\t%s\n", flowMemoMode)
	fmt.Fprintf(w, "flow-memo\tcheckers\t%d\n", len(flowMemoAllStats))
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
	cached, hit := s.memo[key]
	if hit {
		s.counts[flowMemoHits]++
		if f.incompleteShared != 0 {
			// A recomputation could reach an incomplete shared flow type of this invocation.
			s.counts[flowMemoHitsWithIncompleteShared]++
		} else if flowMemoMode == "on" {
			return FlowType{t: cached}
		}
		outermost := s.shadowDepth == 0
		visitsBefore := s.counts[flowMemoVisits]
		var start time.Time
		if outermost && flowMemoTiming {
			start = time.Now()
		}
		s.shadowDepth++
		result := c.computeTypeAtFlowJunction(f, flow, antecedents)
		s.shadowDepth--
		if outermost {
			s.counts[flowMemoOutermostHits]++
			s.counts[flowMemoSavedVisits] += s.counts[flowMemoVisits] - visitsBefore
			if flowMemoTiming {
				s.counts[flowMemoSavedNanos] += int64(time.Since(start))
			}
		}
		if result.t != cached {
			s.counts[flowMemoMismatchType]++
			c.reportFlowMemoMismatch(f, flow, cached, result)
		} else if result.incomplete {
			s.counts[flowMemoMismatchIncomplete]++
			c.reportFlowMemoMismatch(f, flow, cached, result)
		}
		return result
	}
	referenceDependent := f.referenceDependent
	cycles := s.cycles
	diagnostics := s.diagnostics
	result := c.computeTypeAtFlowJunction(f, flow, antecedents)
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
	default:
		s.memo[key] = result.t
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
