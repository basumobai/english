// Package textbook contains the versioned, offline textbook vocabulary corpus.
package textbook

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed data/*.json
var datasets embed.FS

type Entry struct {
	ID         string `json:"id"`
	Unit       int    `json:"unit"`
	UnitTitle  string `json:"unit_title"`
	Lesson     string `json:"lesson"`
	Position   int    `json:"position"`
	Word       string `json:"word"`
	Phonetic   string `json:"phonetic"`
	POS        string `json:"pos"`
	Chinese    string `json:"chinese"`
	Starred    bool   `json:"starred"`
	SourceURL  string `json:"source_url"`
	SourcePage int    `json:"source_page"`
}
type Book struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Kind      string  `json:"kind"`
	Volume    int     `json:"volume"`
	Edition   string  `json:"edition"`
	SourceURL string  `json:"source_url"`
	Entries   []Entry `json:"entries"`
	Hash      string  `json:"-"`
}

func Load() ([]Book, error) {
	books := []Book{}
	seen := map[string]bool{}
	for _, id := range []string{"c1", "c2", "c3", "s1", "s2", "s3", "s4"} {
		raw, err := datasets.ReadFile("data/" + id + ".json")
		if err != nil {
			return nil, err
		}
		var book Book
		if err = json.Unmarshal(raw, &book); err != nil {
			return nil, fmt.Errorf("%s: %w", id, err)
		}
		if book.ID != id || book.Title == "" || book.Edition != "北师大版2019" || len(book.Entries) == 0 {
			return nil, fmt.Errorf("invalid textbook %s", id)
		}
		units := map[int]bool{}
		for i, e := range book.Entries {
			if e.ID == "" || seen[e.ID] || e.Position != i+1 || strings.TrimSpace(e.Word) == "" || strings.TrimSpace(e.Chinese) == "" || e.Lesson == "" || e.UnitTitle == "" || e.SourceURL == "" || e.SourcePage < 0 || e.Unit < 1 || e.Unit > 12 {
				return nil, fmt.Errorf("invalid %s entry %d", id, i+1)
			}
			seen[e.ID] = true
			units[e.Unit] = true
		}
		if len(units) != 3 {
			return nil, fmt.Errorf("%s must have three units", id)
		}
		hash := sha256.Sum256(raw)
		book.Hash = hex.EncodeToString(hash[:])
		books = append(books, book)
	}
	return books, nil
}
