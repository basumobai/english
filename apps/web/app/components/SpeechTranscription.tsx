"use client";
import { useEffect, useRef, useState } from "react";
import { api, requestID, upload, type Model } from "../../lib/api";
type Saved = { id: string; transcript: string | null };
export function SpeechTranscription({
  trainingId,
  taskIndex,
  onTranscript,
  recordingURL,
  readOnly = false,
}: {
  trainingId: string;
  taskIndex: number;
  onTranscript: (text: string) => void;
  recordingURL?: string;
  readOnly?: boolean;
}) {
  const [saved, setSaved] = useState<Saved[]>([]),
    [models, setModels] = useState<Model[]>([]),
    [model, setModel] = useState(0),
    [file, setFile] = useState<File | null>(null),
    [busy, setBusy] = useState(false),
    [status, setStatus] = useState("");
  const active = useRef(true);
  useEffect(() => {
    active.current = true;
    return () => {
      active.current = false;
    };
  }, []);
  const uploadKey = useRef(requestID()),
    recording = useRef(""),
    transcriptionKeys = useRef(new Map<string, string>()),
    lock = useRef(false);
  async function history() {
    const d = await api<{ recordings: Saved[] }>(
      `/learning/recordings?training_id=${encodeURIComponent(trainingId)}&task_index=${taskIndex}`,
    );
    setSaved(d.recordings);
  }
  useEffect(() => {
    let live = true;
    api<{ recordings: Saved[] }>(
      `/learning/recordings?training_id=${encodeURIComponent(trainingId)}&task_index=${taskIndex}`,
    )
      .then((d) => {
        if (live) setSaved(d.recordings);
      })
      .catch((e) => {
        if (live) setStatus(e.message);
      });
    return () => {
      live = false;
    };
  }, [trainingId, taskIndex]);
  useEffect(() => {
    if (readOnly) return;
    let live = true;
    api<{ models: Model[] }>("/ai/models")
      .then((d) => {
        if (live) {
          const ms = d.models.filter((m) => m.kind === "asr");
          setModels(ms);
          setModel(ms[0]?.id || 0);
        }
      })
      .catch((e) => {
        if (live) setStatus(e.message);
      });
    return () => {
      live = false;
    };
  }, [readOnly]);
  function reset() {
    uploadKey.current = requestID();
    recording.current = "";
    setStatus("");
  }
  useEffect(() => {
    reset();
    setFile(null);
  }, [taskIndex, recordingURL]);
  function useText(text: string) {
    if (!active.current) return;
    if (Array.from(text).length > 4000) {
      setStatus(
        "转写超过单项4000字符限制，请从已保存全文中核对并手动选取作答内容。",
      );
      return;
    }
    onTranscript(text);
  }
  async function transcribe(id: string) {
    const key = `${id}:${model}`;
    if (!transcriptionKeys.current.has(key))
      transcriptionKeys.current.set(key, requestID());
    const result = await api<{ text: string; note: string }>(
      `/learning/recordings/${id}/transcribe`,
      { request_id: transcriptionKeys.current.get(key), model_id: model },
    );
    setStatus(result.note);
    useText(result.text);
    await history();
  }
  async function act(fn: () => Promise<void>) {
    if (lock.current) return;
    lock.current = true;
    setBusy(true);
    setStatus("");
    try {
      await fn();
    } catch (e) {
      setStatus((e as Error).message);
    } finally {
      lock.current = false;
      setBusy(false);
    }
  }
  async function save(withTranscription: boolean) {
    if (!recording.current) {
      const audio =
        file ||
        new File([await (await fetch(recordingURL!)).blob()], "recording.webm");
      if (audio.size > 20 * 1024 * 1024) throw Error("录音应小于20MB");
      const form = new FormData();
      form.append("file", audio);
      form.append("training_id", trainingId);
      form.append("task_index", String(taskIndex));
      form.append("request_id", uploadKey.current);
      const d = await upload<{ id: string }>("/learning/recordings", form);
      recording.current = d.id;
    }
    if (withTranscription) await transcribe(recording.current);
    else {
      setStatus("录音已保存到本人账户。");
      await history();
    }
  }
  return (
    <details className="evidence-card speech-transcription">
      <summary>保存录音并自动转写（可选）</summary>
      <p>主动保存后可跨刷新回放。自动转写需要自己核对。</p>
      {!readOnly && (
        <>
          <label>
            ASR模型
            <select
              value={model}
              disabled={busy}
              onChange={(e) => setModel(Number(e.target.value))}
            >
              {!models.length && <option value={0}>暂无授权ASR模型</option>}
              {models.map((m) => (
                <option key={m.id} value={m.id}>
                  {m.name}
                </option>
              ))}
            </select>
          </label>
          {!models.length && (
            <p>
              可以保存和回放录音，或手动填写作答；管理员授权ASR后可自动转写。
            </p>
          )}
          <label>
            或选择音频（小于20MB）
            <input
              type="file"
              accept="audio/*"
              disabled={busy}
              onChange={(e) => {
                setFile(e.target.files?.[0] || null);
                reset();
              }}
            />
          </label>
          <div className="button-row">
            <button
              disabled={busy || (!file && !recordingURL)}
              onClick={() => void act(() => save(false))}
            >
              仅保存录音
            </button>
            <button
              disabled={busy || !model || (!file && !recordingURL)}
              onClick={() => void act(() => save(true))}
            >
              {busy ? "正在处理…" : "保存并转写到我的作答"}
            </button>
          </div>
        </>
      )}
      <p role="status">{status}</p>
      {saved.map((r) => (
        <div key={r.id}>
          <audio
            controls
            preload="none"
            src={`/api/v1/learning/recordings/${r.id}/audio`}
          />
          {r.transcript && (
            <>
              <p className="ai-explanation">{r.transcript}</p>
              {!readOnly && (
                <button disabled={busy} onClick={() => useText(r.transcript!)}>
                  使用已保存转写
                </button>
              )}
            </>
          )}
          {!readOnly && model > 0 && (
            <button
              disabled={busy}
              onClick={() => void act(() => transcribe(r.id))}
            >
              转写这段已保存录音
            </button>
          )}
        </div>
      ))}
      <p className="setting-note">
        转写与文字评价不作为官方听说或专业发音评分。
      </p>
    </details>
  );
}
