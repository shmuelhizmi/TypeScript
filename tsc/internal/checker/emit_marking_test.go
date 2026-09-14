package checker_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/bundled"
	"github.com/microsoft/TypeScript/tsc/internal/checker"
	"github.com/microsoft/TypeScript/tsc/internal/compiler"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/tsoptions"
	"github.com/microsoft/TypeScript/tsc/internal/vfs/vfstest"
	"gotest.tools/v3/assert"
)

// One import of each kind the emit resolver is asked about: a name used as a value, a name used only as a type,
// a name not used at all, and a name re-exported, which the recursive walk marks through the export rather than
// through a reference in the file.
const emitMarkingSource = `
import { usedAsValue, usedAsType, notUsed, passedThrough } from "./exports";

export { passedThrough };

export const value = usedAsValue;
export declare const typed: usedAsType;
`

func newEmitMarkingProgram(t *testing.T) *compiler.Program {
	t.Helper()
	fs := bundled.WrapFS(vfstest.FromMap(map[string]string{
		"/main.ts": emitMarkingSource,
		"/exports.ts": `export const usedAsValue = 1;
export type usedAsType = string;
export const notUsed = 2;
export const passedThrough = 3;
`,
		"/tsconfig.json": `{ "compilerOptions": { "strict": true }, "files": ["main.ts"] }`,
	}, false /*useCaseSensitiveFileNames*/))
	host := compiler.NewCompilerHost("/", fs, bundled.LibPath(), nil, nil, nil)
	parsed, errors := tsoptions.GetParsedCommandLineOfConfigFile("/tsconfig.json", &core.CompilerOptions{}, nil, host, nil)
	assert.Equal(t, len(errors), 0, "expected no errors in parsed command line")
	program := compiler.NewProgram(compiler.ProgramOptions{Config: parsed, Host: host})
	program.BindSourceFiles()
	return program
}

// aliasDeclarationsByName returns every import and export specifier in the file, keyed by the name it binds.
func aliasDeclarationsByName(file *ast.SourceFile) map[string]*ast.Node {
	found := map[string]*ast.Node{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsImportSpecifier(node) || ast.IsExportSpecifier(node) {
			found[node.Name().Text()] = node
		}
		node.ForEachChild(visit)
		return false
	}
	file.AsNode().ForEachChild(visit)
	return found
}

func referencedAliases(resolver *checker.EmitResolver, aliases map[string]*ast.Node) map[string]bool {
	referenced := make(map[string]bool, len(aliases))
	for name, declaration := range aliases {
		referenced[name] = resolver.IsReferencedAliasDeclaration(declaration)
	}
	return referenced
}

// TestMarkLinkedReferencesMatchesChecking pins the invariant the skip rests on: on a file the checker has
// checked, the recursive walk marks nothing the check did not mark already. The two programs reach the same
// state from opposite directions -- one checks and then marks, the other marks and then checks -- so a
// reference the walk finds and checking does not would show up as a difference between them.
func TestMarkLinkedReferencesMatchesChecking(t *testing.T) {
	t.Parallel()

	checkedFirst := newEmitMarkingProgram(t)
	checkedFile := checkedFirst.GetSourceFile("/main.ts")
	checkedChecker, doneChecked := checkedFirst.GetTypeCheckerForFile(t.Context(), checkedFile)
	defer doneChecked()
	checkedChecker.GetDiagnostics(t.Context(), checkedFile)
	checkedChecker.GetEmitResolver().MarkLinkedReferencesRecursively(checkedFile)
	checked := referencedAliases(checkedChecker.GetEmitResolver(), aliasDeclarationsByName(checkedFile))

	markedFirst := newEmitMarkingProgram(t)
	markedFile := markedFirst.GetSourceFile("/main.ts")
	markedChecker, doneMarked := markedFirst.GetTypeCheckerForFile(t.Context(), markedFile)
	defer doneMarked()
	markedAliases := aliasDeclarationsByName(markedFile)
	markedChecker.GetEmitResolver().MarkLinkedReferencesRecursively(markedFile)

	// The walk runs on a file that has not been checked, so it is what marks the value reference here.
	assert.Assert(t, markedChecker.GetEmitResolver().IsReferencedAliasDeclaration(markedAliases["usedAsValue"]),
		"the walk marks the alias a value in the file refers to")

	markedChecker.GetDiagnostics(t.Context(), markedFile)
	marked := referencedAliases(markedChecker.GetEmitResolver(), markedAliases)

	assert.DeepEqual(t, checked, marked)

	// Both orders have to leave something on either side of the line, or they agree about nothing.
	assert.Assert(t, checked["usedAsValue"], "an alias a value reference reaches is referenced")
	assert.Assert(t, !checked["notUsed"], "an alias nothing reaches is not referenced")
	assert.Assert(t, !checked["usedAsType"], "an alias only a type position reaches is not referenced")
}
