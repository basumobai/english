"""Validate all shipped textbook records and CSV exports, without network."""
from pathlib import Path
import csv, hashlib, json

ROOT = Path(__file__).resolve().parents[1]
EXPORT = ROOT / 'data/vocabulary'
manifest = json.loads((EXPORT / 'manifest.json').read_text())
seen = set()
total = stars = 0
ranges = {'c1': range(1,4), 'c2': range(4,7), 'c3': range(7,10), 's1': range(1,4), 's2': range(4,7), 's3': range(7,10), 's4': range(10,13)}
assert [b['id'] for b in manifest['books']] == list(ranges)
for summary in manifest['books']:
    bid = summary['id']
    raw = (ROOT/'services/api/internal/textbook/data'/f'{bid}.json').read_bytes()
    assert hashlib.sha256(raw).hexdigest() == summary['sha256'], f'{bid}: checksum'
    entries = json.loads(raw)['entries']
    assert len(entries) == summary['entries']
    assert {e['unit'] for e in entries} == set(ranges[bid])
    with (EXPORT/f'{bid}.csv').open(encoding='utf-8-sig', newline='') as f:
        rows = list(csv.DictReader(f))
    assert len(rows) == len(entries)
    for i, (e,r) in enumerate(zip(entries, rows), 1):
        assert e['id'] not in seen
        seen.add(e['id'])
        assert e['position'] == i
        assert all(isinstance(e[k],str) for k in ('word','phonetic','pos','chinese','lesson','source_url'))
        assert e['word'] and e['chinese'] and e['lesson']
        assert isinstance(e['starred'],bool) and e['source_page'] >= 0
        assert e['source_url'].startswith('https://')
        assert r['word'] == ('*' if e['starred'] else '') + e['word']
        for k in ('id','unit','unit_title','lesson','phonetic','pos','chinese','source_url','source_page','starred'):
            assert r[k] == str(e[k]), (bid,i,k)
    assert sum(e['starred'] for e in entries) == summary['starred']
    assert sum(e['source_page']==0 for e in entries) == summary['secondary_only']
    for u in summary['units']:
        selected = [e for e in entries if e['unit']==u['unit']]
        assert len(selected) == u['entries']
        assert {l:sum(e['lesson']==l for e in selected) for l in u['lessons']} == u['lessons']
    total += len(entries)
    stars += summary['starred']
assert total == manifest['total_entries']
print(f'Corpus OK: 7 books, 21 units, {total} occurrences, {stars} starred; JSON/CSV and checksums agree.')
