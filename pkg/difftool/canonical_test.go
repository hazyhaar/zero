package difftool

import (
	"reflect"
	"testing"
)

func TestCanonicalLineKey(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "plain ascii",
			in:   "func Hello() {}",
			want: "func Hello() {}",
		},
		{
			name: "leading and trailing whitespace",
			in:   "   \tvar a = 1\t  ",
			want: "var a = 1",
		},
		{
			name: "internal tabs converted to 4 spaces",
			in:   "var\tField\t= 1",
			want: "var    Field    = 1",
		},
		{
			name: "control characters stripped",
			in:   "hello\x00\x1b[31mworld\x07",
			want: "hello[31mworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanonicalLineKey(tt.in)
			if got != tt.want {
				t.Errorf("CanonicalLineKey(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestMatchChangedLines(t *testing.T) {
	lines := []string{
		"package main",
		"func main() {",
		"\tvar\tValue\t= 42",
		"\tprintln(Value)",
		"}",
	}

	changed := []string{
		"var    Value    = 42",
		"println(Value)",
	}

	got := MatchChangedLines(lines, changed)
	want := map[int]bool{
		2: true,
		3: true,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("MatchChangedLines() = %+v; want %+v", got, want)
	}
}
