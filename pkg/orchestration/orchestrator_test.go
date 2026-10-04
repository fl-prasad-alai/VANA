package orchestration

import "testing"

func TestContainsPhrase(t *testing.T) {
	cases := []struct {
		text, phrase string
		want         bool
	}{
		{"can you help me sleep", "help", true},
		{"that was really helpful", "help", false},
		{"i want to kill myself", "kill myself", true},
		{"i don't want to live anymore", "don't want to live", true},
		{"thinking about suicide.", "suicide", true},
		{"suicidesquad is a movie", "suicide", false},
		{"there were two suicides nearby", "suicide", true},
		{"i have been feeling suicidal", "suicide", true},
		{"i think i overdosed", "overdose", true},
		{"i keep overdosing", "overdose", true},
		{"i want to diet this month", "want to die", false},
		{"so much hopelessness", "hopeless", true},
		{"i feel helpless", "help", false},
		{"i feel hopeless", "hopeless", true},
		{"", "help", false},
	}
	for _, c := range cases {
		if got := containsPhrase(c.text, c.phrase); got != c.want {
			t.Errorf("containsPhrase(%q, %q) = %v, want %v", c.text, c.phrase, got, c.want)
		}
	}
}

func TestDevanagariShare(t *testing.T) {
	if got := devanagariShare("Aaj mann bhaari hai"); got != 0 {
		t.Errorf("latin text share = %v", got)
	}
	if got := devanagariShare("आज मन भारी है"); got < 0.99 {
		t.Errorf("devanagari text share = %v", got)
	}
	if hasDevanagari("hello") || !hasDevanagari("hello वन") {
		t.Error("hasDevanagari wrong")
	}
}
