package printer

import (
	"strconv"
	"testing"
)

func BenchmarkTextWriterGetColumn(b *testing.B) {
	for _, width := range []int{80, 2000} {
		b.Run(strconv.Itoa(width), func(b *testing.B) {
			w := NewTextWriter("\n", 4)
			b.ReportAllocs()
			for b.Loop() {
				w.Clear()
				for range width / 8 {
					w.Write("abcdefg ")
					w.GetColumn()
				}
			}
		})
	}
}
