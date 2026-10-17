package subtitles

// iso3 maps the three letter language codes found inside video files to the two letter
// code and the name used for labels.
var iso3 = map[string][2]string{
	"eng": {"en", "English"}, "fra": {"fr", "French"}, "fre": {"fr", "French"}, "spa": {"es", "Spanish"},
	"por": {"pt", "Portuguese"}, "deu": {"de", "German"}, "ger": {"de", "German"}, "ita": {"it", "Italian"},
	"ara": {"ar", "Arabic"}, "zho": {"zh", "Chinese"}, "chi": {"zh", "Chinese"}, "jpn": {"ja", "Japanese"},
	"kor": {"ko", "Korean"}, "hin": {"hi", "Hindi"}, "rus": {"ru", "Russian"}, "tur": {"tr", "Turkish"},
	"nld": {"nl", "Dutch"}, "dut": {"nl", "Dutch"}, "pol": {"pl", "Polish"}, "swe": {"sv", "Swedish"},
	"yor": {"yo", "Yoruba"}, "ibo": {"ig", "Igbo"}, "hau": {"ha", "Hausa"}, "swa": {"sw", "Swahili"},
	"ind": {"id", "Indonesian"}, "tha": {"th", "Thai"}, "vie": {"vi", "Vietnamese"}, "heb": {"he", "Hebrew"},
	"ell": {"el", "Greek"}, "gre": {"el", "Greek"}, "ron": {"ro", "Romanian"}, "rum": {"ro", "Romanian"},
}

// textCodecs are the subtitle formats that can become text (WebVTT). Picture based tracks
// from DVDs and Blu-rays cannot.
var textCodecs = map[string]bool{
	"subrip": true, "srt": true, "ass": true, "ssa": true, "mov_text": true, "webvtt": true, "text": true,
}

// languageFor turns the language tag of a track inside a video ("eng") into the short code
// and the name shown to people. Unknown tags give "und" and no name.
func languageFor(tag string) (code, name string) {
	if l, ok := iso3[lower(tag)]; ok {
		return l[0], l[1]
	}
	return "und", ""
}
