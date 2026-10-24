package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSyntaxReferenceTransferAndRecovery(t *testing.T) {
	s, a, b := learningTestServer(t)
	model := learningTestModel(t, s, "syntax-fixture")
	calls := 0
	s.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		var input struct {
			Messages []chatMessage `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		content := ""
		if strings.Contains(input.Messages[0].Content, "创作不同主题") {
			content = `{"sentence":"The film inspired her."}`
		} else {
			if !strings.Contains(input.Messages[len(input.Messages)-1].Content, "我的独立译文") {
				t.Error("independent attempt not sent to teacher")
			}
			analysis := sentenceAnalysis{Translation: "这本书帮助了我。", MainClause: "The book helped me.", Chunks: []sentenceChunk{{"The book", "主语", "对照你选中的The book，这里是主语。"}}, Grammar: []string{"主谓宾"}, Paraphrase: "This book was helpful.", Check: sentenceCheck{"什么帮助了我？", "The book", "书是主语"}}
			raw, _ := json.Marshal(analysis)
			content = string(raw)
		}
		raw, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": content}, "finish_reason": "stop"}}})
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(raw))}, nil
	})}
	created := a.call("POST", "/api/v1/learning/english/syntax", map[string]string{"request_id": "syntax-with-ai-start", "text": "The book helped me."}, 200)
	id := rawString(created, "id")
	a.call("POST", "/api/v1/learning/english/syntax/"+id+"/reveal", map[string]any{"request_id": "syntax-with-ai-block", "model_id": model}, 409)
	if calls != 0 {
		t.Fatal("AI ran before independent attempt")
	}
	a.call("POST", "/api/v1/learning/english/syntax/"+id+"/attempt", map[string]any{"request_id": "syntax-with-ai-attempt", "marks": []syntaxMark{{0, 8, "subject"}}, "translation": "我的独立译文"}, 200)
	in := map[string]any{"request_id": "syntax-with-ai-reveal", "model_id": model}
	result := a.call("POST", "/api/v1/learning/english/syntax/"+id+"/reveal", in, 200)
	if string(result["analysis"]) == "null" {
		t.Fatal("no reference")
	}
	a.call("POST", "/api/v1/learning/english/syntax/"+id+"/reveal", in, 200)
	if calls != 1 {
		t.Fatal("reference retry called AI again", calls)
	}
	a.call("POST", "/api/v1/learning/english/syntax/"+id+"/attempt", map[string]any{"request_id": "syntax-after-reference", "translation": "peeked"}, 409)
	transfer := map[string]any{"request_id": "syntax-with-ai-transfer", "model_id": model}
	child := a.call("POST", "/api/v1/learning/english/syntax/"+id+"/transfer", transfer, 200)
	next := rawString(child, "id")
	a.call("POST", "/api/v1/learning/english/syntax/"+id+"/transfer", transfer, 200)
	if calls != 2 {
		t.Fatal("transfer replay called AI again", calls)
	}
	fresh := a.call("GET", "/api/v1/learning/english/syntax/"+next, nil, 200)
	if string(fresh["analysis"]) != "null" || rawString(fresh, "parent_id") != id {
		t.Fatal("new sentence leaked analysis or lost lineage")
	}
	own := a.call("GET", "/api/v1/learning/english/syntax", nil, 200)
	other := b.call("GET", "/api/v1/learning/english/syntax", nil, 200)
	if !strings.Contains(string(own["sessions"]), id) || string(other["sessions"]) != "[]" {
		t.Fatal("history isolation")
	}
	exported := a.call("GET", "/api/v1/account/export", nil, 200)
	if !strings.Contains(string(exported["syntax_attempts"]), "我的独立译文") {
		t.Fatal("attempt omitted from export")
	}
}

func TestEnglishIndependentAttemptAndIsolation(t *testing.T) {
	s, a, b := learningTestServer(t)
	learningInsertQuestion(t, s, "english", "reading")
	out := a.call("POST", "/api/v1/learning/attempts", map[string]string{"subject_id": "english", "kind": "reading", "request_id": "english-source-start"}, 200)
	aid := rawString(out, "id")
	input := map[string]string{"request_id": "syntax-create-1", "text": "The book helped me.", "source_attempt_id": aid}
	a.call("POST", "/api/v1/learning/english/syntax", input, 403)
	a.call("POST", "/api/v1/learning/attempts/"+aid+"/submit", map[string]any{"answers": map[string]string{"1": "A"}}, 200)
	created := a.call("POST", "/api/v1/learning/english/syntax", input, 200)
	id := rawString(created, "id")
	if string(created["analysis"]) != "null" {
		t.Fatal("analysis exposed before trying", created)
	}
	b.call("GET", "/api/v1/learning/english/syntax/"+id, nil, 404)
	a.call("POST", "/api/v1/learning/english/syntax/"+id+"/reveal", map[string]any{"request_id": "syntax-reveal-1", "model_id": 1}, 409)
	a.call("POST", "/api/v1/learning/english/syntax/"+id+"/attempt", map[string]any{"request_id": "syntax-attempt-1", "marks": []syntaxMark{{0, 999, "subject"}}}, 400)
	attempt := map[string]any{"request_id": "syntax-attempt-2", "marks": []syntaxMark{{0, 8, "subject"}}, "translation": "书帮助了我"}
	a.call("POST", "/api/v1/learning/english/syntax/"+id+"/attempt", attempt, 200)
	a.call("POST", "/api/v1/learning/english/syntax/"+id+"/attempt", attempt, 200)
	attempt["translation"] = "changed"
	a.call("POST", "/api/v1/learning/english/syntax/"+id+"/attempt", attempt, 409)
	if validSyntaxMarks("a😀b", []syntaxMark{{1, 3, "script"}}) {
		t.Fatal("unsafe role")
	}
	result := a.call("GET", "/api/v1/learning/english/syntax/"+id, nil, 200)
	if !strings.Contains(string(result["attempts"]), "subject") {
		t.Fatal("attempt lost")
	}
	a.call("POST", "/api/v1/learning/english/listening/nonexistent/reflection", map[string]string{"request_id": "listen-reflect-1", "reason": "heard_forgot"}, 403)
}
