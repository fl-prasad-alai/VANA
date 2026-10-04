package orchestration

import (
	"strings"
	"testing"
)

func severityOf(msg string) string {
	m := normalizeIndic(strings.ToLower(msg))
	best := ""
	for _, k := range builtinCrisisKeywords {
		if containsPhrase(m, normalizeIndic(strings.ToLower(k.Keyword))) {
			if k.Severity == "high" {
				return "high"
			}
			best = k.Severity
		}
	}
	return best
}

func TestIndicCrisisDetection(t *testing.T) {
	cases := map[string]string{
		"आजकल जीने का मन ही नहीं करता":                         "high",
		"आजकल जीने का मन नहीं करता":                            "high",
		"लगता है मेरे ना होने से किसी को कोई फर्क नहीं पड़ेगा": "high",
		"sab khatam kar dene ka mann karta hai kabhi kabhi":    "high",
		"जगण्याचा कंटाळा आलाय":                                 "high",
		"jagnyacha kantala aalay": "high",
		"कोणालाच माझी गरज नाही असं वाटतं":     "medium",
		"sometimes I wonder what's the point": "medium",
		// hyperbole and ordinary distress must not trigger the handoff
		"exam mein toh mar gaya aaj":    "",
		"bahut kantala aala office cha": "",
		"मनावर खूप ओझं आहे":             "",
		"सब कुछ है फिर भी खुशी नहीं है": "",
	}
	for msg, want := range cases {
		if got := severityOf(msg); got != want {
			t.Errorf("%q: severity %q, want %q", msg, got, want)
		}
	}
}

func TestCrisisReplyLanguage(t *testing.T) {
	cases := map[string]string{
		"आजकल जीने का मन नहीं करता":       crisisHindi,
		"जगण्याचा कंटाळा आलाय":            crisisMarathi,
		"sab khatam kar dene ka mann hai": crisisHinglish,
		"jagnyacha kantala aalay mala":    crisisMarathiRoman,
		"I want to end my life":           crisisEnglish,
	}
	for msg, want := range cases {
		if got := crisisReply(msg, false); got != want {
			t.Errorf("%q: got reply starting %q", msg, got[:40])
		}
	}
	for _, r := range []string{crisisEnglish, crisisHindi, crisisHinglish, crisisMarathi, crisisMarathiRoman} {
		if !strings.Contains(r, "14416") || !strings.Contains(r, "112") {
			t.Error("crisis reply missing a helpline")
		}
	}
}

func TestCrisisFollowUp(t *testing.T) {
	first, again := crisisReply("जगण्याचा कंटाळा आलाय", false), crisisReply("जगण्याचा कंटाळा आलाय", true)
	if first == again || again != crisisAgainMarathi {
		t.Error("repeat crisis should use the shorter Marathi follow-up")
	}
	for _, r := range []string{crisisEnglish, crisisHindi, crisisAgainHinglish, crisisAgainMarathiRoman} {
		if !isCrisisTemplate(r) || !strings.Contains(r, "14416") || !strings.Contains(r, "112") {
			t.Errorf("crisis text must be recognisable and carry both helplines: %.30q", r)
		}
	}
	if isCrisisTemplate("Take a slow breath with me.") {
		t.Error("ordinary reply detected as crisis template")
	}
}

func TestGuessesGender(t *testing.T) {
	gendered := []string{
		"ऐसा लगता है कि आप खुद के बारे में सवाल कर रही हैं।",
		"आप बहुत थके होंगे",
		"मैं समझ सकता हूँ कि यह मुश्किल है",
		"तू खूप थकलीस",
		"lagta hai aap bahut pressure mein jee rahe ho",
		"main samajh sakti hoon",
	}
	neutral := []string{
		"आपको कैसा लग रहा है?",
		"आपने हिम्मत की, यह आसान नहीं था।",
		"घरवाले कह रहे हैं कि यह वहम है, और आपको अकेलापन लग रहा है।",
		"तुम्हाला कसं वाटतंय?",
		"Aapko kaisa lag raha hai? Main yahin hoon.",
		"It sounds like you're carrying a lot.",
	}
	for _, s := range gendered {
		if !guessesGender(s) {
			t.Errorf("should flag gendered address: %q", s)
		}
	}
	for _, s := range neutral {
		if guessesGender(s) {
			t.Errorf("false positive on neutral text: %q", s)
		}
	}
}
