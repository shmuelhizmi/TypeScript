package compiler

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/bundled"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
	"gotest.tools/v3/assert"
)

func newRootOrderLoader(t *testing.T) *fileLoader {
	t.Helper()
	fs := bundled.WrapFS(vfstest.FromMap(map[string]string{}, false /*useCaseSensitiveFileNames*/))
	return &fileLoader{opts: ProgramOptions{Host: NewCompilerHost("/", fs, bundled.LibPath(), nil, nil, nil)}}
}

func rootTaskPaths(tasks []*parseTask) []string {
	paths := make([]string, len(tasks))
	for i, task := range tasks {
		paths[i] = task.normalizedFilePath
	}
	return paths
}

// TestRootTasksInStartOrder covers the order the root tasks are started in, which a parallel work group
// follows and which decides what the largest files wait behind. It is only a start order: the files a program
// ends up holding are collected from rootTasks, in the order the roots were given.
func TestRootTasksInStartOrder(t *testing.T) {
	t.Parallel()
	loader := newRootOrderLoader(t)
	first := &parseTask{normalizedFilePath: "/first.ts"}
	second := &parseTask{normalizedFilePath: "/second.ts"}
	lib := &parseTask{normalizedFilePath: "/lib.es5.d.ts", libFile: &LibFile{Name: "lib.es5.d.ts"}}

	// A parallel build starts the lib files, which are the largest, before the other roots.
	loader.rootTasks = []*parseTask{first, second, lib}
	assert.DeepEqual(t, rootTaskPaths(loader.rootTasksInStartOrder(false /*singleThreaded*/)),
		[]string{"/lib.es5.d.ts", "/first.ts", "/second.ts"})

	// A single-threaded build keeps the order it was given: its work group runs the last task queued first,
	// and --generateTrace records the order it parses the files in.
	assert.DeepEqual(t, rootTaskPaths(loader.rootTasksInStartOrder(true /*singleThreaded*/)),
		[]string{"/first.ts", "/second.ts", "/lib.es5.d.ts"})

	// With no lib file there is nothing to move.
	loader.rootTasks = []*parseTask{first, second}
	assert.DeepEqual(t, rootTaskPaths(loader.rootTasksInStartOrder(false /*singleThreaded*/)),
		[]string{"/first.ts", "/second.ts"})

	// A root task that names the same file as a lib file keeps the order it was given: the first task started
	// for a path owns it, and starting the lib file first would hand it the root task's place.
	shared := &parseTask{normalizedFilePath: "/lib.es5.d.ts"}
	loader.rootTasks = []*parseTask{first, shared, lib}
	assert.DeepEqual(t, rootTaskPaths(loader.rootTasksInStartOrder(false /*singleThreaded*/)),
		[]string{"/first.ts", "/lib.es5.d.ts", "/lib.es5.d.ts"})
	assert.Equal(t, loader.rootTasksInStartOrder(false /*singleThreaded*/)[1], shared)
}

// rootOrderRoots is deliberately in reverse alphabetical order, so that a list that came out sorted, or in the
// order seven concurrent resolutions happened to finish in, would not pass for the order they were given in.
var rootOrderRoots = []string{"g.ts", "f.ts", "e.ts", "d.ts", "c.ts", "b.ts", "a.ts"}

func rootOrderProgramFiles(t *testing.T, singleThreaded core.Tristate) []string {
	t.Helper()
	files := map[string]string{
		"/tsconfig.json": `{ "compilerOptions": { "strict": true }, "files": ["g.ts", "f.ts", "e.ts", "d.ts", "c.ts", "b.ts", "a.ts"] }`,
	}
	for _, name := range rootOrderRoots {
		files["/"+name] = "export const " + name[:1] + " = 1;\n"
	}
	fs := bundled.WrapFS(vfstest.FromMap(files, false /*useCaseSensitiveFileNames*/))
	host := NewCompilerHost("/", fs, bundled.LibPath(), nil, nil, nil)
	parsed, errors := tsoptions.GetParsedCommandLineOfConfigFile("/tsconfig.json", &core.CompilerOptions{}, nil, host, nil)
	assert.Equal(t, len(errors), 0, "expected no errors in parsed command line")
	program := NewProgram(ProgramOptions{Config: parsed, Host: host, SingleThreaded: singleThreaded})
	names := make([]string, 0, len(program.GetSourceFiles()))
	for _, file := range program.GetSourceFiles() {
		names = append(names, file.FileName())
	}
	return names
}

// TestRootFileOrderSurvivesConcurrentResolution covers what the concurrent resolution must not disturb: each
// root is resolved into the slot its index names, and the program's files are collected from the root tasks in
// the order the configuration gave them, whatever order the resolutions finished in and whichever roots were
// started first.
func TestRootFileOrderSurvivesConcurrentResolution(t *testing.T) {
	t.Parallel()
	parallel := rootOrderProgramFiles(t, core.TSFalse)
	assert.DeepEqual(t, parallel, rootOrderProgramFiles(t, core.TSTrue))

	roots := make([]string, 0, len(rootOrderRoots))
	for _, name := range parallel {
		if !strings.HasPrefix(name, "bundled:") {
			roots = append(roots, name)
		}
	}
	want := make([]string, len(rootOrderRoots))
	for i, name := range rootOrderRoots {
		want[i] = "/" + name
	}
	assert.DeepEqual(t, roots, want)
}
