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
		{"i feel hopeless", "hopeless", true},
		{"", "help", false},
	}
	for _, c := range cases {
		if got := containsPhrase(c.text, c.phrase); got != c.want {
			t.Errorf("containsPhrase(%q, %q) = %v, want %v", c.text, c.phrase, got, c.want)
		}
	}
}
