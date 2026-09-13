package checker

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestInstantiateTypeWithNestedMappers(t *testing.T) {
	t.Parallel()
	c := &Checker{}
	c.couldContainTypeVariables = c.couldContainTypeVariablesWorker
	source := &Type{flags: TypeFlagsTypeParameter, id: 1}
	first := &Type{flags: TypeFlagsString, id: 2}
	second := &Type{flags: TypeFlagsNumber, id: 3}
	outer := newTypeMapper([]*Type{source}, []*Type{first})
	inner := newTypeMapper([]*Type{source}, []*Type{second})
	for range 2 {
		assert.Equal(t, c.instantiateType(source, outer), first)
		c.pushActiveMapper(outer)
		assert.Equal(t, c.instantiateType(source, outer), first)
		assert.Equal(t, c.instantiateType(source, inner), second)
		assert.Equal(t, c.instantiateType(source, outer), first)
		c.clearActiveMapperCaches()
		assert.Equal(t, c.instantiateType(source, outer), first)
		c.popActiveMapper()
		assert.Equal(t, c.instantiateType(source, inner), second)
	}
}

func BenchmarkInstantiateSimpleTypeParameter(b *testing.B) {
	for _, active := range []bool{false, true} {
		name := "new mapper"
		if active {
			name = "active mapper"
		}
		b.Run(name, func(b *testing.B) {
			c := &Checker{}
			c.couldContainTypeVariables = c.couldContainTypeVariablesWorker
			source := &Type{flags: TypeFlagsTypeParameter, id: 1}
			target := &Type{flags: TypeFlagsString, id: 2}
			mapper := newTypeMapper([]*Type{source}, []*Type{target})
			if active {
				c.pushActiveMapper(mapper)
			}
			b.ReportAllocs()
			for b.Loop() {
				c.instantiationCount = 0
				c.instantiateType(source, mapper)
			}
		})
	}
}
