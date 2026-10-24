package textbook

import "testing"

func TestSevenBookCorpus(t *testing.T) {
	books, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{"c1": 359, "c2": 310, "c3": 273, "s1": 306, "s2": 343, "s3": 308, "s4": 238}
	if len(books) != 7 {
		t.Fatal("missing book")
	}
	for _, b := range books {
		if len(b.Entries) != want[b.ID] {
			t.Fatalf("%s: incomplete corpus: %d", b.ID, len(b.Entries))
		}
		units := map[int]map[string]bool{}
		for _, e := range b.Entries {
			if units[e.Unit] == nil {
				units[e.Unit] = map[string]bool{}
			}
			units[e.Unit][e.Lesson] = true
		}
		for unit, lessons := range units {
			for _, lesson := range []string{"Topic Talk", "Lesson 1", "Lesson 2", "Lesson 3", "Writing Workshop"} {
				if !lessons[lesson] {
					t.Fatalf("%s Unit %d missing %s", b.ID, unit, lesson)
				}
			}
		}
	}
}

func TestTextbookCorrectionsAndStars(t *testing.T) {
	books, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	lookup := map[string]Entry{}
	for _, b := range books {
		for _, e := range b.Entries {
			lookup[b.ID+":"+e.Word] = e
		}
	}
	for key, lesson := range map[string]string{"c1:saying": "Lesson 1", "c1:humbly": "Reading Club 2", "c1:intend": "Lesson 3", "c1:pudding": "Lesson 2", "c2:independence": "Lesson 2", "c3:reflect on": "Lesson 1", "s2:philosophy": "Reading Club 2"} {
		if lookup[key].Lesson != lesson {
			t.Fatalf("wrong/missing source location %s", key)
		}
	}
	for _, key := range []string{"c1:meanwhile", "c1:gesture", "c3:carriage", "s1:triangle", "s1:poultry", "s3:outsource", "s4:orphan"} {
		if !lookup[key].Starred {
			t.Fatalf("lost textbook asterisk %s", key)
		}
	}
	if lookup["s1:passionate"].Phonetic != "/ˈpæʃənət/" || lookup["s1:occasionally"].POS != "adv." || lookup["s4:stadium"].Chinese != "体育场，运动场" {
		t.Fatal("secondary transcription errors returned")
	}
	if _, ok := lookup["s4:radium"]; ok {
		t.Fatal("stadium mistranscribed as radium")
	}
}
