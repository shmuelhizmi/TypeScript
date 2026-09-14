package build

import (
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/internal/tspath"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
)

const barrierRoot = "/home/src/workspaces/solution"

// barrierFixture is the view of an app project that references client, which references core.
// The build order is core, client, other, app, later: other and later are not referenced. Every
// project writes below its dist directory, and node_modules links to the core and other packages.
type barrierFixture struct {
	barrier                    *outputBarrier
	core, client, other, later *BuildTask
}

func newBarrierFixture() *barrierFixture {
	fs := vfstest.FromMap(map[string]any{
		barrierRoot + "/packages/core/src/index.ts":   "export {};",
		barrierRoot + "/packages/client/src/index.ts": "export {};",
		barrierRoot + "/packages/other/src/index.ts":  "export {};",
		barrierRoot + "/packages/app/src/index.ts":    "export {};",
		barrierRoot + "/packages/later/src/index.ts":  "export {};",
		barrierRoot + "/node_modules/@ws/core":        vfstest.Symlink(barrierRoot + "/packages/core"),
		barrierRoot + "/node_modules/@ws/other":       vfstest.Symlink(barrierRoot + "/packages/other"),
	}, true)
	owners := newOutputOwners(fs, tspath.ComparePathsOptions{CurrentDirectory: barrierRoot, UseCaseSensitiveFileNames: true})
	newTask := func(name string, upstream ...*BuildTask) *BuildTask {
		task := &BuildTask{config: barrierRoot + "/packages/" + name + "/tsconfig.json", done: make(chan struct{})}
		for _, upstream := range upstream {
			task.upStream = append(task.upStream, &upstreamTask{task: upstream})
		}
		owners.add(task, projectOutputs{
			directories: []string{barrierRoot + "/packages/" + name + "/dist"},
			files:       []string{barrierRoot + "/packages/" + name + "/dist/tsconfig.tsbuildinfo"},
		})
		return task
	}
	f := &barrierFixture{}
	f.core = newTask("core")
	f.client = newTask("client", f.core)
	f.other = newTask("other")
	app := newTask("app", f.client)
	f.later = newTask("later")
	f.barrier = newOutputBarrier(owners, fs, app, func(wait func()) { wait() })
	return f
}

// expectWaits runs observe and checks that it returns only once every one of tasks has finished.
func expectWaits(t *testing.T, observe func(), tasks ...*BuildTask) {
	t.Helper()
	returned := make(chan struct{})
	go func() {
		observe()
		close(returned)
	}()
	for _, task := range tasks {
		select {
		case <-returned:
			t.Fatal("observation did not wait for an unfinished earlier project")
		case <-time.After(50 * time.Millisecond):
		}
		close(task.done)
	}
	select {
	case <-returned:
	case <-time.After(30 * time.Second):
		t.Fatal("observation did not return")
	}
}

func TestOutputBarrierWaitsForEarlierOutputs(t *testing.T) {
	t.Parallel()
	t.Run("output reached through a link in node_modules", func(t *testing.T) {
		t.Parallel()
		f := newBarrierFixture()
		expectWaits(t, func() { f.barrier.FileExists(barrierRoot + "/node_modules/@ws/core/dist/index.d.ts") }, f.core)
	})
	t.Run("realpath of an output reached through a link", func(t *testing.T) {
		t.Parallel()
		f := newBarrierFixture()
		expectWaits(t, func() { f.barrier.Realpath(barrierRoot + "/node_modules/@ws/core/dist/index.d.ts") }, f.core)
	})
	t.Run("output of a directly referenced project", func(t *testing.T) {
		t.Parallel()
		f := newBarrierFixture()
		expectWaits(t, func() { f.barrier.ReadFile(barrierRoot + "/packages/client/dist/index.d.ts") }, f.client)
	})
	t.Run("output of an earlier project that is not referenced", func(t *testing.T) {
		t.Parallel()
		f := newBarrierFixture()
		expectWaits(t, func() { f.barrier.FileExists(barrierRoot + "/node_modules/@ws/other/dist/index.d.ts") }, f.other)
	})
	t.Run("build info of a referenced project", func(t *testing.T) {
		t.Parallel()
		f := newBarrierFixture()
		expectWaits(t, func() { f.barrier.FileExists(barrierRoot + "/packages/core/dist/tsconfig.tsbuildinfo") }, f.core)
	})
	t.Run("output directory that does not exist yet", func(t *testing.T) {
		t.Parallel()
		f := newBarrierFixture()
		expectWaits(t, func() { f.barrier.DirectoryExists(barrierRoot + "/packages/core/dist") }, f.core)
	})
	t.Run("listing of a package directory reached through a link", func(t *testing.T) {
		t.Parallel()
		f := newBarrierFixture()
		expectWaits(t, func() { f.barrier.GetAccessibleEntries(barrierRoot + "/node_modules/@ws/core") }, f.core)
	})
	t.Run("listing above the outputs of every earlier project", func(t *testing.T) {
		t.Parallel()
		f := newBarrierFixture()
		expectWaits(t, func() { f.barrier.GetAccessibleEntries(barrierRoot + "/packages") }, f.core, f.client, f.other)
	})
	t.Run("every referenced project before emit", func(t *testing.T) {
		t.Parallel()
		f := newBarrierFixture()
		expectWaits(t, f.barrier.waitForUpstream, f.core, f.client)
	})
}

func TestOutputBarrierDoesNotWaitForOtherObservations(t *testing.T) {
	t.Parallel()
	t.Run("output of a later project", func(t *testing.T) {
		t.Parallel()
		f := newBarrierFixture()
		expectWaits(t, func() { f.barrier.FileExists(barrierRoot + "/packages/later/dist/index.d.ts") })
	})
	t.Run("source of a referenced project", func(t *testing.T) {
		t.Parallel()
		f := newBarrierFixture()
		expectWaits(t, func() { f.barrier.ReadFile(barrierRoot + "/packages/core/src/index.ts") })
	})
	t.Run("package directory that exists", func(t *testing.T) {
		t.Parallel()
		f := newBarrierFixture()
		expectWaits(t, func() { f.barrier.DirectoryExists(barrierRoot + "/packages/core") })
	})
	t.Run("finished referenced project", func(t *testing.T) {
		t.Parallel()
		f := newBarrierFixture()
		close(f.core.done)
		expectWaits(t, func() { f.barrier.FileExists(barrierRoot + "/packages/core/dist/index.d.ts") })
	})
}

func TestCanonicalPathsResolveLinksInExistingDirectories(t *testing.T) {
	t.Parallel()
	canonical := &canonicalPaths{fs: vfstest.FromMap(map[string]any{
		barrierRoot + "/packages/core/src/index.ts": "export {};",
		barrierRoot + "/node_modules/@ws/core":      vfstest.Symlink(barrierRoot + "/packages/core"),
	}, true)}
	for path, expected := range map[string]string{
		barrierRoot + "/node_modules/@ws/core/dist/index.d.ts": barrierRoot + "/packages/core/dist/index.d.ts",
		barrierRoot + "/node_modules/@ws/core/src/index.ts":    barrierRoot + "/packages/core/src/index.ts",
		barrierRoot + "/packages/core/dist/index.d.ts":         barrierRoot + "/packages/core/dist/index.d.ts",
		barrierRoot + "/packages/core/src/index.ts":            barrierRoot + "/packages/core/src/index.ts",
	} {
		if resolved := canonical.resolve(path); resolved != expected {
			t.Errorf("resolve(%s) = %s, want %s", path, resolved, expected)
		}
	}
	if resolved, expected := canonical.directory(barrierRoot+"/node_modules/@ws/core"), barrierRoot+"/packages/core"; resolved != expected {
		t.Errorf("directory(%s) = %s, want %s", barrierRoot+"/node_modules/@ws/core", resolved, expected)
	}
}
