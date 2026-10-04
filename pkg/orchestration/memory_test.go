package orchestration

import "testing"

func TestStripMarkdown(t *testing.T) {
	in := "## Memory\n- **Riya**, prefers Hinglish\n* NEET exam stress"
	want := "Memory\nRiya, prefers Hinglish\nNEET exam stress"
	if got := stripMarkdown(in); got != want {
		t.Errorf("stripMarkdown = %q, want %q", got, want)
	}
}

func TestTruncateRunes(t *testing.T) {
	if got := truncateRunes("वन वन", 2); got != "वन…" {
		t.Errorf("truncateRunes = %q", got)
	}
	if got := truncateRunes("short", 10); got != "short" {
		t.Errorf("truncateRunes changed short text: %q", got)
	}
}
