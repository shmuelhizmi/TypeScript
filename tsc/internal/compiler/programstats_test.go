package compiler

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
	"gotest.tools/v3/assert"
)

func TestCollectProgramStats(t *testing.T) {
	t.Parallel()

	fs := vfstest.FromMap(map[string]any{
		"/src/index.ts":  "import { box, Box } from \"./lib\";\nimport { other } from \"./other\";\nexport const b: Box<number> = box(other);\n",
		"/src/second.ts": "import { Box } from \"./lib\";\nexport declare const s: Box<string>;\n",
		"/src/other.ts":  "export const other = 1;\n",
		"/src/lib.d.ts": "export declare function box<T>(value: T): Box<T>;\n" +
			"export interface Box<T> { value: T; map<U>(f: (value: T) => U): Box<U>; }\n" +
			"export type Unbox<T> = T extends Box<infer U> ? U : never;\n" +
			"export type Keys<T> = { [K in keyof T]: T[K] };\n" +
			"export type Prefixed<T extends string> = `p-${T}`;\n",
	}, true /*useCaseSensitiveFileNames*/)
	program := NewProgram(ProgramOptions{
		Config: &tsoptions.ParsedCommandLine{
			ParsedConfig: &tsoptions.ParsedOptions{
				FileNames:       []string{"/src/index.ts", "/src/second.ts"},
				CompilerOptions: &core.CompilerOptions{NoLib: core.TSTrue},
			},
		},
		Host: NewCompilerHost("/src", fs, "", nil, nil, nil),
	})

	stats := collectProgramStats(program, 4)

	assert.Equal(t, stats.CheckerCount, 4)
	assert.Equal(t, stats.Files, 4)
	assert.Equal(t, stats.Source.Files, 3)
	assert.Equal(t, stats.Source.Imports, 3)
	assert.Equal(t, stats.Source.GenericReferences, 2)

	lib := stats.Declaration
	assert.Equal(t, lib.Files, 1)
	assert.Equal(t, lib.TypeParameters, 8, "declared, infer and mapped-type parameters")
	assert.Equal(t, lib.ConditionalTypes, 1)
	assert.Equal(t, lib.MappedTypes, 1)
	assert.Equal(t, lib.InferTypes, 1)
	assert.Equal(t, lib.IndexedAccessTypes, 1)
	assert.Equal(t, lib.TemplateLiteralTypes, 1)
	assert.Equal(t, lib.TypeOperators, 1)
	assert.Equal(t, lib.TypeReferences, 15)
	assert.Equal(t, lib.GenericReferences, 3)
	assert.Equal(t, lib.Signatures, 3)
	assert.Equal(t, lib.GenericSignatures, 2)
	assert.Equal(t, lib.Interfaces, 1)
	assert.Equal(t, lib.TypeAliases, 3)
	assert.Equal(t, lib.generic(), 13)

	total := float64(lib.Weight + stats.Source.Weight)
	assert.Equal(t, stats.DeclarationShare, float64(lib.Weight)/total)
	assert.Equal(t, len(stats.TopImported), 1)
	top := stats.TopImported[0]
	assert.Equal(t, top.File, "/src/lib.d.ts")
	assert.Equal(t, top.Importers, 2, "index.ts and second.ts import lib.d.ts directly; other.ts is a source file")
	assert.Equal(t, top.Demand, 2.0/3.0)
	assert.Equal(t, top.Generic, 13)
	assert.Equal(t, stats.DemandWeight, float64(lib.Weight)*2.0/3.0)
	assert.Equal(t, stats.DemandGeneric, 13*2.0/3.0)
}
