package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// The shared parsers can recognize an out-of-scope file. Only English content
// may be configured, practiced or published by this application's HTTP API.
func englishSubjects(value any) bool {
	switch v := value.(type) {
	case map[string]any:
		for key, item := range v {
			if key == "subject_id" || key == "subject" {
				if s, ok := item.(string); ok && !englishSubject(s) {
					return false
				}
			}
			if key == "subjects" || key == "subject_ids" {
				if subjects, ok := item.([]any); ok {
					for _, subject := range subjects {
						if s, ok := subject.(string); ok && !englishSubject(s) {
							return false
						}
					}
				}
			}
			if !englishSubjects(item) {
				return false
			}
		}
	case []any:
		for _, item := range v {
			if !englishSubjects(item) {
				return false
			}
		}
	}
	return true
}
func englishSubject(subject string) bool {
	return subject == "" || subject == "english" || subject == "unknown" || subject == "mixed"
}

func englishScope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, key := range []string{"subject", "subject_id"} {
			for _, subject := range r.URL.Query()[key] {
				if !englishSubject(subject) {
					fail(w, 400, "English 仅支持英语资料与练习")
					return
				}
			}
		}
		contentType := strings.ToLower(r.Header.Get("Content-Type"))
		if strings.HasPrefix(contentType, "multipart/form-data") {
			r.Body = http.MaxBytesReader(w, r.Body, 2*maxPDF+1024*1024)
			if r.ParseMultipartForm(1024*1024) != nil {
				fail(w, 400, "上传过大或格式无效")
				return
			}
			if r.MultipartForm != nil {
				defer r.MultipartForm.RemoveAll()
			}
			for _, key := range []string{"subject", "subject_id"} {
				for _, subject := range r.Form[key] {
					if !englishSubject(subject) {
						fail(w, 400, "English 仅支持英语资料与练习")
						return
					}
				}
			}
		} else if r.Body != nil && (r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH") {
			// Replay exactly the same bytes; endpoint decoders retain their smaller
			// individual limits and unknown-field/trailing-JSON checks.
			body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 2*1024*1024))
			if err != nil {
				fail(w, 413, "请求内容过大")
				return
			}
			_ = r.Body.Close()
			var payload any
			if json.Unmarshal(body, &payload) == nil && !englishSubjects(payload) {
				fail(w, 400, "English 仅支持英语资料与练习")
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
		}
		next.ServeHTTP(w, r)
	})
}
