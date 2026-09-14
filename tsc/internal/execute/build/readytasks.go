package build

// taskDependencies lists, for one task of a build, the tasks it must not start before: those that
// must have finished and those that must merely have started. Dependencies have lower indices than
// the tasks that depend on them.
type taskDependencies struct {
	finished []int
	started  []int
}

// executeReadyBuildTasks calls execute for every task, starting a task as soon as its dependencies
// are satisfied and running at most workerCount tasks at a time, and calls report for the tasks in
// index order as they complete. Waiting for dependencies and reporting happen on the calling
// goroutine, so a task whose report is pending never occupies a worker. A running task that has to
// wait for another task passes the wait to the suspend function it received: the wait runs with the
// task's worker released, and the task takes a worker again afterwards.
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
	for i, task := range dependencies {
		pending[i] = len(task.finished) + len(task.started)
		for _, dependency := range task.finished {
			finishConsumers[dependency] = append(finishConsumers[dependency], i)
		}
		for _, dependency := range task.started {
			startConsumers[dependency] = append(startConsumers[dependency], i)
		}
	}
	workers := make(chan struct{}, workerCount)
	started := make(chan int)
	completed := make(chan int)
	start := func(i int) {
		go func() {
			workers <- struct{}{}
			started <- i
			execute(i, func(wait func()) {
				<-workers
				wait()
				workers <- struct{}{}
			})
			<-workers
			completed <- i
		}()
	}
	release := func(consumers []int) {
		for _, consumer := range consumers {
			pending[consumer]--
			if pending[consumer] == 0 {
				start(consumer)
			}
		}
	}
	for i := range count {
		if pending[i] == 0 {
			start(i)
		}
	}
	finished := make([]bool, count)
	nextReport := 0
	for remaining := count; remaining > 0; {
		select {
		case i := <-started:
			release(startConsumers[i])
		case i := <-completed:
			remaining--
			finished[i] = true
			release(finishConsumers[i])
			for nextReport < count && finished[nextReport] {
				report(nextReport)
				nextReport++
			}
		}
	}
}
