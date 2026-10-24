"use client";
import { useEffect, useState } from "react";
import { api } from "../../lib/api";

export type TextbookScope = { book_id: string; unit: number; lesson: string; starred: "" | "true" | "false" };
export const allTextbooks: TextbookScope = { book_id: "", unit: 0, lesson: "", starred: "" };
type Book = { id: string; title: string; kind: string; entry_count: number; units: {unit: number; title: string; entry_count: number; lessons: string[]}[] };
export function textbookQuery(scope: TextbookScope) {
  const q = new URLSearchParams();
  if (scope.book_id) q.set("book_id", scope.book_id);
  if (scope.unit) q.set("unit", String(scope.unit));
  if (scope.lesson) q.set("lesson", scope.lesson);
  if (scope.starred) q.set("starred", scope.starred);
  return q.toString();
}
export function textbookPayload(scope: TextbookScope) {
  return { book_id: scope.book_id, unit: scope.unit, lesson: scope.lesson, starred: scope.starred ? scope.starred === "true" : null };
}
export function TextbookScopeSelector({ value, onChange, disabled = false }: { value: TextbookScope; onChange: (v: TextbookScope) => void; disabled?: boolean }) {
  const [books, setBooks] = useState<Book[]>([]), [error, setError] = useState("");
  useEffect(() => { let active=true; api<{books:Book[]}>("/textbooks").then(d=>{if(active)setBooks(d.books)}).catch(e=>{if(active)setError(e.message)});return()=>{active=false} }, []);
  const book = books.find(b=>b.id===value.book_id), unit=book?.units.find(u=>u.unit===value.unit);
  return <fieldset className="textbook-scope" disabled={disabled}>
    <legend>教材词汇范围</legend>
    <div className="textbook-filters">
      <label>教材册次<select aria-label="教材册次" value={value.book_id} onChange={e=>onChange({...value,book_id:e.target.value,unit:0,lesson:""})}>
        <option value="">全部词汇（含自建词库）</option>
        {["必修","选择性必修"].map(kind=><optgroup key={kind} label={kind}>{books.filter(b=>b.kind===kind).map(b=><option key={b.id} value={b.id}>{b.title} · {b.entry_count} 条</option>)}</optgroup>)}
      </select></label>
      <label>教材 Unit<select aria-label="教材 Unit" value={value.unit} disabled={!book} onChange={e=>onChange({...value,unit:Number(e.target.value),lesson:""})}>
        <option value={0}>全部 Unit</option>{book?.units.map(u=><option key={u.unit} value={u.unit}>Unit {u.unit} · {u.title}</option>)}
      </select></label>
      <label>教材课节<select aria-label="教材课节" value={value.lesson} disabled={!unit} onChange={e=>onChange({...value,lesson:e.target.value})}>
        <option value="">全部课节</option>{unit?.lessons.map(l=><option key={l} value={l}>{l}</option>)}
      </select></label>
      <label>星号词<select aria-label="星号词" value={value.starred} onChange={e=>onChange({...value,starred:e.target.value as TextbookScope["starred"]})}>
        <option value="">全部（保留星号词）</option><option value="true">仅星号词 *</option><option value="false">仅非星号词</option>
      </select></label>
    </div>
    {error && <p className="error" role="alert">{error}</p>}
  </fieldset>
}
