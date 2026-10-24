-- Book occurrences and learned word senses have separate identities. Repeated
-- words across books/lessons keep every source location and one FSRS history.
CREATE TABLE textbook_books (
 id text PRIMARY KEY, title text NOT NULL, kind text NOT NULL,
 volume integer NOT NULL CHECK(volume BETWEEN 1 AND 4), edition text NOT NULL,
 entry_count integer NOT NULL, dataset_hash text NOT NULL, source_url text NOT NULL
);
CREATE TABLE textbook_entries (
 id text PRIMARY KEY, book_id text NOT NULL REFERENCES textbook_books(id),
 sense_id bigint NOT NULL REFERENCES word_senses(id),
 unit integer NOT NULL CHECK(unit BETWEEN 1 AND 12), unit_title text NOT NULL,
 lesson text NOT NULL, position integer NOT NULL,
 word text NOT NULL, phonetic text NOT NULL, part_of_speech text NOT NULL,
 meaning_zh text NOT NULL, starred boolean NOT NULL,
 source_url text NOT NULL, source_page integer NOT NULL,
 UNIQUE(book_id,position)
);
CREATE INDEX textbook_scope ON textbook_entries(book_id,unit,lesson,position);
CREATE INDEX textbook_sense ON textbook_entries(sense_id,book_id,unit,lesson,starred);
