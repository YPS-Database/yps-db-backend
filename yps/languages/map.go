package ypsl

import (
	"fmt"
	"strings"
)

var nameHacks = map[string]string{}

var languageCodeMap = [][]string{
	{"Afrikaans", "Afrikaans", "af"},
	{"Amharic", "አማርኛ", "am"},
	{"Arabic", "العربية", "ar"},
	{"Mapudungun", "Mapudungun", "arn"},
	{"Moroccan Arabic", "الدارجة المغربية ", "ary"},
	{"Assamese", "অসমীয়া", "as"},
	{"Azerbaijani", "Azərbaycan", "az"},
	{"Bashkir", "Башҡорт", "ba"},
	{"Belarusian", "беларуская", "be"},
	{"Bulgarian", "български", "bg"},
	{"Bengali", "বাংলা", "bn"},
	{"Tibetan", "བོད་ཡིག", "bo"},
	{"Breton", "brezhoneg", "br"},
	{"Bosnian", "bosanski/босански", "bs"},
	{"Catalan", "català", "ca"},
	{"Central Kurdish", "کوردیی ناوەندی", "ckb"},
	{"Corsican", "Corsu", "co"},
	{"Czech", "čeština", "cs"},
	{"Welsh", "Cymraeg", "cy"},
	{"Danish", "dansk", "da"},
	{"German", "Deutsch", "de"},
	{"Lower Sorbian", "dolnoserbšćina", "dsb"},
	{"Divehi", "ދިވެހިބަސް", "dv"},
	{"Greek", "Ελληνικά", "el"},
	{"English", "English", "en"},
	{"Spanish", "español", "es"},
	{"Estonian", "eesti", "et"},
	{"Basque", "euskara", "eu"},
	{"Persian", "فارسى", "fa"},
	{"Finnish", "suomi", "fi"},
	{"Filipino", "Filipino", "fil"},
	{"Faroese", "føroyskt", "fo"},
	{"French", "français", "fr"},
	{"Frisian", "Frysk", "fy"},
	{"Irish", "Gaeilge", "ga"},
	{"Scottish Gaelic", "Gàidhlig", "gd"},
	{"Gilbertese", "Taetae ni Kiribati", "gil"},
	{"Galician", "galego", "gl"},
	{"Swiss German", "Schweizerdeutsch", "gsw"},
	{"Gujarati", "ગુજરાતી", "gu"},
	{"Hausa", "Hausa", "ha"},
	{"Hebrew", "עברית", "he"},
	{"Hindi", "हिंदी", "hi"},
	{"Croatian", "hrvatski", "hr"},
	{"Serbo-Croatian", "srpskohrvatski/српскохрватски", "hrv"},
	{"Upper Sorbian", "hornjoserbšćina", "hsb"},
	{"Hungarian", "magyar", "hu"},
	{"Armenian", "Հայերեն", "hy"},
	{"Indonesian", "Bahasa Indonesia", "id"},
	{"Igbo", "Igbo", "ig"},
	{"Yi", "ꆈꌠꁱꂷ", "ii"},
	{"Icelandic", "íslenska", "is"},
	{"Italian", "italiano", "it"},
	{"Inuktitut", "Inuktitut /ᐃᓄᒃᑎᑐᑦ (ᑲᓇᑕ)", "iu"},
	{"Japanese", "日本語", "ja"},
	{"Georgian", "ქართული", "ka"},
	{"Kazakh", "Қазақша", "kk"},
	{"Greenlandic", "kalaallisut", "kl"},
	{"Khmer", "ខ្មែរ", "km"},
	{"Kannada", "ಕನ್ನಡ", "kn"},
	{"Korean", "한국어", "ko"},
	{"Konkani", "कोंकणी", "kok"},
	{"Kurdish", "Kurdî/کوردی", "ku"},
	{"Kyrgyz", "Кыргыз", "ky"},
	{"Luxembourgish", "Lëtzebuergesch", "lb"},
	{"Lao", "ລາວ", "lo"},
	{"Lithuanian", "lietuvių", "lt"},
	{"Latvian", "latviešu", "lv"},
	{"Maori", "Reo Māori", "mi"},
	{"Macedonian", "македонски јазик", "mk"},
	{"Malayalam", "മലയാളം", "ml"},
	{"Mongolian", "Монгол хэл/ᠮᠤᠨᠭᠭᠤᠯ ᠬᠡᠯᠡ", "mn"},
	{"Mohawk", "Kanien'kéha", "moh"},
	{"Marathi", "मराठी", "mr"},
	{"Malay", "Bahasa Malaysia", "ms"},
	{"Maltese", "Malti", "mt"},
	{"Burmese", "မြန်မာဘာသာ", "my"},
	{"Norwegian (Bokmål)", "norsk (bokmål)", "nb"},
	{"Nepali", "नेपाली (नेपाल)", "ne"},
	{"Dutch", "Nederlands", "nl"},
	{"Norwegian (Nynorsk)", "norsk (nynorsk)", "nn"},
	{"Norwegian", "norsk", "no"},
	{"Occitan", "occitan", "oc"},
	{"Odia", "ଓଡ଼ିଆ", "or"},
	{"Papiamento", "Papiamentu", "pap"},
	{"Punjabi", "ਪੰਜਾਬੀ / پنجابی", "pa"},
	{"Polish", "polski", "pl"},
	{"Dari", "درى", "prs"},
	{"Pashto", "پښتو", "ps"},
	{"Portuguese", "português", "pt"},
	{"K'iche", "K'iche", "quc"},
	{"Quechua", "runasimi", "qu"},
	{"Romansh", "Rumantsch", "rm"},
	{"Romanian", "română", "ro"},
	{"Russian", "русский", "ru"},
	{"Kinyarwanda", "Kinyarwanda", "rw"},
	{"Sanskrit", "संस्कृत", "sa"},
	{"Yakut", "саха", "sah"},
	{"Sami (Northern)", "davvisámegiella", "se"},
	{"Sinhala", "සිංහල", "si"},
	{"Slovak", "slovenčina", "sk"},
	{"Slovenian", "slovenski", "sl"},
	{"Sami (Southern)", "åarjelsaemiengiele", "sma"},
	{"Sami (Lule)", "julevusámegiella", "smj"},
	{"Sami (Inari)", "sämikielâ", "smn"},
	{"Sami (Skolt)", "sääʹmǩiõll", "sms"},
	{"Albanian", "shqip", "sq"},
	{"Serbian", "srpski/српски", "sr"},
	{"Sesotho", "Sesotho sa Leboa", "st"},
	{"Swedish", "svenska", "sv"},
	{"Kiswahili", "Kiswahili", "sw"},
	{"Syriac", "ܣܘܪܝܝܐ", "syc"},
	{"Tamil", "தமிழ்", "ta"},
	{"Telugu", "తెలుగు", "te"},
	{"Tajik", "Тоҷикӣ", "tg"},
	{"Thai", "ไทย", "th"},
	{"Turkmen", "türkmençe", "tk"},
	{"Tswana", "Setswana", "tn"},
	{"Turkish", "Türkçe", "tr"},
	{"Tatar", "Татарча", "tt"},
	{"Tamazight", "Tamazight", "tzm"},
	{"Uyghur", "ئۇيغۇرچە", "ug"},
	{"Ukrainian", "українська", "uk"},
	{"Urdu", "اُردو", "ur"},
	{"Uzbek", "Uzbek/Ўзбек", "uz"},
	{"Vietnamese", "Tiếng Việt", "vi"},
	{"Wolof", "Wolof", "wo"},
	{"Xhosa", "isiXhosa", "xh"},
	{"Yoruba", "Yoruba", "yo"},
	{"Chinese", "中文", "zh"},
	{"Zulu", "isiZulu", "zu"},
}

func GetCode(name string) (code string, err error) {
	name = strings.TrimSpace(name)
	fixedName, fixFound := nameHacks[strings.ToLower(name)]
	if fixFound {
		name = fixedName
	}

	for _, value := range languageCodeMap {
		if strings.EqualFold(name, value[0]) || strings.EqualFold(name, value[1]) {
			return value[2], nil
		}
	}
	return "", fmt.Errorf("could not find language code for [%s]", name)
}

func GetName(code string) string {
	code = strings.TrimSpace(code)
	for _, value := range languageCodeMap {
		if strings.EqualFold(code, value[2]) {
			return value[0]
		}
	}
	return "Unknown"
}
