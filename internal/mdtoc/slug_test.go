package mdtoc

import "testing"

func TestSluggerExamplesFromSpec(t *testing.T) {
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
