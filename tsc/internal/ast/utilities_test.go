package ast_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"gotest.tools/v3/assert"
)

func TestTryGetSymbolId(t *testing.T) {
	t.Parallel()
	symbol := &ast.Symbol{}
	assert.Equal(t, ast.TryGetSymbolId(symbol), ast.SymbolId(0))
	assert.Equal(t, ast.TryGetSymbolId(symbol), ast.SymbolId(0))
	id := ast.GetSymbolId(symbol)
	assert.Assert(t, id != 0)
	assert.Equal(t, ast.TryGetSymbolId(symbol), id)
	assert.Equal(t, ast.GetSymbolId(symbol), id)
}
