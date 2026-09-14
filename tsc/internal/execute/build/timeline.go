package build

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// buildTimeline is a lab instrument: when TSGO_BUILD_TIMELINE is set, every project of a build
// reports the moments it reaches its phases as tab-separated lines on standard error
// (timeline, seconds since the build started, project, event, detail).
type buildTimeline struct {
	start time.Time
	mu    sync.Mutex
}

func newBuildTimeline() *buildTimeline {
	if os.Getenv("TSGO_BUILD_TIMELINE") == "" {
		return nil
	}
	return &buildTimeline{start: time.Now()}
}

func (l *buildTimeline) event(project string, event string, detail string) {
	if l == nil {
		return
	}
	elapsed := time.Since(l.start)
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(os.Stderr, "timeline\t%.3f\t%s\t%s\t%s\n", elapsed.Seconds(), project, event, detail)
}
