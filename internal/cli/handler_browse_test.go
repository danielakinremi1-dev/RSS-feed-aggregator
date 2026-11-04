package cli

import "testing"

func TestPlainText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		max  int
		want string
	}{
		{"plain", "hello world", 50, "hello world"},
		{"tags dropped", `<a href="https://x.test/1">Comments</a>`, 50, "Comments"},
		{"tags separate words", "<p>one</p><p>two</p>", 50, "one two"},
		{"entities decoded", "Fish &amp; chips &lt;3", 50, "Fish & chips <3"},
		{"whitespace collapsed", "  a\n\n\tb   c ", 50, "a b c"},
		{"truncated with ellipsis", "abcdefghij", 5, "abcd…"},
		{"exactly max is kept", "abcde", 5, "abcde"},
		{"multibyte counted as runes", "héllo wörld", 6, "héllo…"},
		{"empty", "", 10, ""},
		{"only tags", "<br/><img src='x'>", 10, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := plainText(tt.in, tt.max); got != tt.want {
				t.Errorf("plainText(%q, %d) = %q, want %q", tt.in, tt.max, got, tt.want)
			}
		})
	}
}
