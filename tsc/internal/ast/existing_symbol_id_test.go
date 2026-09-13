package ast

import "testing"

func TestTryGetSymbolId(t *testing.T) {
	t.Parallel()
	symbol := &Symbol{}
	if id := TryGetSymbolId(symbol); id != 0 {
		t.Fatalf("unassigned ID = %d, want zero", id)
	}
	id := GetSymbolId(symbol)
	if id == 0 || TryGetSymbolId(symbol) != id || GetSymbolId(symbol) != id {
		t.Fatal("assigned symbol ID must remain stable")
	}
}
