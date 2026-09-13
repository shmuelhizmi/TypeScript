package checker

import (
	"fmt"
	"slices"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"gotest.tools/v3/assert"
)

func TestNamedMembersPartition(t *testing.T) {
	t.Parallel()
	c := &Checker{}
	c.compareSymbols = c.compareSymbolsWorker
	container := &ast.Symbol{Flags: ast.SymbolFlagsInterface, Declarations: []*ast.Node{{Loc: core.NewTextRange(100, 200)}}}
	declared := &ast.Symbol{Name: "zDeclared", Flags: ast.SymbolFlagsProperty, ValueDeclaration: &ast.Node{Loc: core.NewTextRange(120, 130)}}
	inherited := &ast.Symbol{Name: "aInherited", Flags: ast.SymbolFlagsProperty, ValueDeclaration: &ast.Node{Loc: core.NewTextRange(20, 30)}}
	reserved := &ast.Symbol{Name: "reserved", Flags: ast.SymbolFlagsProperty}
	typeOnly := &ast.Symbol{Name: "typeOnly", Flags: ast.SymbolFlagsInterface}
	members := ast.SymbolTable{"zDeclared": declared, "aInherited": inherited, "\xFEcall": reserved, "typeOnly": typeOnly}
	for range 20 {
		assert.Assert(t, slices.Equal(c.getNamedMembers(members, container), []*ast.Symbol{declared, inherited}))
		assert.Assert(t, slices.Equal(c.getNamedMembers(members, nil), []*ast.Symbol{inherited, declared}))
		assert.Assert(t, slices.Equal(c.getNamedMembers(ast.SymbolTable{"zDeclared": declared}, container), []*ast.Symbol{declared}))
		assert.Assert(t, slices.Equal(c.getNamedMembers(ast.SymbolTable{"aInherited": inherited}, container), []*ast.Symbol{inherited}))
		assert.Equal(t, len(c.getNamedMembers(ast.SymbolTable{"typeOnly": typeOnly}, container)), 0)
		assert.Equal(t, len(c.getNamedMembers(nil, container)), 0)
	}
}

func BenchmarkNamedMembers(b *testing.B) {
	for _, size := range []int{8, 128, 1024} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			c := &Checker{}
			c.compareSymbols = c.compareSymbolsWorker
			container := &ast.Symbol{Flags: ast.SymbolFlagsInterface, Declarations: []*ast.Node{{Loc: core.NewTextRange(100, 200)}}}
			members := make(ast.SymbolTable, size)
			for i := range size {
				pos := 20
				if i%3 == 0 {
					pos = 120
				}
				name := fmt.Sprintf("property%04d", i)
				members[name] = &ast.Symbol{Name: name, Flags: ast.SymbolFlagsProperty, ValueDeclaration: &ast.Node{Loc: core.NewTextRange(pos, pos+1)}}
			}
			b.ReportAllocs()
			for b.Loop() {
				c.getNamedMembers(members, container)
			}
		})
	}
}
