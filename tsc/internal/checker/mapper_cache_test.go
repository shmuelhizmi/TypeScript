package checker

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"gotest.tools/v3/assert"
)

func TestInstantiateTypeWithActiveMapperCaches(t *testing.T) {
	t.Parallel()
	c := &Checker{}
	c.couldContainTypeVariables = c.couldContainTypeVariablesWorker
	source := &Type{flags: TypeFlagsTypeParameter, id: 1}
	first := &Type{flags: TypeFlagsString, id: 2}
	second := &Type{flags: TypeFlagsNumber, id: 3}
	outer := newTypeMapper([]*Type{source}, []*Type{first})
	inner := newTypeMapper([]*Type{source}, []*Type{second})
	alias := &TypeAlias{symbol: &ast.Symbol{Name: "Alias"}}
	// The second round finds the cache maps that the first round left in the slots.
	for round := range 2 {
		// A mapper that is not active is pushed for the instantiation only, and nothing is cached for it.
		assert.Equal(t, c.instantiateType(source, outer), first)
		assert.Equal(t, len(c.activeMappers), 0)
		assert.Equal(t, len(c.activeTypeMappersCaches), 0)
		assert.Equal(t, len(c.activeTypeMappersTypeIdCaches), 0)

		c.pushActiveMapper(outer)
		assert.Equal(t, len(c.activeTypeMappersCaches), 1)
		assert.Equal(t, len(c.activeTypeMappersTypeIdCaches), 1)
		assert.Equal(t, c.activeTypeMappersCaches[0] == nil, round == 0, "round %d", round)
		assert.Equal(t, c.activeTypeMappersTypeIdCaches[0] == nil, round == 0, "round %d", round)

		// The first alias-free instantiation under an active mapper creates the type id cache and records the result.
		count := c.instantiationCount
		assert.Equal(t, c.instantiateType(source, outer), first)
		assert.Equal(t, c.instantiationCount, count+1)
		assert.Equal(t, len(c.activeTypeMappersTypeIdCaches[0]), 1)
		assert.Equal(t, len(c.activeTypeMappersCaches[0]), 0)
		// A repeated instantiation is served from the cache.
		assert.Equal(t, c.instantiateType(source, outer), first)
		assert.Equal(t, c.instantiationCount, count+1)
		// An aliased instantiation goes through the hashed cache.
		assert.Equal(t, c.instantiateTypeWithAlias(source, outer, alias), first)
		assert.Equal(t, c.instantiationCount, count+2)
		assert.Equal(t, len(c.activeTypeMappersCaches[0]), 1)
		assert.Equal(t, c.instantiateTypeWithAlias(source, outer, alias), first)
		assert.Equal(t, c.instantiationCount, count+2)
		// A nested mapper that is not active is not cached under the active one.
		assert.Equal(t, c.instantiateType(source, inner), second)
		assert.Equal(t, c.instantiationCount, count+3)
		assert.Equal(t, len(c.activeTypeMappersTypeIdCaches[0]), 1)
		assert.Equal(t, len(c.activeTypeMappersCaches[0]), 1)
		assert.Equal(t, len(c.activeMappers), 1)
		// Clearing drops the entries and keeps the maps.
		c.clearActiveMapperCaches()
		assert.Equal(t, len(c.activeTypeMappersTypeIdCaches[0]), 0)
		assert.Equal(t, len(c.activeTypeMappersCaches[0]), 0)
		assert.Equal(t, c.instantiateType(source, outer), first)
		assert.Equal(t, c.instantiationCount, count+4)
		assert.Equal(t, len(c.activeTypeMappersTypeIdCaches[0]), 1)

		c.popActiveMapper()
		assert.Equal(t, len(c.activeMappers), 0)
		assert.Equal(t, len(c.activeTypeMappersCaches), 0)
		assert.Equal(t, len(c.activeTypeMappersTypeIdCaches), 0)
		assert.Equal(t, c.instantiateType(source, inner), second)
	}
}
