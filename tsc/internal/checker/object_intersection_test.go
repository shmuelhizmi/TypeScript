package checker_test

import (
	"slices"
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

// The property names are deliberately not in alphabetical order: an intersection lists them in the order its
// constituents declare them, first constituent first, and a test whose expectation is sorted would not see that
// order change. The last two constituents are the kinds whose property list is not what a lookup by name on them
// finds: a module re-exporting a value as a type only, whose list holds a name a lookup answers nil for, and a
// primitive, whose members come from its apparent type rather than from itself.
const objectIntersectionSource = `
interface A { b: string; shared: string; a: string; }
interface B { a2: string; shared: string; c: string; }

declare const a: A;
declare const b: B;
declare const ab: A & B;

declare const withModule: typeof import("./reexport") & { extra: string };

declare const withPrimitive: { own: string } & string;

const readOwn: string = ab.b;
const readShared: string = ab.shared;
`

func newObjectIntersectionProgram(t *testing.T, source string) *compiler.Program {
	t.Helper()
	fs := bundled.WrapFS(vfstest.FromMap(map[string]string{
		"/main.ts":  source,
		"/other.ts": `export const reexportedAsTypeOnly = "";`,
		// The star re-export carries the value through as a type only: the module's property list holds the name,
		// and a lookup for it on the module's type answers nil.
		"/reexport.ts":   "export type * from \"./other\";\nexport const own = \"\";\n",
		"/tsconfig.json": `{ "compilerOptions": { "strict": true }, "files": ["main.ts"] }`,
	}, false /*useCaseSensitiveFileNames*/))
	host := compiler.NewCompilerHost("/", fs, bundled.LibPath(), nil, nil, nil)
	parsed, errors := tsoptions.GetParsedCommandLineOfConfigFile("/tsconfig.json", &core.CompilerOptions{}, nil, host, nil)
	assert.Equal(t, len(errors), 0, "expected no errors in parsed command line")
	program := compiler.NewProgram(compiler.ProgramOptions{Config: parsed, Host: host})
	program.BindSourceFiles()
	return program
}

// objectIntersectionTypeOf returns the type of a variable in the fixture, by the name it is declared with.
func objectIntersectionTypeOf(t *testing.T, c *checker.Checker, file *ast.SourceFile, name string) *checker.Type {
	t.Helper()
	for _, statement := range file.Statements.Nodes {
		if !ast.IsVariableStatement(statement) {
			continue
		}
		for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			if declaration.Name().Text() == name {
				return c.GetTypeAtLocation(declaration.Name())
			}
		}
	}
	t.Fatalf("no variable named %s in the fixture", name)
	return nil
}

func intersectionPropertyNames(symbols []*ast.Symbol) []string {
	names := make([]string, len(symbols))
	for i, symbol := range symbols {
		names[i] = symbol.Name
	}
	return names
}

func intersectionPropertyNamed(symbols []*ast.Symbol, name string) *ast.Symbol {
	for _, symbol := range symbols {
		if symbol.Name == name {
			return symbol
		}
	}
	return nil
}

func TestPropertiesOfObjectIntersection(t *testing.T) {
	t.Parallel()
	program := newObjectIntersectionProgram(t, objectIntersectionSource)
	c, done := program.GetTypeChecker(t.Context())
	defer done()
	file := program.GetSourceFile("/main.ts")

	aType := objectIntersectionTypeOf(t, c, file, "a")
	bType := objectIntersectionTypeOf(t, c, file, "b")
	abType := objectIntersectionTypeOf(t, c, file, "ab")
	properties := c.GetPropertiesOfType(abType)

	// Every name either constituent declares, once, in the order they declare them.
	assert.DeepEqual(t, intersectionPropertyNames(properties), []string{"b", "shared", "a", "a2", "c"})

	// A name a single constituent declares is listed as that constituent's own symbol, which is the symbol a
	// by-name lookup on the intersection returns for it as well.
	for _, single := range []struct {
		name        string
		constituent *checker.Type
	}{{"b", aType}, {"a", aType}, {"a2", bType}, {"c", bType}} {
		listed := intersectionPropertyNamed(properties, single.name)
		assert.Equal(t, listed, c.GetPropertyOfType(single.constituent, single.name))
		assert.Equal(t, listed, c.GetPropertyOfType(abType, single.name))
	}

	// A name two constituents declare is synthesized from both, as it was before.
	shared := intersectionPropertyNamed(properties, "shared")
	assert.Assert(t, shared.CheckFlags&ast.CheckFlagsSynthetic != 0)
	assert.Assert(t, shared != c.GetPropertyOfType(aType, "shared"))
	assert.Assert(t, shared != c.GetPropertyOfType(bType, "shared"))
	assert.Equal(t, c.TypeToString(c.GetTypeOfSymbol(shared)), "string")

	// A name the intersection does not have still misses, whatever the listing put in the cache.
	assert.Assert(t, c.GetPropertyOfType(abType, "neverDeclared") == nil)
}

func TestPropertiesOfIntersectionProbedByName(t *testing.T) {
	t.Parallel()
	program := newObjectIntersectionProgram(t, objectIntersectionSource)
	c, done := program.GetTypeChecker(t.Context())
	defer done()
	file := program.GetSourceFile("/main.ts")

	// A module's property list can hold a name a lookup on the same type answers nil for, so an intersection
	// with a module constituent is enumerated by probing every constituent: the name is left out, as it is
	// without this listing.
	withModule := intersectionPropertyNames(c.GetPropertiesOfType(objectIntersectionTypeOf(t, c, file, "withModule")))
	assert.DeepEqual(t, withModule, []string{"own", "extra"})
	assert.Assert(t, !slices.Contains(withModule, "reexportedAsTypeOnly"), "listed: %v", withModule)

	// A primitive constituent contributes the members of its apparent type, which its own property list does
	// not hold at all.
	withPrimitive := intersectionPropertyNames(c.GetPropertiesOfType(objectIntersectionTypeOf(t, c, file, "withPrimitive")))
	assert.Assert(t, slices.Contains(withPrimitive, "own"), "declared property: %v", withPrimitive)
	assert.Assert(t, slices.Contains(withPrimitive, "charAt"), "apparent members of the string constituent: %v", withPrimitive)
}

func TestObjectIntersectionChecksClean(t *testing.T) {
	t.Parallel()
	// The two reads at the end of the fixture go through the property cache, one for a name a single
	// constituent declares and one for a name both declare.
	program := newObjectIntersectionProgram(t, objectIntersectionSource)
	diagnostics := program.GetSemanticDiagnostics(t.Context(), program.GetSourceFile("/main.ts"))
	for _, diagnostic := range diagnostics {
		t.Errorf("unexpected diagnostic: %s", diagnostic.MessageText())
	}
}
