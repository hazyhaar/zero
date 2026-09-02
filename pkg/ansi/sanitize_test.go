package ansi

import "testing"

func TestSanitizeFileLine(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "tabs expansion",
			in:   "\tdef foo():\n\t\treturn 1",
			want: "    def foo():\n        return 1",
		},
		{
			name: "control byte strip",
			in:   "line\x00\x07with\x1b[31mbad",
			want: "linewith[31mbad",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeFileLine(tt.in)
			if got != tt.want {
				t.Errorf("SanitizeFileLine(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}
