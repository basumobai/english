-- Derive the type from the headword, excluding abbreviation expansions.
-- Keep senses and all existing learning records intact.
ALTER TABLE words ADD COLUMN entry_kind text GENERATED ALWAYS AS
 (CASE WHEN regexp_replace(lemma, ' +\([^)]*\)$', '') ~ '[[:space:]]|\.{3}|…' THEN 'phrase' ELSE 'word' END) STORED;
CREATE INDEX words_entry_kind_idx ON words(entry_kind);
