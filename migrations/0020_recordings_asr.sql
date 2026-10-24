ALTER TABLE ai_models DROP CONSTRAINT ai_models_kind_check;
ALTER TABLE ai_models ADD CONSTRAINT ai_models_kind_check CHECK(kind IN ('chat','tts','asr'));
CREATE TABLE learning_recordings (
 id text PRIMARY KEY,user_id text NOT NULL REFERENCES users(id),training_id text NOT NULL REFERENCES training_sessions(id),
 task_index integer NOT NULL,mime text NOT NULL,audio bytea NOT NULL,hash text NOT NULL,request_id text NOT NULL,digest text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),UNIQUE(user_id,request_id)
);
CREATE TRIGGER immutable_recording BEFORE UPDATE OR DELETE ON learning_recordings FOR EACH ROW EXECUTE FUNCTION learning_immutable_blob();
CREATE TABLE learning_transcriptions (
 id text PRIMARY KEY,recording_id text NOT NULL REFERENCES learning_recordings(id),user_id text NOT NULL REFERENCES users(id),
 model_id bigint NOT NULL REFERENCES ai_models(id),model_version text NOT NULL,text text NOT NULL,
 request_id text NOT NULL,digest text NOT NULL,elapsed_ms integer NOT NULL,created_at timestamptz NOT NULL DEFAULT now(),UNIQUE(user_id,request_id)
);
CREATE TRIGGER immutable_transcription BEFORE UPDATE OR DELETE ON learning_transcriptions FOR EACH ROW EXECUTE FUNCTION learning_immutable_blob();

