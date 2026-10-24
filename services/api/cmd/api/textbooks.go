package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"example.com/els/api/internal/textbook"
)

func (s *server) syncTextbooks(ctx context.Context) error {
	books, err := textbook.Load()
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(19283749)`); err != nil {
		return err
	}
	for _, book := range books {
		var old string
		if err = tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT dataset_hash FROM textbook_books WHERE id=$1),'')`, book.ID).Scan(&old); err != nil {
			return err
		}
		if old == book.Hash {
			continue
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO textbook_books(id,title,kind,volume,edition,entry_count,dataset_hash,source_url) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(id) DO UPDATE SET title=EXCLUDED.title,entry_count=EXCLUDED.entry_count,dataset_hash=EXCLUDED.dataset_hash,source_url=EXCLUDED.source_url`, book.ID, book.Title, book.Kind, book.Volume, book.Edition, len(book.Entries), book.Hash, book.SourceURL)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM textbook_entries WHERE book_id=$1`, book.ID); err != nil {
			return err
		}
		for _, e := range book.Entries {
			var word, sense int64
			err = tx.QueryRowContext(ctx, `INSERT INTO words(lemma) VALUES($1) ON CONFLICT((lower(lemma))) DO UPDATE SET lemma=words.lemma RETURNING id`, e.Word).Scan(&word)
			if err != nil {
				return err
			}
			err = tx.QueryRowContext(ctx, `INSERT INTO word_senses(word_id,part_of_speech,meaning_zh,phonetic) VALUES($1,$2,$3,$4) ON CONFLICT(word_id,part_of_speech,meaning_zh) DO UPDATE SET phonetic=CASE WHEN word_senses.phonetic='' THEN EXCLUDED.phonetic ELSE word_senses.phonetic END RETURNING id`, word, e.POS, e.Chinese, e.Phonetic).Scan(&sense)
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO textbook_entries(id,book_id,sense_id,unit,unit_title,lesson,position,word,phonetic,part_of_speech,meaning_zh,starred,source_url,source_page) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`, e.ID, book.ID, sense, e.Unit, e.UnitTitle, e.Lesson, e.Position, e.Word, e.Phonetic, e.POS, e.Chinese, e.Starred, e.SourceURL, e.SourcePage)
			if err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

type textbookScope struct {
	Book    string
	Unit    int
	Lesson  string
	Starred *bool
}

func (f textbookScope) validate() error {
	if f.Book != "" && map[string]int{"c1": 1, "c2": 4, "c3": 7, "s1": 1, "s2": 4, "s3": 7, "s4": 10}[f.Book] == 0 {
		return fmt.Errorf("教材册次无效")
	}
	if f.Unit < 0 || f.Unit > 12 || f.Unit > 0 && f.Book == "" || f.Lesson != "" && f.Unit == 0 || len([]rune(f.Lesson)) > 80 {
		return fmt.Errorf("请按册次、Unit、课节选择范围")
	}
	if f.Unit > 0 {
		start := map[string]int{"c1": 1, "c2": 4, "c3": 7, "s1": 1, "s2": 4, "s3": 7, "s4": 10}[f.Book]
		if f.Unit < start || f.Unit > start+2 {
			return fmt.Errorf("Unit 不属于所选册次")
		}
	}
	return nil
}
func scopeQuery(r *http.Request) (textbookScope, error) {
	q := r.URL.Query()
	f := textbookScope{Book: q.Get("book_id"), Lesson: q.Get("lesson")}
	if q.Get("unit") != "" {
		n, e := strconv.Atoi(q.Get("unit"))
		if e != nil {
			return f, fmt.Errorf("Unit 无效")
		}
		f.Unit = n
	}
	if q.Get("starred") != "" {
		b, e := strconv.ParseBool(q.Get("starred"))
		if e != nil {
			return f, fmt.Errorf("星号筛选无效")
		}
		f.Starred = &b
	}
	return f, f.validate()
}

// An EXISTS filter preserves one candidate per sense even when a word occurs
// in several lessons. All supplied scope fields must match one occurrence.
func textbookSenseFilter(first int) string {
	return fmt.Sprintf(` AND (($%d='' AND $%d=0 AND $%d='' AND $%d::boolean IS NULL) OR EXISTS(SELECT 1 FROM textbook_entries te WHERE te.sense_id=s.id AND ($%d='' OR te.book_id=$%d) AND ($%d=0 OR te.unit=$%d) AND ($%d='' OR te.lesson=$%d) AND ($%d::boolean IS NULL OR te.starred=$%d)))`, first, first+1, first+2, first+3, first, first, first+1, first+1, first+2, first+2, first+3, first+3)
}
func (s *server) textbookCatalog(w http.ResponseWriter, r *http.Request) {
	var raw json.RawMessage
	err := s.db.QueryRowContext(r.Context(), `SELECT COALESCE(jsonb_agg(row_to_json(b) ORDER BY b.id),'[]') FROM (SELECT id,title,kind,volume,edition,entry_count,source_url,(SELECT COALESCE(jsonb_agg(row_to_json(u) ORDER BY u.unit),'[]') FROM (SELECT unit,min(unit_title) AS title,count(*) AS entry_count,jsonb_agg(DISTINCT lesson) AS lessons FROM textbook_entries WHERE book_id=tb.id GROUP BY unit) u) AS units FROM textbook_books tb) b`).Scan(&raw)
	if err != nil {
		s.dbError(w, "textbook catalog", err)
		return
	}
	respond(w, 200, map[string]any{"books": raw, "edition": "北师大版2019"})
}
