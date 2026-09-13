package checker

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"gotest.tools/v3/assert"
)

func TestMergedSymbolStore(t *testing.T) {
	t.Parallel()
	first := &Checker{}
	second := &Checker{}
	source := &ast.Symbol{Name: "source"}
	target := &ast.Symbol{Name: "target"}
	replacement := &ast.Symbol{Name: "replacement"}
	unmerged := &ast.Symbol{Name: "unmerged"}

	assert.Equal(t, first.getMergedSymbol(nil), (*ast.Symbol)(nil))
	// A symbol without an id cannot have been recorded, and looking it up assigns none.
	assert.Equal(t, first.getMergedSymbol(source), source)
	assert.Equal(t, ast.TryGetSymbolId(source), ast.SymbolId(0))

	first.recordMergedSymbol(target, source)
	assert.Equal(t, first.getMergedSymbol(source), target)
	assert.Equal(t, second.getMergedSymbol(source), source)
	// A symbol with an id but no recorded merge maps to itself, whether its id falls into a page
	// the store has allocated or beyond the last one.
	ast.GetSymbolId(unmerged)
	assert.Equal(t, first.getMergedSymbol(unmerged), unmerged)
	for range 512 {
		ast.GetSymbolId(&ast.Symbol{})
	}
	later := &ast.Symbol{Name: "later"}
	ast.GetSymbolId(later)
	assert.Equal(t, first.getMergedSymbol(later), later)

	first.recordMergedSymbol(replacement, source)
	assert.Equal(t, first.getMergedSymbol(source), replacement)
	second.recordMergedSymbol(source, target)
	assert.Equal(t, second.getMergedSymbol(target), source)
	assert.Equal(t, first.getMergedSymbol(target), target)
}
