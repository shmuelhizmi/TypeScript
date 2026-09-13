package printer

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/internal/core"
	"gotest.tools/v3/assert"
)

func TestTextWriterGetColumn(t *testing.T) {
	t.Parallel()
	type step struct {
		op   string // "write", "line", "indent" or "clear"
		text string
		want core.UTF16Offset
	}
	tests := []struct {
		name  string
		steps []step
	}{
		{"ascii", []step{{"write", "abc", 3}, {"write", "def", 6}}},
		{"two-byte rune", []step{{"write", "ab", 2}, {"write", "é", 3}, {"write", "c", 4}}},
		{"surrogate pair", []step{{"write", "😀", 2}, {"write", "x", 3}}},
		{"non-ascii then ascii", []step{{"write", "é", 1}, {"write", "abc", 4}}},
		{"utf-8 sequence split across writes", []step{{"write", "a\xc3", 2}, {"write", "\xa9b", 3}}},
		{"newline inside a write", []step{{"write", "ab\ncd", 2}, {"write", "e", 3}}},
		{"crlf inside a write", []step{{"write", "ab\r\ncd", 2}}},
		{"line separator inside a write", []step{{"write", "ab\u2028cd", 2}}},
		{"non-ascii before a newline", []step{{"write", "é\nab", 2}, {"write", "c", 3}}},
		{"write line", []step{{"write", "abc", 3}, {"line", "", 0}, {"write", "d", 1}}},
		{"indented line", []step{{"write", "x", 1}, {"indent", "", 1}, {"line", "", 4}, {"write", "a", 5}}},
		{"clear", []step{{"write", "abc", 3}, {"clear", "", 0}, {"write", "de", 2}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			w := NewTextWriter("\n", 4).(*textWriter)
			for i, step := range test.steps {
				switch step.op {
				case "write":
					w.Write(step.text)
				case "line":
					w.WriteLine()
				case "indent":
					w.IncreaseIndent()
				case "clear":
					w.Clear()
				}
				got := w.GetColumn()
				assert.Equal(t, got, step.want, "step %d", i)
				if !w.IsAtStartOfLine() {
					assert.Equal(t, got, core.UTF16Len(w.String()[w.linePos:]), "step %d", i)
				}
				// Asking again does not change the answer.
				assert.Equal(t, w.GetColumn(), got, "step %d", i)
			}
		})
	}
}
