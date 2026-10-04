// pkg/orchestration/crisis_indic.go

package orchestration

import (
	"regexp"
	"strings"

	"emerald-moss-api/pkg/database"
)

// builtinCrisisKeywords cover how people in India commonly express wanting to
// die, in Hindi, Hinglish and Marathi. They are versioned with the code and
// added to the keywords in the database. Native-speaker and clinician review
// is still recommended before relying on them alone.
var builtinCrisisKeywords = func() []database.CrisisKeyword {
	high := []string{
		// Hindi (Devanagari)
		"जीने का मन नहीं", "जीने का मन ही नहीं", "जीने की इच्छा नहीं", "जीने की इच्छा ही नहीं", "मर जाना चाहता", "मर जाना चाहती", "मरना चाहता", "मरना चाहती",
		"मर जाऊं", "मर जाऊँ", "आत्महत्या", "खुद को खत्म", "अपने आप को खत्म", "सब खत्म कर दूं", "सब खत्म कर दूँ",
		"जान दे दूं", "जान दे दूँ", "जिंदगी खत्म", "मेरे ना होने से", "मेरे न होने से", "मेरे ना रहने से",
		// Hinglish
		"jeene ka mann nahi", "jeene ka man nahi", "jine ka mann nahi", "jine ka man nahi", "jeene ka mann hi nahi", "jeene ka man hi nahi", "jine ka mann hi nahi", "sab khatam kar",
		"mar jaun", "mar jaaun", "mar jana chahta", "mar jaana chahta", "mar jana chahti", "mar jaana chahti",
		"marna chahta", "marna chahti", "khud ko khatam", "jaan de dun", "zindagi khatam", "jindagi khatam",
		"mere na hone se", "mere na rehne se",
		// Marathi (Devanagari)
		"जगण्याचा कंटाळा", "जगायचा कंटाळा", "जगावंसं वाटत नाही", "जगावसं वाटत नाही", "जगावंसंच वाटत नाही", "सगळं संपवून", "मरून जावं",
		"मेलेलं बरं", "जीव द्यावा", "स्वतःला संपव", "माझ्याशिवाय सगळ्यांचं भलं",
		// Marathi (romanised)
		"jagnyacha kantala", "jagaycha kantala", "jagnyacha kantal", "sagla sampavun", "sagla sampvun",
		"sampavun takava", "marun jav", "melela bara", "jeev dyava",
		// English phrasings the database list misses
		"no reason to live", "better off without me", "don't want to wake up", "dont want to wake up",
		"want to disappear forever", "end it all",
	}
	medium := []string{
		"what's the point", "whats the point", "what the point", "what is the point", "point of all this",
		"point of anything", "no point in living", "nothing matters anymore", "why even try", "कोई फर्क नहीं पड़ेगा", "kisi ko fark nahi",
		"कोणालाच माझी गरज नाही", "konalach mazi garaj nahi", "बोझ हूं", "बोझ हूँ", "bojh hoon",
	}
	selfHarm, hopelessness := "self-harm", "hopelessness"
	var out []database.CrisisKeyword
	for _, k := range high {
		out = append(out, database.CrisisKeyword{Keyword: k, Severity: "high", Category: &selfHarm})
	}
	for _, k := range medium {
		out = append(out, database.CrisisKeyword{Keyword: k, Severity: "medium", Category: &hopelessness})
	}
	return out
}()

// normalizeIndic folds spelling variants that differ only in nukta or
// chandrabindu, so "ख़त्म" matches "खत्म" and "दूँ" matches "दूं"
func normalizeIndic(s string) string {
	s = strings.ReplaceAll(s, "़", "")
	return strings.ReplaceAll(s, "ँ", "ं")
}

var (
	marathiDeva  = regexp.MustCompile(`आहे|नाही|मला|तुम्ही|वाटत|आलाय|आलंय|करायच|कंटाळा|सगळं|जगण|जगाय|माझ`)
	marathiRoman = regexp.MustCompile(`\b(aahe|ahe|mala|nahiye|vatat|vatta|kantala|jagnyacha|jagaycha|sagla|majhi|mazi|konalach)\b`)
	hindiRoman   = regexp.MustCompile(`\b(hai|nahi|mann|man|karta|karti|mujhe|main|hoon|hun|kya|kar|dun|jaun|jeene|jine|lagta|sab)\b`)
)

// crisisReply returns the fixed crisis message in the person's language and
// script. When the previous reply was already a crisis message, a shorter
// follow-up is used instead of repeating the same text.
func crisisReply(message string, repeat bool) string {
	lower := strings.ToLower(message)
	pick := func(first, again string) string {
		if repeat {
			return again
		}
		return first
	}
	switch {
	case hasDevanagari(message) && marathiDeva.MatchString(message):
		return pick(crisisMarathi, crisisAgainMarathi)
	case hasDevanagari(message):
		return pick(crisisHindi, crisisAgainHindi)
	case marathiRoman.MatchString(lower):
		return pick(crisisMarathiRoman, crisisAgainMarathiRoman)
	case hindiRoman.MatchString(lower):
		return pick(crisisHinglish, crisisAgainHinglish)
	default:
		return pick(crisisEnglish, crisisAgainEnglish)
	}
}

// isCrisisTemplate reports whether a stored VANA reply was one of the fixed crisis messages
func isCrisisTemplate(reply string) bool {
	return strings.Contains(reply, "**Tele-MANAS")
}

// genderedAddress catches Hindi/Hinglish/Marathi verb forms that assume the
// person's gender (or give VANA a gender), e.g. "आप कर रही हैं", "main sun raha hoon"
var genderedAddress = regexp.MustCompile(
	`आप [^।?!.\n]{0,40}?(रही|रहे|सकती|सकते|गई|गए|सोचती|सोचते|चाहती|चाहते|करती|करते|होती|होते|थकी|थके) (हैं|हो|होंगी|होंगे)` +
		`|मैं [^।?!.\n]{0,30}?(रहा|रही|सकता|सकती|चाहता|चाहती) (हूँ|हूं)` +
		`|थकली|थकलास|थकलीस|आलीस|आलास|गेलीस|गेलास|करतेस|करतोस` +
		`|(?i:\baap\b[^.?!\n]{0,40}?\b(rahi|rahe|sakti|sakte|gayi|gaye|sochti|sochte|chahti|chahte|karti|karte)\s+(hain|ho)\b)` +
		`|(?i:\bmain\b[^.?!\n]{0,30}?\b(raha|rahi|sakta|sakti|chahta|chahti)\s+(hoon|hun|hu)\b)`)

func guessesGender(reply string) bool { return genderedAddress.MatchString(reply) }

const crisisEnglish = `I'm really glad you told me. What you're carrying sounds unbearably heavy right now, and you deserve a real person beside you, not just me.

Please call **Tele-MANAS on 14416** (or **1-800-891-4416**). It's free, confidential, open 24x7, and you can talk in your own language.
If you are in immediate danger, call **112** or go to the nearest hospital.

If you can, also tell one person you trust how you're feeling right now. I'm right here with you.`

const crisisHindi = `आपने यह बात मुझसे कही, इसके लिए शुक्रिया। सुनकर लग रहा है कि मन पर बहुत भारी बोझ है, और इस वक़्त आपके पास एक असली इंसान का साथ होना ज़रूरी है।

कृपया अभी **Tele-MANAS को 14416** (या **1-800-891-4416**) पर कॉल कीजिए। यह मुफ़्त है, गोपनीय है, 24x7 खुला है, और अपनी भाषा में बात की जा सकती है।
अगर अभी ख़तरा है, तो **112** पर कॉल कीजिए या पास के अस्पताल जाइए।

हो सके तो किसी भरोसेमंद इंसान को भी अभी बताइए कि आपको कैसा लग रहा है। मैं यहीं हूँ।`

const crisisHinglish = `Aapne mujhse ye baat kahi, iske liye shukriya. Sunke lag raha hai ki mann pe bahut bhaari bojh hai, aur abhi aapke saath ek asli insaan ka hona zaroori hai.

Please abhi **Tele-MANAS ko 14416** (ya **1-800-891-4416**) pe call kijiye. Ye free hai, confidential hai, 24x7 khula hai, aur apni bhasha mein baat ki ja sakti hai.
Agar abhi khatra hai to **112** pe call kijiye ya paas ke hospital jaiye.

Ho sake to kisi bharosemand insaan ko bhi abhi bataiye ki aapko kaisa lag raha hai. Main yahin hoon.`

const crisisMarathi = `तुम्ही हे मला सांगितलंत, त्याबद्दल मनापासून धन्यवाद. ऐकून असं वाटतंय की मनावर खूप मोठं ओझं आहे, आणि आत्ता तुमच्यासोबत एखादी खरी व्यक्ती असणं गरजेचं आहे.

कृपया आत्ताच **Tele-MANAS ला 14416** (किंवा **1-800-891-4416**) वर फोन करा. हे मोफत, गोपनीय आणि 24x7 सुरू आहे, आणि मराठीतही बोलता येतं.
लगेच धोका वाटत असेल तर **112** वर फोन करा किंवा जवळच्या रुग्णालयात जा.

जमलं तर विश्वासातल्या एखाद्या व्यक्तीला आत्ता तुम्हाला कसं वाटतंय ते सांगा. मी इथेच आहे.`

const crisisMarathiRoman = `Tumhi he mala sangitlat, tyabaddal manapasun dhanyavaad. Aikun asa vatatay ki manavar khup motha ojha aahe, ani atta tumchyasobat ekhadi khari vyakti asna garjecha aahe.

Krupaya attach **Tele-MANAS la 14416** (kiva **1-800-891-4416**) var phone kara. He moft, gopniya ani 24x7 suru aahe, ani Marathitahi bolta yeta.
Lagech dhoka vatat asel tar **112** var phone kara kiva javalchya hospital madhe ja.

Jamla tar vishwasatlya ekhadya vyaktila atta tumhala kasa vatatay te sanga. Mi ithech aahe.`

const crisisAgainEnglish = `I'm still right here with you. You don't have to carry this alone tonight.

Have you been able to call **Tele-MANAS on 14416**? If you are in danger right now, please call **112**.
If calling feels hard, could someone you trust come and sit with you while you do?`

const crisisAgainHindi = `मैं अभी भी यहीं हूँ, आपके साथ। यह सब अकेले उठाना ज़रूरी नहीं है।

क्या **Tele-MANAS (14416)** पर बात हो पाई? अगर अभी ख़तरा है, तो कृपया **112** पर कॉल कीजिए।
कॉल करना मुश्किल लगे, तो किसी भरोसेमंद इंसान को पास बुला लीजिए।`

const crisisAgainHinglish = `Main abhi bhi yahin hoon, aapke saath. Ye sab akele uthana zaroori nahi hai.

Kya **Tele-MANAS (14416)** pe baat ho paayi? Agar abhi khatra hai to please **112** pe call kijiye.
Call karna mushkil lage to kisi bharosemand insaan ko paas bula lijiye.`

const crisisAgainMarathi = `मी अजूनही इथेच आहे, तुमच्यासोबत. हे सगळं एकट्याने पेलायची गरज नाही.

**Tele-MANAS (14416)** वर बोलणं झालं का? लगेच धोका वाटत असेल तर कृपया **112** वर फोन करा.
फोन करणं अवघड वाटत असेल, तर विश्वासातल्या एखाद्या व्यक्तीला जवळ बोलवा.`

const crisisAgainMarathiRoman = `Mi ajunhi ithech aahe, tumchyasobat. He sagla ektyane pelaychi garaj nahi.

**Tele-MANAS (14416)** var bolna zhala ka? Lagech dhoka vatat asel tar krupaya **112** var phone kara.
Phone karna avghad vatat asel tar vishwasatlya ekhadya vyaktila javal bolva.`
