package checker

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"gotest.tools/v3/assert"
)

func TestGetNamedMembers(t *testing.T) {
	t.Parallel()
	c := &Checker{}
	c.compareSymbols = c.compareSymbolsWorker
	container := &ast.Symbol{Flags: ast.SymbolFlagsInterface, Declarations: []*ast.Node{{Loc: core.NewTextRange(100, 200)}}}
	declared := &ast.Symbol{Name: "zDeclared", Flags: ast.SymbolFlagsProperty, ValueDeclaration: &ast.Node{Loc: core.NewTextRange(120, 130)}}
	inherited := &ast.Symbol{Name: "aInherited", Flags: ast.SymbolFlagsProperty, ValueDeclaration: &ast.Node{Loc: core.NewTextRange(20, 30)}}
	reserved := &ast.Symbol{Name: "reserved", Flags: ast.SymbolFlagsProperty}
	typeOnly := &ast.Symbol{Name: "typeOnly", Flags: ast.SymbolFlagsInterface}
	members := ast.SymbolTable{"zDeclared": declared, "aInherited": inherited, "\xFEcall": reserved, "typeOnly": typeOnly}
	// Map iteration order changes from run to run; repeating covers both visiting orders of the two named members.
	for range 20 {
		assert.DeepEqual(t, symbolNames(c.getNamedMembers(members, container)), []string{"zDeclared", "aInherited"})
		assert.DeepEqual(t, symbolNames(c.getNamedMembers(members, nil)), []string{"aInherited", "zDeclared"})
		assert.DeepEqual(t, symbolNames(c.getNamedMembers(ast.SymbolTable{"zDeclared": declared}, container)), []string{"zDeclared"})
		assert.DeepEqual(t, symbolNames(c.getNamedMembers(ast.SymbolTable{"aInherited": inherited}, container)), []string{"aInherited"})
		assert.Equal(t, len(c.getNamedMembers(ast.SymbolTable{"typeOnly": typeOnly}, container)), 0)
		assert.Equal(t, len(c.getNamedMembers(nil, container)), 0)
	}
}

// symbolNames returns the names of symbols in order, for readable assertions.
func symbolNames(symbols []*ast.Symbol) []string {
	names := make([]string, len(symbols))
	for i, symbol := range symbols {
		names[i] = symbol.Name
	}
	return names
}
