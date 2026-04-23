package mdtoc

import "testing"

// TestGitHubSluggerExamplesFromSpec verifies the documented GitHub-compatible examples and collision handling.
func TestGitHubSluggerExamplesFromSpec(t *testing.T) {
	slugger := NewSlugger()
	cases := []struct{ input, want string }{
		{"Open source", "open-source"},
		{"This'll be a Helpful Section About the Greek Letter Θ!", "thisll-be-a-helpful-section-about-the-greek-letter-θ"},
		{"Übergrößenträger & naïve façade – déjà vu!", "übergrößenträger-naïve-façade-déjà-vu"},
		{"中文 русский عربى", "中文-русский-عربى"},
		{"🚀 !!!", "section"},
		{"API", "api"},
		{"API", "api-1"},
	}
	for _, tc := range cases {
		if got := slugger.Next(tc.input); got != tc.want {
			t.Fatalf("slugger.Next(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// TestGitLabSluggerExamplesFromDocs verifies the documented GitLab heading-ID rules and collision handling.
func TestGitLabSluggerExamplesFromDocs(t *testing.T) {
	slugger := NewGitLabSlugger()
	cases := []struct{ input, want string }{
		{"This heading has spaces in it", "this-heading-has-spaces-in-it"},
		{"This heading has a :thumbsup: in it", "this-heading-has-a-thumbsup-in-it"},
		{"This heading has Unicode in it: 한글", "this-heading-has-unicode-in-it-한글"},
		{"This heading has spaces in it", "this-heading-has-spaces-in-it-1"},
		{"This heading has spaces in it", "this-heading-has-spaces-in-it-2"},
		{"This heading has 3.5 in it (and parentheses)", "this-heading-has-35-in-it-and-parentheses"},
	}
	for _, tc := range cases {
		if got := slugger.Next(tc.input); got != tc.want {
			t.Fatalf("gitlab slugger.Next(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// TestSluggerProfilesDiffer documents the intentional GitHub/GitLab behavior differences.
func TestSluggerProfilesDiffer(t *testing.T) {
	cases := []struct {
		input       string
		github, lab string
	}{
		{input: "Version 3.5", github: "version-3-5", lab: "version-35"},
		{input: "A+B", github: "a-b", lab: "ab"},
		{input: "foo_bar baz", github: "foo-bar-baz", lab: "foo_bar-baz"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			if got := NewSlugger().Next(tc.input); got != tc.github {
				t.Fatalf("github slugger.Next(%q) = %q, want %q", tc.input, got, tc.github)
			}
			if got := NewGitLabSlugger().Next(tc.input); got != tc.lab {
				t.Fatalf("gitlab slugger.Next(%q) = %q, want %q", tc.input, got, tc.lab)
			}
		})
	}
}

// TestExtractPlainText verifies inline Markdown text extraction for headings.
func TestExtractPlainText(t *testing.T) {
	got, err := ExtractPlainText("This'll be a _Helpful_ [Section](#x) about `Go`")
	if err != nil {
		t.Fatalf("ExtractPlainText error: %v", err)
	}
	want := "This'll be a Helpful Section about Go"
	if got != want {
		t.Fatalf("ExtractPlainText() = %q, want %q", got, want)
	}
}
