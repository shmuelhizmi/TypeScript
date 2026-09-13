package checker

import (
	"strconv"
	"testing"
)

func BenchmarkGetConditionalTypeInstantiationCacheHit(b *testing.B) {
	for _, size := range []int{1, 2, 4} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			c := &Checker{}
			parameters := make([]*Type, size)
			arguments := make([]*Type, size)
			for i := range size {
				parameters[i] = &Type{flags: TypeFlagsTypeParameter, id: TypeId(i + 1)}
				arguments[i] = &Type{flags: TypeFlagsString, id: TypeId(i + size + 1)}
			}
			result := &Type{flags: TypeFlagsNumber, id: TypeId(size*2 + 1)}
			root := &ConditionalRoot{outerTypeParameters: parameters, instantiations: map[CacheHashKey]*Type{getConditionalTypeKey(arguments, nil, false): result}}
			conditional := &Type{flags: TypeFlagsConditional, data: &ConditionalType{root: root}}
			mapper := newTypeMapper(parameters, arguments)
			b.ReportAllocs()
			for b.Loop() {
				c.getConditionalTypeInstantiation(conditional, mapper, false, nil)
			}
		})
	}
}
