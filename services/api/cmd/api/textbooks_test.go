package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestTextbookImportScopeAndHistoryIntegration(t *testing.T) {
	s, a, _ := learningTestServer(t)
	for i := 0; i < 2; i++ {
		if err := s.syncTextbooks(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := s.db.QueryRow(`SELECT count(*) FROM textbook_entries`).Scan(&count); err != nil || count != 2137 {
		t.Fatal("seed/replay", count, err)
	}
	cat := a.call("GET", "/api/v1/textbooks", nil, 200)
	var books []map[string]any
	if json.Unmarshal(cat["books"], &books) != nil || len(books) != 7 {
		t.Fatal("catalog")
	}
	r := a.call("GET", "/api/v1/vocabulary?book_id=s1&unit=1&lesson=Lesson%201&starred=true&q=triangle", nil, 200)
	var entries []struct {
		ID        int64  `json:"id"`
		Word      string `json:"word"`
		Locations []struct {
			Starred bool   `json:"starred"`
			Lesson  string `json:"lesson"`
		} `json:"locations"`
	}
	if err := json.Unmarshal(r["entries"], &entries); err != nil || len(entries) != 1 || entries[0].Word != "triangle" {
		t.Fatal("filter", string(r["entries"]), err)
	}
	found := false
	for _, l := range entries[0].Locations {
		found = found || l.Starred && l.Lesson == "Lesson 1"
	}
	if !found {
		t.Fatal("lost occurrence asterisk")
	}
	for _, query := range []string{"book_id=c4", "book_id=s4&unit=1", "unit=1", "book_id=c1&unit=one", "starred=nope"} {
		a.call("GET", "/api/v1/vocabulary?"+query, nil, 400)
	}
	start := a.call("POST", "/api/v1/study/sessions", map[string]any{"mode": "practice", "types": []string{"spelling"}, "limit": 100, "book_id": "s1", "unit": 1, "lesson": "Lesson 1", "starred": true}, 201)
	var id int64
	json.Unmarshal(start["id"], &id)
	var outside int
	err := s.db.QueryRow(`SELECT count(*) FROM study_questions q WHERE session_id=$1 AND NOT EXISTS(SELECT 1 FROM textbook_entries e WHERE e.sense_id=q.sense_id AND e.book_id='s1' AND e.unit=1 AND e.lesson='Lesson 1' AND e.starred)`, id).Scan(&outside)
	if err != nil || outside != 0 {
		t.Fatal("practice escaped scope", outside, err)
	}
	var questions int
	s.db.QueryRow(`SELECT count(*) FROM study_questions WHERE session_id=$1`, id).Scan(&questions)
	if questions == 0 {
		t.Fatal("scope produced no questions")
	}
	if _, err = s.db.Exec(`INSERT INTO review_states(user_id,sense_id,dimension,attempts,correct_count,next_review_at) VALUES('demo',$1,'spelling',4,3,'2030-01-01')`, entries[0].ID); err != nil {
		t.Fatal(err)
	}
	a.call("POST", "/api/v1/vocabulary/suspend", map[string]any{"sense_id": entries[0].ID, "suspended": true}, 200)
	if err = s.syncTextbooks(context.Background()); err != nil {
		t.Fatal(err)
	}
	var retained bool
	s.db.QueryRow(`SELECT attempts=4 AND correct_count=3 AND next_review_at='2030-01-01' AND EXISTS(SELECT 1 FROM suspended_senses WHERE user_id='demo' AND sense_id=$1) FROM review_states WHERE user_id='demo' AND sense_id=$1 AND dimension='spelling'`, entries[0].ID).Scan(&retained)
	if !retained {
		t.Fatal("seed reset existing learning history")
	}
	a.call("GET", fmt.Sprintf("/api/v1/study/sessions/%d", id), nil, 200)
}
