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

	// index.ts and second.ts share a checker; other.ts and lib.d.ts have their own.
	checkerOf := map[string]int{"/src/index.ts": 0, "/src/second.ts": 0, "/src/other.ts": 1, "/src/lib.d.ts": 2}
	associations := make([]int, len(program.files))
	for i, file := range program.files {
		associations[i] = checkerOf[file.FileName()]
	}
	stats := collectProgramStats(program, 4, associations)

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

	assert.Equal(t, len(stats.Checkers), 4)
	first := stats.Checkers[0]
	assert.Equal(t, first.SourceFiles, 2)
	assert.Equal(t, first.SourceGeneric, 0)
	assert.Equal(t, first.ReachableDeclarationFiles, 1, "lib.d.ts is reached from index.ts and second.ts, counted once")
	assert.Equal(t, first.ReachableDeclarationWeight, lib.Weight)
	assert.Equal(t, first.ReachableDeclarationGeneric, 13)
	assert.Equal(t, stats.Checkers[1].SourceFiles, 1)
	assert.Equal(t, stats.Checkers[1].ReachableDeclarationFiles, 0, "other.ts imports nothing")
	assert.Equal(t, stats.Checkers[2].SourceFiles, 0, "a checker owning only a declaration file has no source reach")
	assert.Equal(t, stats.Checkers[2].ReachableDeclarationFiles, 0)
	assert.Equal(t, stats.Checkers[3].SourceFiles, 0)
}
