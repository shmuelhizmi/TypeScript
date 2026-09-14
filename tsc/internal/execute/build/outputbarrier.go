package build

import (
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/microsoft/TypeScript/tsc/internal/collections"
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
	"github.com/microsoft/TypeScript/tsc/internal/vfs"
)

// projectOutputs is where a project writes: the directories it owns as a whole and the files it
// writes elsewhere.
type projectOutputs struct {
	directories []string
	files       []string
}

// outputOwners indexes the outputs of the projects of a build by the canonical form of their paths
// (see canonicalPaths), so that a path reaching an output through a linked package directory in
// node_modules names the same owner as the output itself.
type outputOwners struct {
	canonical   *canonicalPaths
	options     tspath.ComparePathsOptions
	tasks       []*BuildTask                 // in build order
	index       map[*BuildTask]int           // position in tasks
	directories map[tspath.Path][]*BuildTask // a project writes anywhere below these
	files       map[tspath.Path][]*BuildTask // a project writes these outside the directories it owns
	enclosing   map[tspath.Path][]*BuildTask // proper ancestors of owned paths: a project may create entries there
}

func newOutputOwners(fs vfs.FS, options tspath.ComparePathsOptions) *outputOwners {
	return &outputOwners{
		canonical:   &canonicalPaths{fs: fs},
		options:     options,
		index:       map[*BuildTask]int{},
		directories: map[tspath.Path][]*BuildTask{},
		files:       map[tspath.Path][]*BuildTask{},
		enclosing:   map[tspath.Path][]*BuildTask{},
	}
}

// collectOutputOwners indexes what every project of the build order writes: its outDir and
// declarationDir as a whole, its build info, and, when it has no outDir, its outputs next to the
// sources.
func (o *Orchestrator) collectOutputOwners() *outputOwners {
	owners := newOutputOwners(o.opts.Sys.FS(), o.comparePathsOptions)
	for _, config := range o.order {
		task := o.getTask(o.toPath(config))
		if task.resolved == nil {
			continue
		}
		options := task.resolved.CompilerOptions()
		var outputs projectOutputs
		if options.OutDir != "" {
			outputs.directories = append(outputs.directories, options.OutDir)
		}
		if options.DeclarationDir != "" {
			outputs.directories = append(outputs.directories, options.DeclarationDir)
		}
		if buildInfo := task.resolved.GetBuildInfoFileName(); buildInfo != "" {
			outputs.files = append(outputs.files, buildInfo)
		}
		if options.OutDir == "" {
			outputs.files = slices.AppendSeq(outputs.files, task.resolved.GetOutputFileNames())
		}
		owners.add(task, outputs)
	}
	return owners
}

// add records the outputs of the next task in build order.
func (owners *outputOwners) add(task *BuildTask, outputs projectOutputs) {
	owners.index[task] = len(owners.tasks)
	owners.tasks = append(owners.tasks, task)
	for _, directory := range outputs.directories {
		path, _ := owners.directoryPath(directory)
		owners.directories[path] = appendOwner(owners.directories[path], task)
		owners.addEnclosing(path, task)
	}
	for _, file := range outputs.files {
		path := owners.filePath(file)
		owners.files[path] = appendOwner(owners.files[path], task)
		owners.addEnclosing(path, task)
	}
}

func (owners *outputOwners) addEnclosing(path tspath.Path, task *BuildTask) {
	for {
		parent := path.GetDirectoryPath()
		if parent == path {
			return
		}
		path = parent
		owners.enclosing[path] = appendOwner(owners.enclosing[path], task)
	}
}

// appendOwner adds task to owners unless it is the last one already: the outputs of a project are
// added together.
func appendOwner(owners []*BuildTask, task *BuildTask) []*BuildTask {
	if n := len(owners); n > 0 && owners[n-1] == task {
		return owners
	}
	return append(owners, task)
}

// filePath is the canonical path of a file observation. Links in the directory of path are
// resolved but path itself is not, so a file a project is going to write is named where it will be.
func (owners *outputOwners) filePath(path string) tspath.Path {
	return owners.toPath(owners.canonical.resolve(owners.absolute(path)))
}

// directoryPath is the canonical path of a directory observation, following a link at path itself,
// and the resolved path it was computed from.
func (owners *outputOwners) directoryPath(path string) (tspath.Path, string) {
	resolved := owners.canonical.directory(owners.absolute(path))
	return owners.toPath(resolved), resolved
}

func (owners *outputOwners) absolute(path string) string {
	return tspath.GetNormalizedAbsolutePath(path, owners.options.CurrentDirectory)
}

func (owners *outputOwners) toPath(path string) tspath.Path {
	return tspath.ToPath(path, owners.options.CurrentDirectory, owners.options.UseCaseSensitiveFileNames)
}

// owning returns the projects whose outputs an observation of path can show: those owning path or
// a directory above it and, for a listing, those that may create entries below it.
func (owners *outputOwners) owning(path tspath.Path, listing bool) []*BuildTask {
	var result []*BuildTask
	result = append(result, owners.files[path]...)
	for directory := path; ; {
		result = append(result, owners.directories[directory]...)
		parent := directory.GetDirectoryPath()
		if parent == directory {
			break
		}
		directory = parent
	}
	if listing {
		result = append(result, owners.enclosing[path]...)
	}
	return result
}

// writtenByEarlierProject reports whether a project before task in build order writes file.
func (owners *outputOwners) writtenByEarlierProject(file string, task *BuildTask) bool {
	return slices.ContainsFunc(owners.files[owners.filePath(file)], func(owner *BuildTask) bool {
		return owners.index[owner] < owners.index[task]
	})
}

// canonicalPaths resolves the links in the existing part of paths. A build creates files and
// directories but no links, so a directory once resolved stays resolved for the whole build.
type canonicalPaths struct {
	fs          vfs.FS
	directories collections.SyncMap[string, string]
}

// resolve returns path with the links in its directory resolved. path itself is not followed.
func (c *canonicalPaths) resolve(path string) string {
	directory := tspath.GetDirectoryPath(path)
	if directory == path {
		return path
	}
	return tspath.CombinePaths(c.directory(directory), tspath.GetBaseFileName(path))
}

// directory returns directory with its links resolved, whether or not it exists yet.
func (c *canonicalPaths) directory(directory string) string {
	if resolved, ok := c.directories.Load(directory); ok {
		return resolved
	}
	resolved := c.fs.Realpath(directory)
	if resolved == directory {
		// Realpath returns its argument both for a directory without links and for one that does
		// not exist yet, so the parent decides.
		if parent := tspath.GetDirectoryPath(directory); parent != directory {
			resolved = tspath.CombinePaths(c.directory(parent), tspath.GetBaseFileName(directory))
		}
	}
	c.directories.Store(directory, resolved)
	return resolved
}

// canOverlapUpstream reports whether task may start before the projects it references have
// finished, waiting for a project earlier in the build order only when its program observes that
// project's outputs (see outputBarrier). That reproduces a build that waits upfront only when
// nothing else the task does depends on those projects: it is built in full, because it has no
// build info or --force is given, so its up-to-date status reads none of their outputs; no project
// is skipped because of upstream errors; the build is neither a dry run nor a watch; no content
// mapper reads its files outside the compiler's file system; and no earlier project writes its
// build info.
func (o *Orchestrator) canOverlapUpstream(task *BuildTask) bool {
	if len(task.upStream) == 0 || task.resolved == nil || len(task.resolved.FileNames()) == 0 || task.status != nil ||
		len(task.resolved.ContentMappers()) != 0 ||
		o.builderCount() <= 1 ||
		o.opts.Command.BuildOptions.StopBuildOnErrors.IsTrue() ||
		o.opts.Command.BuildOptions.Dry.IsTrue() ||
		o.opts.Command.CompilerOptions.Watch.IsTrue() {
		return false
	}
	buildInfo := task.resolved.GetBuildInfoFileName()
	if buildInfo == "" || o.outputOwners.writtenByEarlierProject(buildInfo, task) {
		return false
	}
	return o.opts.Command.BuildOptions.Force.IsTrue() || !o.opts.Sys.FS().FileExists(buildInfo)
}

// referencedTasks returns the tasks of the projects this project references, directly or through
// other references.
func (t *BuildTask) referencedTasks() []*BuildTask {
	var tasks []*BuildTask
	var seen collections.Set[*BuildTask]
	var collect func(task *BuildTask)
	collect = func(task *BuildTask) {
		for _, upstream := range task.upStream {
			if seen.AddIfAbsent(upstream.task) {
				tasks = append(tasks, upstream.task)
				collect(upstream.task)
			}
		}
	}
	collect(t)
	return tasks
}

func (t *BuildTask) finished() bool {
	select {
	case <-t.done:
		return true
	default:
		return false
	}
}

// outputBarrier is the file system view of a project that started before the projects it references
// finished. Every observation of what an unfinished project earlier in the build order may write
// waits for that project first, so the program sees what a build that waited upfront would have
// seen. Projects later in the build order are not waited for, as in a build that waits upfront, and
// waiting only for earlier projects cannot deadlock.
type outputBarrier struct {
	vfs.FS
	owners     *outputOwners
	index      int          // position of the project in build order
	referenced []*BuildTask // the projects the project references, directly or through other references
	suspend    func(wait func())
	now        func() time.Time
	waiting    sync.Mutex    // one wait at a time: the task's worker is released once
	waited     time.Duration // time spent waiting so far, updated under waiting
	settled    atomic.Bool   // every earlier project has finished; nothing is left to wait for
}

var _ vfs.FS = (*outputBarrier)(nil)

func newOutputBarrier(owners *outputOwners, fs vfs.FS, task *BuildTask, suspend func(wait func()), now func() time.Time) *outputBarrier {
	b := &outputBarrier{FS: fs, owners: owners, index: owners.index[task], referenced: task.referencedTasks(), suspend: suspend, now: now}
	b.settle()
	return b
}

// settle records that nothing is left to wait for once every earlier project has finished.
func (b *outputBarrier) settle() {
	if len(b.unfinished(b.owners.tasks[:b.index])) == 0 {
		b.settled.Store(true)
	}
}

// observeFile waits for the projects whose outputs an observation of the file at path can show.
func (b *outputBarrier) observeFile(path string) {
	if b.settled.Load() {
		return
	}
	b.wait(b.owners.owning(b.owners.filePath(path), false))
}

// observeDirectory waits like observeFile for the directory at path, following a link at path
// itself, and also for the projects that may create the directory when it does not exist yet.
func (b *outputBarrier) observeDirectory(path string) {
	if b.settled.Load() {
		return
	}
	directory, resolved := b.owners.directoryPath(path)
	b.wait(b.owners.owning(directory, false))
	if enclosing := b.unfinished(b.owners.enclosing[directory]); len(enclosing) != 0 && !b.owners.canonical.fs.DirectoryExists(resolved) {
		b.wait(enclosing)
	}
}

// observeListing waits like observeDirectory and also for the projects that may create entries
// anywhere below the directory at path.
func (b *outputBarrier) observeListing(path string) {
	if b.settled.Load() {
		return
	}
	directory, _ := b.owners.directoryPath(path)
	b.wait(b.owners.owning(directory, true))
}

// waitForUpstream waits for every referenced project to finish.
func (b *outputBarrier) waitForUpstream() {
	if b.settled.Load() {
		return
	}
	b.wait(b.referenced)
}

// wait waits for the earlier projects among owners that have not finished, with the task's worker
// released.
func (b *outputBarrier) wait(owners []*BuildTask) {
	if len(owners) == 0 {
		return
	}
	if pending := b.unfinished(owners); len(pending) != 0 {
		b.waiting.Lock()
		defer b.waiting.Unlock()
		if pending = b.unfinished(pending); len(pending) != 0 {
			start := b.now()
			b.suspend(func() {
				for _, task := range pending {
					<-task.done
				}
			})
			b.waited += b.now().Sub(start)
		}
	}
	b.settle()
}

// waitedTime returns the time spent waiting for other projects so far.
func (b *outputBarrier) waitedTime() time.Duration {
	b.waiting.Lock()
	defer b.waiting.Unlock()
	return b.waited
}

// unfinished returns the projects among owners that are earlier in the build order and have not
// finished.
func (b *outputBarrier) unfinished(owners []*BuildTask) []*BuildTask {
	var unfinished []*BuildTask
	for _, owner := range owners {
		if b.owners.index[owner] < b.index && !owner.finished() {
			unfinished = append(unfinished, owner)
		}
	}
	return unfinished
}

func (b *outputBarrier) FileExists(path string) bool {
	b.observeFile(path)
	return b.FS.FileExists(path)
}

func (b *outputBarrier) ReadFile(path string) (contents string, ok bool) {
	b.observeFile(path)
	return b.FS.ReadFile(path)
}

func (b *outputBarrier) Realpath(path string) string {
	// Realpath returns its argument for a path that does not exist yet.
	b.observeFile(path)
	return b.FS.Realpath(path)
}

func (b *outputBarrier) Stat(path string) vfs.FileInfo {
	// The modification time of a directory changes with its entries.
	b.observeListing(path)
	return b.FS.Stat(path)
}

func (b *outputBarrier) DirectoryExists(path string) bool {
	b.observeDirectory(path)
	return b.FS.DirectoryExists(path)
}

func (b *outputBarrier) GetAccessibleEntries(path string) vfs.Entries {
	b.observeListing(path)
	return b.FS.GetAccessibleEntries(path)
}

func (b *outputBarrier) WalkDir(root string, walkFn vfs.WalkDirFunc) error {
	b.observeListing(root)
	return b.FS.WalkDir(root, walkFn)
}
