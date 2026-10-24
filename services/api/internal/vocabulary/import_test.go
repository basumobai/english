package vocabulary

import "testing"

func TestPhraseImportAndClassification(t *testing.T) {
	for _, text := range []string{"look up,,,查阅", "take...for granted\t\t\t认为理所当然", "turn one’s back (on sb/sth),,,不理睬"} {
		p := Parse(text)
		if p.Valid() != nil || len(p.Entries) != 1 || Kind(p.Entries[0].Word) != "phrase" {
			t.Fatalf("phrase lost: %+v", p)
		}
	}
	for word, kind := range map[string]string{"AI (artificial intelligence)": "word", "2D (2-dimensional)": "word", "bachelor’s degree": "phrase", "take...for granted": "phrase", "so far": "phrase", "am (ante meridiem)": "word"} {
		if Kind(word) != kind {
			t.Error(word, Kind(word))
		}
	}
	for _, text := range []string{"look up,,,", "word,,,词", "=HYPERLINK(x),,,恶意公式", "look <script>,,,字符"} {
		if Parse(text).Valid() == nil {
			t.Error("invalid entry accepted", text)
		}
	}
}

func TestFourColumnImport(t *testing.T) {
	for _, text := range []string{"\ufeffword,phonetic,pos,chinese\nadapt,/əˈdæpt/,v.,适应\n\"look up\",/lʊk ʌp/,phr.,\"查询,查阅\"", "单词\t音标\t词性\t中文\nadapt\t/əˈdæpt/\tv.\t适应\nlook up\t/lʊk ʌp/\tphr.\t查阅", "adapt /əˈdæpt/ v. 适应\nlook up /lʊk ʌp/ phr. 查阅"} {
		p := Parse(text)
		if p.Valid() != nil || len(p.Entries) != 2 || p.Entries[1].Word != "look up" {
			t.Fatalf("valid four-field import failed: %+v", p)
		}
	}
	p := Parse("adapt /əˈdæpt/ v. 适应\nadapt /əˈdæpt/ v. 适应\nadapt /əˈdæpt/ n. 另一词性")
	if p.Valid() != nil || len(p.Entries) != 2 || p.Duplicates != 1 {
		t.Fatal("sense or duplicate handling failed")
	}
	capital := Parse("English /ˈɪŋɡlɪʃ/ n. 英语\nenglish /ˈɪŋɡlɪʃ/ n. 英语\nI /aɪ/ pron. 我")
	if capital.Valid() != nil || len(capital.Entries) != 2 || capital.Entries[0].Word != "English" || capital.Entries[1].Word != "I" || capital.Duplicates != 1 {
		t.Fatal("textbook capitalization lost or case duplicate retained")
	}
	for _, text := range []string{"adapt,missing,column", "adapt,,v.,适应", "=formula,/ipa/,v.,适应", "word,phonetic,pos,chinese", "adapt /ipa/ v. 适应\nbroken line", "adapt /ipa/ v. 适应\nadapt /other/ v. 适应"} {
		if Parse(text).Valid() == nil {
			t.Fatalf("invalid import accepted: %s", text)
		}
	}
}
