package checker

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"gotest.tools/v3/assert"
)

func TestGetFlowReferenceRootName(t *testing.T) {
	t.Parallel()
	f := &ast.NodeFactory{}
	a := func() *ast.Node { return f.NewIdentifier("a") }
	thisInTypeQuery := f.NewIdentifier("this")
	ast.SetParentInChildren(f.NewTypeQueryNode(f.NewQualifiedName(thisInTypeQuery, f.NewIdentifier("x")), nil))
	tests := []struct {
		name      string
		reference *ast.Node
		want      string
	}{
		{"identifier", a(), "a"},
		{"property access", f.NewPropertyAccessExpression(a(), nil, f.NewIdentifier("b"), ast.NodeFlagsNone), "a"},
		{"element access", f.NewElementAccessExpression(a(), nil, f.NewNumericLiteral("0", ast.TokenFlagsNone), ast.NodeFlagsNone), "a"},
		{"parenthesized non-null access", f.NewPropertyAccessExpression(f.NewParenthesizedExpression(f.NewNonNullExpression(a(), ast.NodeFlagsNone)), nil, f.NewIdentifier("b"), ast.NodeFlagsNone), "a"},
		{"satisfies", f.NewSatisfiesExpression(a(), f.NewKeywordTypeNode(ast.KindStringKeyword)), "a"},
		{"this property", f.NewPropertyAccessExpression(f.NewKeywordExpression(ast.KindThisKeyword), nil, f.NewIdentifier("x"), ast.NodeFlagsNone), ""},
		{"this in type query", thisInTypeQuery, ""},
		{"meta property", f.NewMetaProperty(ast.KindNewKeyword, f.NewIdentifier("target")), ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, getFlowReferenceRootName(test.reference), test.want)
		})
	}
}
