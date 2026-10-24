package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnglishScope(t *testing.T) {
	for _, tc := range []struct {
		path, body string
		status     int
	}{
		{"/api/v1/learning/attempts", `{"subject_id":"english","answers":{"1":"physics"}}`, 204},
		{"/api/v1/learning/attempts", `{"subject_id":"math"}`, 400},
		{"/api/v1/admin/resources/x/tags", `{"classification":{"subjects":["english","physics"]}}`, 400},
		{"/api/v1/resources?subject_id=biology", `{}`, 400},
		{"/api/v1/resources?subject_id=english&subject_id=chinese", `{}`, 400},
		{"/api/v1/admin/resources/x/tags", `{"subject_id":"unknown"}`, 204},
	} {
		t.Run(tc.path+tc.body, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, _ := io.ReadAll(r.Body)
				if string(b) != tc.body {
					t.Fatal("request body changed")
				}
				w.WriteHeader(204)
			})
			r := httptest.NewRequest("POST", tc.path, bytes.NewBufferString(tc.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			englishScope(next).ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}
func TestEnglishScopeMultipart(t *testing.T) {
	var body bytes.Buffer
	m := multipart.NewWriter(&body)
	_ = m.WriteField("subject_id", "physics")
	_ = m.Close()
	r := httptest.NewRequest("POST", "/api/v1/admin/learning/documents", &body)
	r.Header.Set("Content-Type", m.FormDataContentType())
	w := httptest.NewRecorder()
	englishScope(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("non-English upload reached handler") })).ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatalf("got %d", w.Code)
	}
}
