package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"

	"example.com/els/api/internal/platform"
	"example.com/els/api/internal/security"
)

func (s *server) createWordSets(w http.ResponseWriter, r *http.Request) {
	var p learningJobPayload
	if decodeLimit(r, &p, 8000) != nil || p.SenseID <= 0 || (p.Direction != "en_zh" && p.Direction != "zh_en") {
		fail(w, 400, "词义和方向无效")
		return
	}
	if _, err := platform.WordRules(p.Difficulty); err != nil {
		fail(w, 400, err.Error())
		return
	}
	if _, _, err := s.taskModels(r.Context(), "word_options", false); err != nil {
		fail(w, 400, err.Error())
		return
	}
	id, err := s.enqueueLearning(r.Context(), "word_options", fmt.Sprintf("%d:%s:%s:%d", p.SenseID, p.Direction, p.Difficulty, time.Now().UnixNano()), p)
	if err != nil {
		s.dbError(w, "queue words", err)
		return
	}
	respond(w, 202, map[string]string{"id": id})
}

type wordSet struct {
	Prompt    string            `json:"prompt"`
	Choices   []platform.Choice `json:"choices"`
	CorrectID string            `json:"correct_id"`
	Rationale string            `json:"rationale"`
}

func (s *server) generateWordSet(ctx context.Context, p learningJobPayload) (any, error) {
	var word, pos, meaning string
	err := s.db.QueryRowContext(ctx, `SELECT w.lemma,ws.part_of_speech,ws.meaning_zh FROM word_senses ws JOIN words w ON w.id=ws.word_id WHERE ws.id=$1`, p.SenseID).Scan(&word, &pos, &meaning)
	if err != nil {
		return nil, err
	}
	models, _, err := s.taskModels(ctx, "word_options", false)
	if err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(map[string]any{"word": word, "part_of_speech": pos, "meaning": meaning, "difficulty": p.Difficulty, "direction": p.Direction})
	text, err := s.learningCompletion(ctx, models[0], s.taskPrompt(ctx, "word_options")+` 返回{"prompt":"题面，可加短语境以消除多义歧义","choices":[{"id":"A","text":"选项正文","pos":"v.","similar_group":"g1"}],"correct_id":"A","rationale":"为何只有一个答案及选项相似性理由"}。选项ID固定A/B/C/D。相似组：普通四组不同；困难两项一组（含正确项），其余分别独立成组；绝望四项同一组。zh_en正确text必须逐字等于word，en_zh正确text必须逐字等于meaning。词性单独放pos字段，不混入text。`, string(raw))
	if err != nil {
		return nil, err
	}
	var set wordSet
	if err = platform.Decode(text, &set); err != nil {
		return nil, err
	}
	if err = platform.ValidateWord(p.Difficulty, set.CorrectID, set.Choices); err != nil {
		return nil, err
	}
	if set.Prompt == "" || len([]rune(set.Prompt)) > 800 || set.Rationale == "" {
		return nil, fmt.Errorf("词汇题面或审核理由无效")
	}
	for _, c := range set.Choices {
		if c.ID == set.CorrectID && ((p.Direction == "en_zh" && c.Text != meaning) || (p.Direction == "zh_en" && c.Text != word)) {
			return nil, fmt.Errorf("正确答案与词库不一致")
		}
	}
	id := security.Token(18)
	choices, _ := json.Marshal(set.Choices)
	_, err = s.learningExec(ctx, `INSERT INTO learning_word_sets(id,sense_id,direction,difficulty,prompt,choices,correct_id,rationale) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, id, p.SenseID, p.Direction, p.Difficulty, set.Prompt, string(choices), set.CorrectID, set.Rationale)
	return map[string]string{"id": id}, err
}
func (s *server) adminWordSets(w http.ResponseWriter, r *http.Request) {
	var raw json.RawMessage
	err := s.db.QueryRowContext(r.Context(), `SELECT COALESCE(jsonb_agg(row_to_json(t)),'[]') FROM (SELECT s.*,w.lemma FROM learning_word_sets s JOIN word_senses ws ON ws.id=s.sense_id JOIN words w ON w.id=ws.word_id ORDER BY s.created_at DESC LIMIT 100) t`).Scan(&raw)
	if err != nil {
		s.dbError(w, "word sets", err)
		return
	}
	respond(w, 200, map[string]any{"sets": raw})
}
func (s *server) approveWordSet(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Approved bool   `json:"approved"`
		Password string `json:"admin_password"`
	}
	if decode(r, &in) != nil {
		fail(w, 400, "请求无效")
		return
	}
	if !s.adminSession(w, r) {
		return
	}
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		s.dbError(w, "word approval", err)
		return
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(r.Context(), `UPDATE learning_word_sets SET approved=$2 WHERE id=$1`, r.PathValue("id"), in.Approved)
	if err == nil {
		n, _ := res.RowsAffected()
		if n == 0 {
			fail(w, 404, "选项集不存在")
			return
		}
		err = audit(tx, r, "approve_word_set", r.PathValue("id"))
	}
	if err == nil {
		err = tx.Commit()
	}
	if err != nil {
		s.dbError(w, "approve words", err)
		return
	}
	respond(w, 200, map[string]bool{"saved": true})
}
func wordRoundPublic(id, prompt, difficulty, direction string, choices []platform.Choice, started, deadline time.Time, answered *time.Time, correct *bool, picked *string, correctID string) map[string]any {
	public := []map[string]string{}
	for _, c := range choices {
		p := map[string]string{"id": c.ID, "text": c.Text}
		if difficulty == "despair" {
			p["pos"] = c.POS
		}
		public = append(public, p)
	}
	result := map[string]any{"id": id, "prompt": prompt, "difficulty": difficulty, "direction": direction, "choices": public, "started_at": started, "deadline": deadline, "server_time": time.Now(), "answered_at": answered, "duration_ms": deadline.Sub(started).Milliseconds()}
	if answered != nil {
		result["correct"] = correct
		result["picked"] = picked
		result["correct_id"] = correctID
	}
	return result
}
func (s *server) loadWordRound(ctx context.Context, id string) (map[string]any, error) {
	var prompt, difficulty, direction, correctID, word, phonetic, pos, meaning, entryKind, rationale string
	var choices []byte
	var started, deadline time.Time
	var answered *time.Time
	var correct *bool
	var picked *string
	err := s.db.QueryRowContext(ctx, `SELECT s.prompt,s.difficulty,s.direction,r.choice_order,r.started_at,r.deadline,r.answered_at,r.correct,r.picked,s.correct_id,w.lemma,ws.phonetic,ws.part_of_speech,ws.meaning_zh,w.entry_kind,s.rationale FROM learning_word_rounds r JOIN learning_word_sets s ON s.id=r.set_id JOIN word_senses ws ON ws.id=s.sense_id JOIN words w ON w.id=ws.word_id WHERE r.id=$1 AND r.user_id=$2`, id, userID(ctx)).Scan(&prompt, &difficulty, &direction, &choices, &started, &deadline, &answered, &correct, &picked, &correctID, &word, &phonetic, &pos, &meaning, &entryKind, &rationale)
	var cs []platform.Choice
	_ = json.Unmarshal(choices, &cs)
	out := wordRoundPublic(id, prompt, difficulty, direction, cs, started, deadline, answered, correct, picked, correctID)
	if answered != nil {
		out["vocabulary"] = map[string]string{"word": word, "phonetic": phonetic, "pos": pos, "chinese": meaning, "entry_kind": entryKind}
		out["rationale"] = rationale
	}
	return out, err
}
func (s *server) nextWordRound(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Direction  string `json:"direction"`
		Difficulty string `json:"difficulty"`
		RequestID  string `json:"request_id"`
		Seconds    int    `json:"seconds"`
		Book       string `json:"book_id"`
		Unit       int    `json:"unit"`
		Lesson     string `json:"lesson"`
		Starred    *bool  `json:"starred"`
		Kind       string `json:"entry_kind"`
	}
	if decode(r, &in) != nil || (in.Direction != "en_zh" && in.Direction != "zh_en") || !requestIDPattern.MatchString(in.RequestID) {
		fail(w, 400, "方向或请求编号无效")
		return
	}
	ms, err := platform.WordRules(in.Difficulty)
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	scope := textbookScope{Book: in.Book, Unit: in.Unit, Lesson: in.Lesson, Starred: in.Starred, Kind: in.Kind}
	if err := scope.validate(); err != nil {
		fail(w, 400, err.Error())
		return
	}
	if in.Seconds != 0 {
		if in.Seconds != 2 && in.Seconds != 3 && in.Seconds != 6 {
			fail(w, 400, "计时仅支持2、3或6秒")
			return
		}
		ms = in.Seconds * 1000
	}
	payload, _ := json.Marshal(in)
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		s.dbError(w, "word round", err)
		return
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(r.Context(), `SELECT pg_advisory_xact_lock(hashtextextended($1,38291))`, userID(r.Context())); err != nil {
		s.dbError(w, "word lock", err)
		return
	}
	var old string
	var prior []byte
	err = tx.QueryRowContext(r.Context(), `SELECT id,request_payload FROM learning_word_rounds WHERE user_id=$1 AND request_id=$2`, userID(r.Context()), in.RequestID).Scan(&old, &prior)
	if err == nil {
		var previous any
		_ = json.Unmarshal(prior, &previous)
		var current any
		_ = json.Unmarshal(payload, &current)
		if string(prior) != "{}" && payloadDigest(previous) != payloadDigest(current) {
			fail(w, 409, "请求编号已用于另一组挑战设置")
			return
		}
		tx.Rollback()
		out, e := s.loadWordRound(r.Context(), old)
		if e != nil {
			s.dbError(w, "word replay", e)
			return
		}
		respond(w, 200, out)
		return
	}
	if err != sql.ErrNoRows {
		s.dbError(w, "word replay", err)
		return
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE learning_word_rounds SET answered_at=deadline,correct=false,picked='' WHERE user_id=$1 AND answered_at IS NULL AND deadline<=clock_timestamp()`, userID(r.Context()))
	if err != nil {
		s.dbError(w, "expire round", err)
		return
	}
	err = tx.QueryRowContext(r.Context(), `SELECT id FROM learning_word_rounds WHERE user_id=$1 AND answered_at IS NULL`, userID(r.Context())).Scan(&old)
	if err == nil {
		tx.Commit()
		out, e := s.loadWordRound(r.Context(), old)
		if e != nil {
			s.dbError(w, "active word", e)
			return
		}
		respond(w, 200, out)
		return
	}
	if err != sql.ErrNoRows {
		s.dbError(w, "active round", err)
		return
	}
	var setID, prompt string
	var raw []byte
	err = tx.QueryRowContext(r.Context(), `SELECT ws.id,ws.prompt,ws.choices FROM learning_word_sets ws JOIN word_senses s ON s.id=ws.sense_id JOIN words w ON w.id=s.word_id WHERE ws.approved AND ws.direction=$1 AND ws.difficulty=$2 AND NOT EXISTS(SELECT 1 FROM suspended_senses x WHERE x.user_id=$3 AND x.sense_id=ws.sense_id)`+textbookSenseFilter(4)+` ORDER BY (ws.id=COALESCE((SELECT set_id FROM learning_word_rounds WHERE user_id=$3 ORDER BY started_at DESC LIMIT 1),'')),COALESCE((SELECT correct FROM learning_word_rounds WHERE user_id=$3 AND set_id=ws.id ORDER BY started_at DESC LIMIT 1),true),(SELECT count(*) FROM learning_word_rounds r WHERE r.user_id=$3 AND r.set_id=ws.id),random() LIMIT 1`, in.Direction, in.Difficulty, userID(r.Context()), scope.Book, scope.Unit, scope.Lesson, scope.Starred, scope.Kind).Scan(&setID, &prompt, &raw)
	if err == sql.ErrNoRows {
		fail(w, 404, "当前教材范围、词条类型和难度暂无已审核选项集；可放宽范围，或请管理员生成并审核")
		return
	}
	if err != nil {
		s.dbError(w, "pick word", err)
		return
	}
	var choices []platform.Choice
	if json.Unmarshal(raw, &choices) != nil {
		fail(w, 500, "选项数据无效")
		return
	}
	rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
	raw, _ = json.Marshal(choices)
	id := security.Token(18)
	_, err = tx.ExecContext(r.Context(), `INSERT INTO learning_word_rounds(id,user_id,set_id,request_id,choice_order,started_at,deadline,request_payload) VALUES($1,$2,$3,$4,$5,clock_timestamp(),clock_timestamp()+make_interval(secs=>$6::double precision/1000),$7)`, id, userID(r.Context()), setID, in.RequestID, string(raw), ms, string(payload))
	if err == nil {
		err = tx.Commit()
	}
	if err != nil {
		s.dbError(w, "start word", err)
		return
	}
	out, err := s.loadWordRound(r.Context(), id)
	if err != nil {
		s.dbError(w, "load word", err)
		return
	}
	respond(w, 201, out)
}
func (s *server) answerWordRound(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Picked string `json:"picked"`
	}
	if decode(r, &in) != nil || len(in.Picked) > 20 {
		fail(w, 400, "选项无效")
		return
	}
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		s.dbError(w, "word answer", err)
		return
	}
	defer tx.Rollback()
	var choices []byte
	var correctID string
	var ended *time.Time
	var late bool
	err = tx.QueryRowContext(r.Context(), `SELECT r.choice_order,s.correct_id,r.answered_at,r.deadline<=clock_timestamp() FROM learning_word_rounds r JOIN learning_word_sets s ON s.id=r.set_id WHERE r.id=$1 AND r.user_id=$2 FOR UPDATE OF r`, r.PathValue("id"), userID(r.Context())).Scan(&choices, &correctID, &ended, &late)
	if err == sql.ErrNoRows {
		fail(w, 404, "题目不存在")
		return
	}
	if err != nil {
		s.dbError(w, "lock word", err)
		return
	}
	if ended == nil {
		var cs []platform.Choice
		_ = json.Unmarshal(choices, &cs)
		valid := in.Picked == ""
		for _, v := range cs {
			if v.ID == in.Picked {
				valid = true
			}
		}
		if !valid {
			fail(w, 400, "选项不存在")
			return
		}
		if late {
			in.Picked = ""
		}
		_, err = tx.ExecContext(r.Context(), `UPDATE learning_word_rounds SET picked=$3,correct=$4,answered_at=clock_timestamp() WHERE id=$1 AND user_id=$2`, r.PathValue("id"), userID(r.Context()), in.Picked, !late && in.Picked == correctID)
	}
	if err == nil {
		err = tx.Commit()
	}
	if err != nil {
		s.dbError(w, "save word", err)
		return
	}
	out, err := s.loadWordRound(r.Context(), r.PathValue("id"))
	if err != nil {
		s.dbError(w, "word result", err)
		return
	}
	respond(w, 200, out)
}
