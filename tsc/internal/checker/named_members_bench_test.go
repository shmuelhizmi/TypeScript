package checker

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
)

func BenchmarkGetNamedMembers(b *testing.B) {
	for _, size := range []int{8, 128, 1024} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
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
