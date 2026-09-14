package build

import (
	"cmp"
	"slices"
)

// taskDependencies lists, for one task of a build, the tasks it must not start before: those that
// must all have finished, those that must all have started, and those of which at least one must
// have finished. Dependencies have lower indices than the tasks that depend on them.
type taskDependencies struct {
	finished    []int
	started     []int
	oneFinished []int
}

// workerRequest is a task waiting for a worker: one that is ready to start, or one that is resuming
// after a wait.
type workerRequest struct {
	index  int
	resume bool
}

// workerQueue holds the tasks waiting for a worker in the order they receive one: tasks resuming
// after a wait before tasks starting, since a resuming task has work to continue while a starting
// task can only get as far as its dependencies allow, and among those the lowest index first, as a
// task only ever waits for tasks with lower indices.
type workerQueue struct {
	requests []workerRequest
}

func (q *workerQueue) push(request workerRequest) {
	at, _ := slices.BinarySearchFunc(q.requests, request, compareWorkerRequests)
	q.requests = slices.Insert(q.requests, at, request)
}

func (q *workerQueue) pop() (workerRequest, bool) {
	if len(q.requests) == 0 {
		return workerRequest{}, false
	}
	request := q.requests[0]
	q.requests = q.requests[1:]
	return request, true
}

func compareWorkerRequests(a, b workerRequest) int {
	if a.resume != b.resume {
		if a.resume {
			return -1
		}
		return 1
	}
	return cmp.Compare(a.index, b.index)
}

// executeReadyBuildTasks calls execute for every task, starting a task as soon as its dependencies
// are satisfied and running at most workerCount tasks at a time, and calls report for the tasks in
// index order as they complete. Waiting for dependencies and reporting happen on the calling
// goroutine, so a task whose report is pending never occupies a worker. A running task that has to
// wait for another task passes the wait to the suspend function it received: the wait runs with the
// task's worker released, and the task takes a worker again afterwards. Tasks waiting for a worker
// receive one in the order of workerQueue.
func executeReadyBuildTasks(dependencies []taskDependencies, workerCount int, execute func(index int, suspend func(wait func())), report func(index int)) {
	count := len(dependencies)
	if count == 0 {
		return
	}
	if workerCount <= 1 {
		for i := range count {
			execute(i, func(wait func()) { wait() })
			report(i)
		}
		return
	}
	pending := make([]int, count)
	startConsumers := make([][]int, count)
	finishConsumers := make([][]int, count)
	oneFinishedConsumers := make([][]int, count)
	oneFinished := make([]bool, count)
	for i, task := range dependencies {
		pending[i] = len(task.finished) + len(task.started)
		for _, dependency := range task.finished {
			finishConsumers[dependency] = append(finishConsumers[dependency], i)
		}
		for _, dependency := range task.started {
			startConsumers[dependency] = append(startConsumers[dependency], i)
		}
		if len(task.oneFinished) != 0 {
			pending[i]++
			for _, dependency := range task.oneFinished {
				oneFinishedConsumers[dependency] = append(oneFinishedConsumers[dependency], i)
			}
		}
	}
	var waiting workerQueue
	release := func(consumers []int) {
		for _, consumer := range consumers {
			pending[consumer]--
			if pending[consumer] == 0 {
				waiting.push(workerRequest{index: consumer})
			}
		}
	}
	for i := range count {
		if pending[i] == 0 {
			waiting.push(workerRequest{index: i})
		}
	}
	// A task hands its worker back on released while it waits and on completed when it is done. A
	// task that has waited asks for a worker again on resumed and receives it on its grant channel.
	released := make(chan struct{})
	resumed := make(chan int)
	completed := make(chan int)
	grants := make([]chan struct{}, count)
	for i := range grants {
		grants[i] = make(chan struct{}, 1)
	}
	free := workerCount
	grant := func() {
		for free > 0 {
			request, ok := waiting.pop()
			if !ok {
				return
			}
			free--
			if request.resume {
				grants[request.index] <- struct{}{}
				continue
			}
			i := request.index
			go func() {
				execute(i, func(wait func()) {
					released <- struct{}{}
					wait()
					resumed <- i
					<-grants[i]
				})
				completed <- i
			}()
			release(startConsumers[i])
		}
	}
	finished := make([]bool, count)
	nextReport := 0
	grant()
	for remaining := count; remaining > 0; {
		select {
		case <-released:
			free++
		case i := <-resumed:
			waiting.push(workerRequest{index: i, resume: true})
		case i := <-completed:
			free++
			remaining--
			finished[i] = true
			release(finishConsumers[i])
			for _, consumer := range oneFinishedConsumers[i] {
				if !oneFinished[consumer] {
					oneFinished[consumer] = true
					release([]int{consumer})
				}
			}
			for nextReport < count && finished[nextReport] {
				report(nextReport)
				nextReport++
			}
		}
		grant()
	}
}
