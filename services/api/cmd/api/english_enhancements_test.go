package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"example.com/els/api/internal/vocabulary"
)

func TestPhraseScopeAndTwoSecondRounds(t *testing.T) {
	s, a, _ := learningTestServer(t)
	if err := s.syncTextbooks(context.Background()); err != nil {
		t.Fatal(err)
	}
	rows, err := s.db.Query(`SELECT lemma,entry_kind FROM words`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var word, kind string
		if err = rows.Scan(&word, &kind); err != nil {
			t.Fatal(err)
		}
		if kind != vocabulary.Kind(word) {
			t.Fatal("SQL/Go classification mismatch", word, kind)
		}
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	r := a.call("GET", "/api/v1/vocabulary?book_id=c1&entry_kind=phrase", nil, 200)
	var es []struct {
		Word string `json:"word"`
		Kind string `json:"entry_kind"`
	}
	if json.Unmarshal(r["entries"], &es) != nil || len(es) == 0 {
		t.Fatal("no phrases")
	}
	for _, e := range es {
		if e.Kind != "phrase" || vocabulary.Kind(e.Word) != "phrase" {
			t.Fatal("filter escaped", e)
		}
	}
	a.call("GET", "/api/v1/vocabulary?entry_kind=invalid", nil, 400)
	var phrase int64
	if err = s.db.QueryRow(`SELECT sense_id FROM textbook_entries WHERE book_id='c1' AND word='all in all' LIMIT 1`).Scan(&phrase); err != nil {
		t.Fatal(err)
	}
	_, err = s.db.Exec(`INSERT INTO learning_word_sets(id,sense_id,direction,difficulty,prompt,choices,correct_id,rationale,approved) VALUES('phrase-set',$1,'en_zh','normal','all in all','[{"id":"A","text":"总而言之"},{"id":"B","text":"明天"},{"id":"C","text":"笔"},{"id":"D","text":"红色"}]','A','已核对唯一答案',true)`, phrase)
	if err != nil {
		t.Fatal(err)
	}
	input := map[string]any{"direction": "en_zh", "difficulty": "normal", "request_id": "phrase-round-two-seconds", "seconds": 2, "book_id": "c1", "entry_kind": "phrase"}
	r = a.call("POST", "/api/v1/learning/words/next", input, 201)
	id := rawString(r, "id")
	if _, ok := r["vocabulary"]; ok {
		t.Fatal("answer vocabulary exposed early")
	}
	if _, ok := r["rationale"]; ok {
		t.Fatal("rationale exposed early")
	}
	var ms int64
	json.Unmarshal(r["duration_ms"], &ms)
	if ms < 1990 || ms > 2010 {
		t.Fatal("wrong deadline", ms)
	}
	replay := a.call("POST", "/api/v1/learning/words/next", input, 200)
	if rawString(replay, "id") != id {
		t.Fatal("retry made a new round")
	}
	input["seconds"] = 6
	a.call("POST", "/api/v1/learning/words/next", input, 409)
	if _, err = s.db.Exec(`UPDATE learning_word_rounds SET deadline=now()-interval '1 second' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	r = a.call("POST", "/api/v1/learning/words/"+id+"/answer", map[string]string{"picked": "A"}, 200)
	if string(r["correct"]) != "false" || len(r["vocabulary"]) == 0 {
		t.Fatal("late answer or missing feedback", r)
	}
	a.call("POST", "/api/v1/vocabulary/suspend", map[string]any{"sense_id": phrase, "suspended": true}, 200)
	input["request_id"] = "phrase-round-suspended"
	a.call("POST", "/api/v1/learning/words/next", input, 404)
	input["seconds"] = 4
	a.call("POST", "/api/v1/learning/words/next", input, 400)
	session := a.call("POST", "/api/v1/study/sessions", map[string]any{"mode": "practice", "types": []string{"spelling"}, "limit": 100, "book_id": "c1", "entry_kind": "phrase"}, 201)
	var sid int64
	json.Unmarshal(session["id"], &sid)
	var outside, total int
	s.db.QueryRow(`SELECT count(*),count(*) FILTER(WHERE w.entry_kind<>'phrase') FROM study_questions q JOIN word_senses ws ON ws.id=q.sense_id JOIN words w ON w.id=ws.word_id WHERE q.session_id=$1`, sid).Scan(&total, &outside)
	if total == 0 || outside != 0 {
		t.Fatal("study escaped phrase scope", total, outside)
	}
}

func TestListeningReflectionRestoreAndIsolation(t *testing.T) {
	s, a, b := learningTestServer(t)
	_, err := s.db.Exec(`INSERT INTO training_sessions(id,user_id,kind,title,payload,answers) VALUES('reflect-lesson','demo','listening','fixture','{}','{}')`)
	if err != nil {
		t.Fatal(err)
	}
	in := map[string]string{"request_id": "reflect-save-one", "reason": "heard_forgot", "notes": "Friday, three o’clock"}
	a.call("POST", "/api/v1/learning/english/listening/reflect-lesson/reflection", in, 200)
	a.call("POST", "/api/v1/learning/english/listening/reflect-lesson/reflection", in, 200)
	restored := a.call("GET", "/api/v1/learning/english/listening/reflect-lesson/reflection", nil, 200)
	var reflection struct {
		Reason  string    `json:"reason"`
		Notes   string    `json:"notes"`
		Created time.Time `json:"created_at"`
	}
	json.Unmarshal(restored["reflection"], &reflection)
	if reflection.Reason != "heard_forgot" || reflection.Notes != in["notes"] || reflection.Created.IsZero() {
		t.Fatal("reflection lost", reflection)
	}
	b.call("GET", "/api/v1/learning/english/listening/reflect-lesson/reflection", nil, 404)
	b.call("POST", "/api/v1/learning/english/listening/reflect-lesson/reflection", in, 403)
	in["notes"] = "changed"
	a.call("POST", "/api/v1/learning/english/listening/reflect-lesson/reflection", in, 409)
}
