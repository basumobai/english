"use client";
import { useEffect, useRef, useState } from "react";
import { api, requestID } from "../../lib/api";
const steps: Record<string, string> = {
  not_understood:
    "先用短句低速重听，核对音与词，再遮住原文听一遍；最后换新材料检验。",
  heard_forgot:
    "先不重听，用关键词复述人物、时间、地点和转折；再重听补齐遗漏，下一篇先做简短笔记。",
};
export function ListeningReflection({ trainingId }: { trainingId: string }) {
  const [reason, setReason] = useState("not_understood"),
    [notes, setNotes] = useState(""),
    [next, setNext] = useState(""),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false);
  const pending = useRef<{
      request_id: string;
      reason: string;
      notes: string;
    } | null>(null),
    dirty = useRef(false),
    lock = useRef(false);
  useEffect(() => {
    let live = true;
    api<{ reflection: { reason: string; notes: string } | null }>(
      `/learning/english/listening/${trainingId}/reflection`,
    )
      .then((d) => {
        if (live && !dirty.current && d.reflection) {
          setReason(d.reflection.reason);
          setNotes(d.reflection.notes);
          setNext(steps[d.reflection.reason]);
        }
      })
      .catch((e) => {
        if (live) setError(e.message);
      });
    return () => {
      live = false;
    };
  }, [trainingId]);
  async function save() {
    if (lock.current) return;
    lock.current = true;
    setBusy(true);
    setError("");
    try {
      if (!pending.current)
        pending.current = { request_id: requestID(), reason, notes };
      const d = await api<{ next_step: string }>(
        `/learning/english/listening/${trainingId}/reflection`,
        pending.current,
      );
      setNext(d.next_step);
      pending.current = null;
    } catch (e) {
      setError((e as Error).message);
    } finally {
      lock.current = false;
      setBusy(false);
    }
  }
  return (
    <section className="evidence-card listening-reflection">
      <h3>听力卡在哪一步？</h3>
      <label>
        这次更接近
        <select
          value={reason}
          disabled={busy}
          onChange={(e) => {
            dirty.current = true;
            pending.current = null;
            setReason(e.target.value);
            setNext("");
          }}
        >
          <option value="not_understood">声音和词没对应上：没听懂</option>
          <option value="heard_forgot">当时听懂了，做题时记不住</option>
        </select>
      </label>
      <label>
        先不重听，记下还记得的人物、时间、地点和转折
        <textarea
          value={notes}
          maxLength={1000}
          disabled={busy}
          onChange={(e) => {
            dirty.current = true;
            pending.current = null;
            setNotes(e.target.value);
          }}
        />
      </label>
      <button disabled={busy} onClick={() => void save()}>
        {busy ? "正在记录…" : "记录并选择训练步骤"}
      </button>
      {next && <p role="status">{next}</p>}
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      <p className="setting-note">
        本人反思会保存；用于选择练法，不直接折算长期能力。
      </p>
    </section>
  );
}
