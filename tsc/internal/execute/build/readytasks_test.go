package build

import (
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestReadyBuildTasksBypassDependencyAndReportWaits(t *testing.T) {
	t.Parallel()
	// The first task is blocked. Its dependent must not consume the second
	// worker, and completed later tasks must not hold that worker for reporting.
	dependencies := []taskDependencies{{}, {finished: []int{0}}, {}, {}, {}, {finished: []int{1, 2}}}
	releaseFirst := make(chan struct{})
	lastIndependent := make(chan struct{})
	returned := make(chan struct{})
	var executed [6]atomic.Bool
	var active atomic.Int32
	var exceeded atomic.Bool
	var reports []int
	go func() {
		executeReadyBuildTasks(dependencies, 2, func(i int, _ func(func())) {
			if active.Add(1) > 2 {
				exceeded.Store(true)
			}
			defer active.Add(-1)
			for _, upstream := range dependencies[i].finished {
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
		// Tasks 2, 3 and 4 ran on the free worker while task 0 stayed blocked.
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

func TestReadyBuildTasksStartAfterDependencyStarted(t *testing.T) {
	t.Parallel()
	// Task 1 may start once task 0 has started; task 2 only once task 0 has finished.
	dependencies := []taskDependencies{{}, {started: []int{0}}, {finished: []int{0}}}
	releaseFirst := make(chan struct{})
	secondStarted := make(chan struct{})
	returned := make(chan struct{})
	var firstFinished atomic.Bool
	var reports []int
	go func() {
		executeReadyBuildTasks(dependencies, 2, func(i int, _ func(func())) {
			switch i {
			case 0:
				<-releaseFirst
				firstFinished.Store(true)
			case 1:
				close(secondStarted)
			case 2:
				if !firstFinished.Load() {
					t.Error("task 2 ran before task 0 finished")
				}
			}
		}, func(i int) {
			reports = append(reports, i)
		})
		close(returned)
	}()
	select {
	case <-secondStarted:
	case <-time.After(30 * time.Second):
		close(releaseFirst)
		<-returned
		t.Fatal("task 1 did not start while task 0 was running")
	}
	close(releaseFirst)
	<-returned
	if !slices.Equal(reports, []int{0, 1, 2}) {
		t.Fatalf("report order = %v", reports)
	}
}

func TestReadyBuildTasksStartAfterOneDependencyFinished(t *testing.T) {
	t.Parallel()
	// Task 2 may start once tasks 0 and 1 have started and one of them has finished.
	dependencies := []taskDependencies{{}, {}, {started: []int{0, 1}, oneFinished: []int{0, 1}}}
	releaseFirst := make(chan struct{})
	thirdStarted := make(chan struct{})
	returned := make(chan struct{})
	var secondFinished atomic.Bool
	var reports []int
	go func() {
		executeReadyBuildTasks(dependencies, 3, func(i int, _ func(func())) {
			switch i {
			case 0:
				<-releaseFirst
			case 1:
				secondFinished.Store(true)
			case 2:
				if !secondFinished.Load() {
					t.Error("task 2 ran before task 1 finished")
				}
				close(thirdStarted)
			}
		}, func(i int) {
			reports = append(reports, i)
		})
		close(returned)
	}()
	select {
	case <-thirdStarted:
	case <-time.After(30 * time.Second):
		close(releaseFirst)
		<-returned
		t.Fatal("task 2 did not start while task 0 was running")
	}
	close(releaseFirst)
	<-returned
	if !slices.Equal(reports, []int{0, 1, 2}) {
		t.Fatalf("report order = %v", reports)
	}
}

func TestReadyBuildTasksSuspendReleasesWorker(t *testing.T) {
	t.Parallel()
	// Tasks 0 and 1 hold both workers until task 2 has run, which it can only do
	// on the worker that task 0 releases by suspending.
	dependencies := []taskDependencies{{}, {started: []int{0}}, {started: []int{1}}}
	thirdRan := make(chan struct{})
	var closeThirdRan sync.Once
	returned := make(chan struct{})
	var reports []int
	go func() {
		executeReadyBuildTasks(dependencies, 2, func(i int, suspend func(func())) {
			switch i {
			case 0:
				suspend(func() { <-thirdRan })
			case 1:
				<-thirdRan
			case 2:
				closeThirdRan.Do(func() { close(thirdRan) })
			}
		}, func(i int) {
			reports = append(reports, i)
		})
		close(returned)
	}()
	select {
	case <-returned:
	case <-time.After(30 * time.Second):
		closeThirdRan.Do(func() { close(thirdRan) })
		<-returned
		t.Fatal("suspending a task did not release its worker")
	}
	if !slices.Equal(reports, []int{0, 1, 2}) {
		t.Fatalf("report order = %v", reports)
	}
}

func TestReadyBuildTasksSerialOrder(t *testing.T) {
	t.Parallel()
	var operations []int
	executeReadyBuildTasks([]taskDependencies{{}, {finished: []int{0}}, {}, {finished: []int{1, 2}}}, 1,
		func(i int, _ func(func())) { operations = append(operations, i*2) },
		func(i int) { operations = append(operations, i*2+1) })
	if !slices.Equal(operations, []int{0, 1, 2, 3, 4, 5, 6, 7}) {
		t.Fatalf("serial order = %v", operations)
	}
}

func TestReadyBuildTasksDiamondAndDuplicateReferences(t *testing.T) {
	t.Parallel()
	dependencies := []taskDependencies{{}, {finished: []int{0}}, {finished: []int{0}}, {finished: []int{1, 2, 2}}, {finished: []int{3}}}
	var executions [5]atomic.Int32
	var reports []int
	executeReadyBuildTasks(dependencies, 8, func(i int, _ func(func())) {
		for _, dependency := range dependencies[i].finished {
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
	executeReadyBuildTasks(nil, 4,
		func(int, func(func())) { t.Fatal("callback invoked for empty graph") },
		func(int) { t.Fatal("callback invoked for empty graph") })
}
