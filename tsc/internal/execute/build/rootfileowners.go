package build

import (
	"github.com/microsoft/TypeScript/tsc/internal/tspath"
)

// rootFileOwners maps every root file of the projects in a build whose type checking is enabled to
// the config path of the project that lists it, or to sharedRootFile when several projects do.
// A project consults it to skip type checking of sources it reached only through node_modules:
// those are workspace packages whose owning project reports their diagnostics.
type rootFileOwners map[tspath.Path]tspath.Path

// sharedRootFile marks a file listed as a root file by more than one project. Every project may
// rely on such a file being type checked by another one.
const sharedRootFile tspath.Path = ""

func (owners rootFileOwners) add(file tspath.Path, project tspath.Path) {
	if owner, listed := owners[file]; !listed {
		owners[file] = project
	} else if owner != project {
		owners[file] = sharedRootFile
	}
}

// checkedByAnotherProject reports whether a project other than the one at configPath type checks file.
func (owners rootFileOwners) checkedByAnotherProject(configPath tspath.Path, file tspath.Path) bool {
	owner, listed := owners[file]
	return listed && owner != configPath
}

// collectRootFileOwners indexes the root files of every project in the build order whose type
// checking is enabled. Every such project builds and reports its diagnostics in this invocation,
// except under --stopBuildOnErrors, where a project can be skipped because of upstream errors;
// the index is nil in that case so every project keeps checking all of its files.
func (o *Orchestrator) collectRootFileOwners() rootFileOwners {
	if o.opts.Command.BuildOptions.StopBuildOnErrors.IsTrue() {
		return nil
	}
	owners := rootFileOwners{}
	for _, config := range o.order {
		path := o.toPath(config)
		task := o.getTask(path)
		if task.resolved == nil || task.resolved.CompilerOptions().NoCheck.IsTrue() {
			continue
		}
		for file := range task.resolved.FileNamesByPath() {
			owners.add(file, path)
		}
	}
	return owners
}
