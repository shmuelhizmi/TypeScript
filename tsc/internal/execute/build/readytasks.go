package build

// executeReadyBuildTasks calls execute for every task, starting a task as soon as the tasks it
// depends on have finished and running at most workerCount tasks at a time, and calls report for
// the tasks in index order as they complete. Dependencies must have lower indices than the tasks
// that depend on them. Workers only execute: waiting for dependencies and reporting happen on the
// calling goroutine, so a task whose report is pending never occupies a worker.
func executeReadyBuildTasks(dependencies [][]int, workerCount int, execute func(int), report func(int)) {
	count := len(dependencies)
	if count == 0 {
		return
	}
	if workerCount <= 1 {
		for i := range count {
			execute(i)
			report(i)
		}
		return
	}
	workerCount = min(workerCount, count)
	pending := make([]int, count)
	consumers := make([][]int, count)
	ready := make(chan int, count)
	completed := make(chan int, workerCount)
	stopped := make(chan struct{}, workerCount)
	for i, upstream := range dependencies {
		pending[i] = len(upstream)
		for _, dependency := range upstream {
			consumers[dependency] = append(consumers[dependency], i)
		}
		if len(upstream) == 0 {
			ready <- i
		}
	}
	for range workerCount {
		go func() {
			for i := range ready {
				execute(i)
				completed <- i
			}
			stopped <- struct{}{}
		}()
	}
	finished := make([]bool, count)
	nextReport := 0
	for range count {
		i := <-completed
		finished[i] = true
		for _, consumer := range consumers[i] {
			pending[consumer]--
			if pending[consumer] == 0 {
				ready <- consumer
			}
		}
		for nextReport < count && finished[nextReport] {
			report(nextReport)
			nextReport++
		}
	}
	close(ready)
	for range workerCount {
		<-stopped
	}
}
