package compiler

import (
	"fmt"
	"os"
	"slices"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/json"
)

// Research instrument (not for release): when TSGO_PROGRAM_STATS is set, creating a checker pool prints one
// `program-stats {json}` line to stderr describing the program's composition (declaration versus source files, the
// association weights, syntactic generic constructs) and the direct import demand of each declaration file. It is used
// to study which program statistics predict whether more checkers pay off. Nothing runs when the variable is unset.

const programStatsEnv = "TSGO_PROGRAM_STATS"

const programStatsTopImported = 30

type programStatsFileKind struct {
	Files                int `json:"files"`
	Nodes                int `json:"nodes"`
	Bytes                int `json:"bytes"`
	Weight               int `json:"weight"`
	Imports              int `json:"imports"`
	TypeParameters       int `json:"typeParameters"`
	ConditionalTypes     int `json:"conditionalTypes"`
	MappedTypes          int `json:"mappedTypes"`
	InferTypes           int `json:"inferTypes"`
	IndexedAccessTypes   int `json:"indexedAccessTypes"`
	TemplateLiteralTypes int `json:"templateLiteralTypes"`
	TypeOperators        int `json:"typeOperators"`
	TypeQueries          int `json:"typeQueries"`
	TypeReferences       int `json:"typeReferences"`
	GenericReferences    int `json:"genericReferences"`
	Signatures           int `json:"signatures"`
	GenericSignatures    int `json:"genericSignatures"`
	Interfaces           int `json:"interfaces"`
	TypeAliases          int `json:"typeAliases"`
	Classes              int `json:"classes"`
}

// generic is the count of constructs whose instantiation creates new types: type parameters, conditional, mapped,
// infer and indexed-access types, template literal types.
func (k *programStatsFileKind) generic() int {
	return k.TypeParameters + k.ConditionalTypes + k.MappedTypes + k.InferTypes + k.IndexedAccessTypes + k.TemplateLiteralTypes
}

type programStatsImportedFile struct {
	File      string  `json:"file"`
	Weight    int     `json:"weight"`
	Importers int     `json:"importers"`
	Demand    float64 `json:"demand"`
	Generic   int     `json:"generic"`
}

// programStatsChecker describes one checker's share of the program: the source files associated with it and the
// declaration files those source files reach through imports (transitively, through source and declaration files).
// The reachable declaration weight is the static proxy for the library work every checker repeats.
type programStatsChecker struct {
	Checker                     int `json:"checker"`
	SourceFiles                 int `json:"sourceFiles"`
	SourceWeight                int `json:"sourceWeight"`
	SourceGeneric               int `json:"sourceGeneric"`
	ReachableDeclarationFiles   int `json:"reachableDeclarationFiles"`
	ReachableDeclarationWeight  int `json:"reachableDeclarationWeight"`
	ReachableDeclarationGeneric int `json:"reachableDeclarationGeneric"`
}

type programStats struct {
	CheckerCount     int                  `json:"checkerCount"`
	Files            int                  `json:"files"`
	Declaration      programStatsFileKind `json:"declaration"`
	Source           programStatsFileKind `json:"source"`
	DeclarationShare float64              `json:"declarationShare"`
	// DemandWeight is the sum over declaration files of weight × (direct source importers / source files): the
	// declaration weight an average source file pulls in directly. DemandGeneric is the same sum over generic().
	DemandWeight  float64                    `json:"demandWeight"`
	DemandGeneric float64                    `json:"demandGeneric"`
	TopImported   []programStatsImportedFile `json:"topImported"`
	Checkers      []programStatsChecker      `json:"checkers"`
}

func programStatsEnabled() bool {
	return os.Getenv(programStatsEnv) != ""
}

func (p *checkerPool) reportProgramStats(associations []int) {
	if !programStatsEnabled() {
		return
	}
	stats := collectProgramStats(p.program, len(p.checkers), associations)
	encoded, err := json.Marshal(stats)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "program-stats %s\n", encoded)
}

// collectProgramStats computes the statistics; associations[i] is the checker index of program.files[i].
func collectProgramStats(program *Program, checkerCount int, associations []int) *programStats {
	files := program.files
	stats := &programStats{CheckerCount: checkerCount, Files: len(files)}
	perFile := make([]programStatsFileKind, len(files))
	for i, file := range files {
		counts := &perFile[i]
		counts.Files = 1
		counts.Nodes = file.NodeCount
		counts.Bytes = len(file.Text())
		counts.Weight = getCheckerAssociationBaseWeight(file.NodeCount, len(file.Text()))
		counts.Imports = len(file.Imports())
		countProgramStatsNodes(file.AsNode(), counts)
		if file.IsDeclarationFile {
			stats.Declaration.add(counts)
		} else {
			stats.Source.add(counts)
		}
	}
	total := stats.Declaration.Weight + stats.Source.Weight
	if total > 0 {
		stats.DeclarationShare = float64(stats.Declaration.Weight) / float64(total)
	}

	imports := programStatsImports(program)
	importers := make([]int, len(files))
	for i, file := range files {
		if file.IsDeclarationFile {
			continue
		}
		for _, index := range imports[i] {
			if files[index].IsDeclarationFile {
				importers[index]++
			}
		}
	}
	var imported []programStatsImportedFile
	for i, file := range files {
		if !file.IsDeclarationFile || importers[i] == 0 {
			continue
		}
		demand := float64(importers[i]) / float64(max(stats.Source.Files, 1))
		stats.DemandWeight += float64(perFile[i].Weight) * demand
		stats.DemandGeneric += float64(perFile[i].generic()) * demand
		imported = append(imported, programStatsImportedFile{File: file.FileName(), Weight: perFile[i].Weight, Importers: importers[i], Demand: demand, Generic: perFile[i].generic()})
	}
	slices.SortStableFunc(imported, func(a, b programStatsImportedFile) int { return b.Importers - a.Importers })
	stats.TopImported = imported[:min(len(imported), programStatsTopImported)]

	stats.Checkers = make([]programStatsChecker, checkerCount)
	reached := make([]bool, len(files))
	var queue []int
	for c := range stats.Checkers {
		entry := &stats.Checkers[c]
		entry.Checker = c
		clear(reached)
		queue = queue[:0]
		for i, file := range files {
			if file.IsDeclarationFile || associations[i] != c {
				continue
			}
			entry.SourceFiles++
			entry.SourceWeight += perFile[i].Weight
			entry.SourceGeneric += perFile[i].generic()
			reached[i] = true
			queue = append(queue, i)
		}
		for len(queue) > 0 {
			index := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			for _, next := range imports[index] {
				if reached[next] {
					continue
				}
				reached[next] = true
				queue = append(queue, next)
				if files[next].IsDeclarationFile {
					entry.ReachableDeclarationFiles++
					entry.ReachableDeclarationWeight += perFile[next].Weight
					entry.ReachableDeclarationGeneric += perFile[next].generic()
				}
			}
		}
	}
	return stats
}

// programStatsImports returns, per file index, the indices of the program files it imports (each once).
func programStatsImports(program *Program) [][]int {
	files := program.files
	fileIndices := make(map[*ast.SourceFile]int, len(files))
	for i, file := range files {
		fileIndices[file] = i
	}
	imports := make([][]int, len(files))
	seen := make(map[int]struct{})
	for i, file := range files {
		clear(seen)
		for _, resolved := range program.resolvedModules[file.Path()] {
			if resolved == nil || !resolved.IsResolved() {
				continue
			}
			index, ok := fileIndices[program.GetSourceFileForResolvedModule(resolved.ResolvedFileName)]
			if !ok || index == i {
				continue
			}
			if _, done := seen[index]; done {
				continue
			}
			seen[index] = struct{}{}
			imports[i] = append(imports[i], index)
		}
	}
	return imports
}

func (k *programStatsFileKind) add(other *programStatsFileKind) {
	k.Files += other.Files
	k.Nodes += other.Nodes
	k.Bytes += other.Bytes
	k.Weight += other.Weight
	k.Imports += other.Imports
	k.TypeParameters += other.TypeParameters
	k.ConditionalTypes += other.ConditionalTypes
	k.MappedTypes += other.MappedTypes
	k.InferTypes += other.InferTypes
	k.IndexedAccessTypes += other.IndexedAccessTypes
	k.TemplateLiteralTypes += other.TemplateLiteralTypes
	k.TypeOperators += other.TypeOperators
	k.TypeQueries += other.TypeQueries
	k.TypeReferences += other.TypeReferences
	k.GenericReferences += other.GenericReferences
	k.Signatures += other.Signatures
	k.GenericSignatures += other.GenericSignatures
	k.Interfaces += other.Interfaces
	k.TypeAliases += other.TypeAliases
	k.Classes += other.Classes
}

func countProgramStatsNodes(root *ast.Node, counts *programStatsFileKind) {
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindTypeParameter:
			counts.TypeParameters++
		case ast.KindConditionalType:
			counts.ConditionalTypes++
		case ast.KindMappedType:
			counts.MappedTypes++
		case ast.KindInferType:
			counts.InferTypes++
		case ast.KindIndexedAccessType:
			counts.IndexedAccessTypes++
		case ast.KindTemplateLiteralType:
			counts.TemplateLiteralTypes++
		case ast.KindTypeOperator:
			counts.TypeOperators++
		case ast.KindTypeQuery:
			counts.TypeQueries++
		case ast.KindTypeReference, ast.KindExpressionWithTypeArguments:
			counts.TypeReferences++
			if node.TypeArgumentList() != nil {
				counts.GenericReferences++
			}
		case ast.KindCallSignature, ast.KindConstructSignature, ast.KindMethodSignature, ast.KindFunctionType, ast.KindConstructorType,
			ast.KindFunctionDeclaration, ast.KindMethodDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
			counts.Signatures++
			if node.FunctionLikeData().TypeParameters != nil {
				counts.GenericSignatures++
			}
		case ast.KindInterfaceDeclaration:
			counts.Interfaces++
		case ast.KindTypeAliasDeclaration:
			counts.TypeAliases++
		case ast.KindClassDeclaration:
			counts.Classes++
		}
		node.ForEachChild(visit)
		return false
	}
	root.ForEachChild(visit)
}
