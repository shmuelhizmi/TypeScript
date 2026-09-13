package checker

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"gotest.tools/v3/assert"
)

func TestGetConditionalTypeInstantiationCacheKey(t *testing.T) {
	t.Parallel()
	c := &Checker{}
	first := &Type{flags: TypeFlagsTypeParameter, id: 1}
	second := &Type{flags: TypeFlagsTypeParameter, id: 2}
	target := &Type{flags: TypeFlagsString, id: 3}
	plain := &Type{flags: TypeFlagsNumber, id: 4}
	aliased := &Type{flags: TypeFlagsBoolean, id: 5}
	constrained := &Type{flags: TypeFlagsNever, id: 6}
	alias := &TypeAlias{symbol: &ast.Symbol{Name: "Alias"}, typeArguments: []*Type{target}}
	// One and two outer type parameters take different paths to the same cache keys.
	for _, parameters := range [][]*Type{{first}, {first, second}} {
		arguments := make([]*Type, len(parameters))
		for i := range arguments {
			arguments[i] = target
		}
		root := &ConditionalRoot{outerTypeParameters: parameters, instantiations: map[CacheHashKey]*Type{
			getConditionalTypeKey(arguments, nil, false):   plain,
			getConditionalTypeKey(arguments, alias, false): aliased,
			getConditionalTypeKey(arguments, nil, true):    constrained,
		}}
		conditional := &Type{flags: TypeFlagsConditional, data: &ConditionalType{root: root}}
		mapper := newTypeMapper(parameters, arguments)
		assert.Equal(t, c.getConditionalTypeInstantiation(conditional, mapper, false, nil), plain)
		assert.Equal(t, c.getConditionalTypeInstantiation(conditional, mapper, false, alias), aliased)
		assert.Equal(t, c.getConditionalTypeInstantiation(conditional, mapper, true, nil), constrained)
	}
}
