package checker

import "testing"

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
