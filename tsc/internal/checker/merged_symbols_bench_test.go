package checker

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
)

func BenchmarkGetMergedSymbol(b *testing.B) {
	c := &Checker{}
	symbols := make([]*ast.Symbol, 1024)
	for i := range symbols {
		symbols[i] = &ast.Symbol{Name: "symbol"}
		ast.GetSymbolId(symbols[i])
		if i%8 == 0 {
			c.recordMergedSymbol(&ast.Symbol{Name: "merged"}, symbols[i])
		}
	}
	b.ReportAllocs()
	for b.Loop() {
		for _, symbol := range symbols {
			c.getMergedSymbol(symbol)
		}
	}
}
