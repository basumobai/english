ALTER TABLE learning_word_rounds ADD COLUMN request_payload jsonb NOT NULL DEFAULT '{}';
CREATE INDEX learning_word_rounds_user_set_started ON learning_word_rounds(user_id,set_id,started_at DESC);
CREATE INDEX learning_word_rounds_user_started ON learning_word_rounds(user_id,started_at DESC);
