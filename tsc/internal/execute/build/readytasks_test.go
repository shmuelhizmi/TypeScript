package build

import (
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

func TestReadyBuildTasksBypassDependencyAndReportWaits(t *testing.T) {
	t.Parallel()
	// The first task is blocked. Its dependent must not consume the second
	// worker, and completed later tasks must not hold that worker for reporting.
	dependencies := [][]int{nil, {0}, nil, nil, nil, {1, 2}}
	releaseFirst := make(chan struct{})
	lastIndependent := make(chan struct{})
	returned := make(chan struct{})
	var executed [6]atomic.Bool
	var active atomic.Int32
	var exceeded atomic.Bool
	var reports []int
	go func() {
		executeReadyBuildTasks(dependencies, 2, func(i int) {
			if active.Add(1) > 2 {
				exceeded.Store(true)
			}
			defer active.Add(-1)
			for _, upstream := range dependencies[i] {
				if !executed[upstream].Load() {
					t.Errorf("task %d ran before dependency %d", i, upstream)
				}
			}
			if i == 0 {
				<-releaseFirst
			}
			executed[i].Store(true)
			if i == 4 {
				close(lastIndependent)
			}
		}, func(i int) {
			reports = append(reports, i)
		})
		close(returned)
	}()
	select {
	case <-lastIndependent:
	case <-time.After(30 * time.Second):
		close(releaseFirst)
		<-returned
		t.Fatal("ready tasks stalled behind a dependency or reporting wait")
	}
	close(releaseFirst)
	<-returned
	if exceeded.Load() {
		t.Fatal("worker limit exceeded")
	}
	if !slices.Equal(reports, []int{0, 1, 2, 3, 4, 5}) {
		t.Fatalf("report order = %v", reports)
	}
}

func TestReadyBuildTasksSerialOrder(t *testing.T) {
	t.Parallel()
	var operations []int
	executeReadyBuildTasks([][]int{nil, {0}, nil, {1, 2}}, 1,
		func(i int) { operations = append(operations, i*2) },
		func(i int) { operations = append(operations, i*2+1) })
	if !slices.Equal(operations, []int{0, 1, 2, 3, 4, 5, 6, 7}) {
		t.Fatalf("serial order = %v", operations)
	}
}

func TestReadyBuildTasksDiamondAndDuplicateReferences(t *testing.T) {
	t.Parallel()
	dependencies := [][]int{nil, {0}, {0}, {1, 2, 2}, {3}}
	var executions [5]atomic.Int32
	var reports []int
	executeReadyBuildTasks(dependencies, 8, func(i int) {
		for _, dependency := range dependencies[i] {
			if executions[dependency].Load() != 1 {
				t.Errorf("task %d ran before dependency %d", i, dependency)
			}
		}
		if executions[i].Add(1) != 1 {
			t.Errorf("task %d ran more than once", i)
		}
	}, func(i int) { reports = append(reports, i) })
	if !slices.Equal(reports, []int{0, 1, 2, 3, 4}) {
		t.Fatalf("report order = %v", reports)
	}
}

func TestReadyBuildTasksEmpty(t *testing.T) {
	t.Parallel()
	fail := func(int) { t.Fatal("callback invoked for empty graph") }
	executeReadyBuildTasks(nil, 4, fail, fail)
}
