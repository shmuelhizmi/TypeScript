package checker

import (
	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"gotest.tools/v3/assert"
	"testing"
)

func TestMergedSymbolStore(t *testing.T) {
	t.Parallel()
	first := &Checker{}
	second := &Checker{}
	source := &ast.Symbol{Name: "source"}
	target := &ast.Symbol{Name: "target"}
	replacement := &ast.Symbol{Name: "replacement"}
	adjacent := &ast.Symbol{Name: "adjacent"}
	assert.Equal(t, first.getMergedSymbol(nil), (*ast.Symbol)(nil))
	assert.Equal(t, first.getMergedSymbol(source), source)
	first.recordMergedSymbol(target, source)
	assert.Equal(t, first.getMergedSymbol(source), target)
	assert.Equal(t, first.getMergedSymbol(adjacent), adjacent)
	assert.Equal(t, second.getMergedSymbol(source), source)
	first.recordMergedSymbol(replacement, source)
	assert.Equal(t, first.getMergedSymbol(source), replacement)
	second.recordMergedSymbol(source, target)
	assert.Equal(t, second.getMergedSymbol(target), source)
	assert.Equal(t, first.getMergedSymbol(target), target)
}
