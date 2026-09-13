package compiler

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
	"gotest.tools/v3/assert"
)

func TestCheckerCountForGenericPressure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                       string
		reachableGenericConstructs int
		sourceWeight               int
		want                       int
	}{
		{"no reachable generic constructs", 0, 1, sourceDominatedCheckerCount},
		{"few constructs over much source", 7_000, 16_000_000, sourceDominatedCheckerCount},
		{"under the threshold", 839, 393_000, sourceDominatedCheckerCount},
		{"exactly at the threshold", 10, 1_000, sourceDominatedCheckerCount},
		{"just above the threshold", 11, 1_000, baseCheckerCount},
		{"many constructs over little source", 134_000, 1_700_000, baseCheckerCount},
		{"a library-only program has no source weight", 500, 0, baseCheckerCount},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, checkerCountForGenericPressure(test.reachableGenericConstructs, test.sourceWeight), test.want)
		})
	}
}

func TestDefaultCheckerCount(t *testing.T) {
	t.Parallel()
	generic := "export declare function box<T>(value: T): Box<T>;\nexport interface Box<T> { value: T }\n" +
		"export type Unbox<T> = T extends Box<infer U> ? U : never;\n"
	plain := "export declare function box(value: number): number;\nexport interface Box { value: number }\n"
	// Source weight is deliberately large (text counts toward the association weight) so that the plain
	// library keeps the program source-dominated while the generic one does not.
	source := "import { box } from \"./lib\";\nexport const b = box(1);\n" + strings.Repeat("// padding to give the file weight\n", 40)
	tests := []struct {
		name  string
		files map[string]any
		want  int
	}{
		{
			name:  "source over a plain library",
			files: map[string]any{"/src/index.ts": source, "/src/lib.d.ts": plain},
			want:  sourceDominatedCheckerCount,
		},
		{
			name:  "source over a generic library",
			files: map[string]any{"/src/index.ts": source, "/src/lib.d.ts": generic},
			want:  baseCheckerCount,
		},
		{
			name:  "a generic library nothing imports does not count",
			files: map[string]any{"/src/index.ts": source, "/src/lib.d.ts": plain, "/src/unused.d.ts": generic},
			want:  sourceDominatedCheckerCount,
		},
		{
			name:  "a generic library reached through another source file counts",
			files: map[string]any{"/src/index.ts": "import { b } from \"./middle\";\nexport const c = b;\n" + strings.Repeat("// padding\n", 40), "/src/middle.ts": source, "/src/lib.d.ts": generic},
			want:  baseCheckerCount,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fs := vfstest.FromMap(test.files, true /*useCaseSensitiveFileNames*/)
			fileNames := make([]string, 0, len(test.files))
			for name := range test.files {
				fileNames = append(fileNames, name)
			}
			program := NewProgram(ProgramOptions{
				Config: &tsoptions.ParsedCommandLine{
					ParsedConfig: &tsoptions.ParsedOptions{
						FileNames:       fileNames,
						CompilerOptions: &core.CompilerOptions{NoLib: core.TSTrue},
					},
				},
				Host: NewCompilerHost("/src", fs, "", nil, nil, nil),
			})
			assert.Equal(t, defaultCheckerCount(program), test.want)
		})
	}
}
