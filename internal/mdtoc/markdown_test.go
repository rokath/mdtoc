package mdtoc

import "testing"

// TestExtractPlainTextPreservesLiteralPunctuation guards the semantic text used
// by sluggers: Markdown syntax disappears, but literal identifier characters,
// escaped punctuation, and decoded entities must survive (#123).
func TestExtractPlainTextPreservesLiteralPunctuation(t *testing.T) {
	cases := []struct {
		markup string
		want   string
	}{
		{"Xxxx_Yyyy", "Xxxx_Yyyy"},
		{"foo_bar_baz", "foo_bar_baz"},
		{"Über_Größe_中文", "Über_Größe_中文"},
		{"_prefix suffix_", "prefix suffix"},
		{"_prefix and suffix_", "prefix and suffix"},
		{"_prefix", "_prefix"},
		{"suffix_", "suffix_"},
		{"__COUNTER__", "COUNTER"},
		{"Missing `__COUNTER__`", "Missing __COUNTER__"},
		{`Missing \_\_COUNTER\_\_`, "Missing __COUNTER__"},
		{"Missing &#95;&#95;COUNTER&#95;&#95;", "Missing __COUNTER__"},
		{"Missing &#x5f;&#x5f;COUNTER&#x5f;&#x5f;", "Missing __COUNTER__"},
		{"Missing &lowbar;&lowbar;COUNTER&lowbar;&lowbar;", "Missing __COUNTER__"},
		{"`&#95;` and `\\_`", "&#95; and \\_"},
		{`\&lowbar; and &amp;lowbar;`, "&lowbar; and &lowbar;"},
		{"`` `__COUNTER__` ``", "`__COUNTER__`"},
		{"``a`b``", "a`b"},
		{"` a  b `", "a b"},
		{"``unmatched `tail", "``unmatched `tail"},
		{`\` + "`not code" + `\` + "`", "`not code`"},
		{"**Bold _and italic_** ~~old~~", "Bold and italic old"},
		{"*foo **bar***", "foo bar"},
		{"A * B ~ C", "A * B ~ C"},
		{`C:\path\file`, `C:\path\file`},
		{"[**API** `__LINE__`](https://example.com/a_(b))", "API __LINE__"},
		{"[array `[x]`](target)", "array [x]"},
		{`[a \] b](target)`, "a ] b"},
		{"![alt _text_](image.png)", "alt text"},
	}
	for _, tc := range cases {
		t.Run(tc.markup, func(t *testing.T) {
			got, err := ExtractPlainText(tc.markup)
			if err != nil || got != tc.want {
				t.Fatalf("ExtractPlainText(%q) = %q, %v; want %q", tc.markup, got, err, tc.want)
			}
		})
	}
}

// TestExtractPlainTextPreservingWhitespaceWithCode keeps GitHub's repeated-space
// slug behavior while applying code-span padding and entity decoding only once.
func TestExtractPlainTextPreservingWhitespaceWithCode(t *testing.T) {
	got, err := ExtractPlainTextPreservingWhitespace("Missing  ` __COUNTER__ ` &#95; `&#95;`")
	want := "Missing  __COUNTER__ _ &#95;"
	if err != nil || got != want {
		t.Fatalf("ExtractPlainTextPreservingWhitespace() = %q, %v; want %q", got, err, want)
	}
}
