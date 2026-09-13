package build

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/tspath"
)

func TestRootFileOwners(t *testing.T) {
	t.Parallel()
	const lib, app, tool tspath.Path = "/solution/lib/tsconfig.json", "/solution/app/tsconfig.json", "/solution/tool/tsconfig.json"
	owners := rootFileOwners{}
	owners.add("/solution/lib/src/index.ts", lib)
	owners.add("/solution/app/src/index.ts", app)
	owners.add("/solution/shared/src/index.ts", lib)
	owners.add("/solution/shared/src/index.ts", app)
	owners.add("/solution/shared/src/index.ts", app)

	tests := []struct {
		name    string
		project tspath.Path
		file    tspath.Path
		want    bool
	}{
		{"file owned by another project", app, "/solution/lib/src/index.ts", true},
		{"file owned by the asking project", lib, "/solution/lib/src/index.ts", false},
		{"file not listed by any project", app, "/solution/node_modules/dep/index.ts", false},
		{"file shared with another project", app, "/solution/shared/src/index.ts", true},
		{"file shared by other projects only", tool, "/solution/shared/src/index.ts", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := owners.checkedByAnotherProject(test.project, test.file); got != test.want {
				t.Errorf("checkedByAnotherProject(%q, %q) = %v, want %v", test.project, test.file, got, test.want)
			}
		})
	}
	if rootFileOwners(nil).checkedByAnotherProject(app, "/solution/lib/src/index.ts") {
		t.Error("nil index reported a file as checked by another project")
	}
}
