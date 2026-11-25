package diff

import "testing"

func Test_LineDiffNoWhitespace_IgnoresIndentation(t *testing.T) {
	o2(t, "abc\ndef", "  abc\n\t def  ", "")
}

func Test_LineDiffNoWhitespace_IgnoresInternalWhitespace(t *testing.T) {
	o2(t, "a    b", "a b", "")
}

func Test_LineDiffNoWhitespace_ShowsRealChanges(t *testing.T) {
	o2(t, "abc\ndef", "abc\nxyz", " abc\n-def\n+xyz")
}

func Test_LineDiffNoWhitespace_MixedChanges(t *testing.T) {
	o2(t, "  abc\n  def\n  xyz", "abc\ndef   changed\nxyz", " abc\n-def\n+def changed\n xyz")
}

func Test_LineDiffNoWhitespace_WhitespaceInChangedLines(t *testing.T) {
	o2(t, "a  b\nc", "a b\nd", " a b\n-c\n+d")
}

func o2(t *testing.T, a, b, expected string) {
	t.Helper()
	actual := LineDiffNoWhitespace(a, b)
	if actual != expected {
		t.Errorf("Got %#v instead of expected %#v", actual, expected)
	}
}
