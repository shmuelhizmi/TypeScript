package checker

import (
	"testing"

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
	// The second round finds the cache map that the first round left in the slot.
	for round := range 2 {
		// A mapper that is not active is pushed for the instantiation only, and nothing is cached for it.
		assert.Equal(t, c.instantiateType(source, outer), first)
		assert.Equal(t, len(c.activeMappers), 0)
		assert.Equal(t, len(c.activeTypeMappersCaches), 0)

		c.pushActiveMapper(outer)
		assert.Equal(t, len(c.activeTypeMappersCaches), 1)
		assert.Equal(t, c.activeTypeMappersCaches[0] == nil, round == 0, "round %d", round)

		// The first instantiation under an active mapper creates the cache and records the result.
		count := c.instantiationCount
		assert.Equal(t, c.instantiateType(source, outer), first)
		assert.Equal(t, c.instantiationCount, count+1)
		assert.Equal(t, len(c.activeTypeMappersCaches[0]), 1)
		assert.Equal(t, c.instantiateType(source, outer), first)
		assert.Equal(t, c.instantiationCount, count+1)
		// A nested mapper that is not active is not cached under the active one.
		assert.Equal(t, c.instantiateType(source, inner), second)
		assert.Equal(t, c.instantiationCount, count+2)
		assert.Equal(t, len(c.activeTypeMappersCaches[0]), 1)
		assert.Equal(t, len(c.activeMappers), 1)
		// Clearing drops the entries and keeps the map.
		c.clearActiveMapperCaches()
		assert.Equal(t, len(c.activeTypeMappersCaches[0]), 0)
		assert.Equal(t, c.instantiateType(source, outer), first)
		assert.Equal(t, c.instantiationCount, count+3)
		assert.Equal(t, len(c.activeTypeMappersCaches[0]), 1)

		c.popActiveMapper()
		assert.Equal(t, len(c.activeMappers), 0)
		assert.Equal(t, len(c.activeTypeMappersCaches), 0)
		assert.Equal(t, c.instantiateType(source, inner), second)
	}
}
