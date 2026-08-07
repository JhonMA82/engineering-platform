#!/usr/bin/env python3
from pathlib import Path
import json,re
ROOT=Path(__file__).resolve().parents[1]
REQ=['README.md','AGENTS.md','platform/catalog.json','platform/golden-paths.json','platform/feature-packs.json','golden-paths/README.md','feature-packs/README.md','harness/README.md','docs/07-quality/quality-gates.md']
def fail(m): print('ERROR:',m); raise SystemExit(1)
for x in REQ:
    if not (ROOT/x).exists(): fail('missing '+x)
for x in ['platform/catalog.json','platform/golden-paths.json','platform/feature-packs.json','platform/compatibility.json']:
    try: json.loads((ROOT/x).read_text())
    except Exception as e: fail(f'invalid JSON {x}: {e}')
c=json.loads((ROOT/'platform/catalog.json').read_text()); ids=[x['id'] for x in c['technology_catalog']]
if len(ids)!=len(set(ids)): fail('duplicate technology id')
g=json.loads((ROOT/'platform/golden-paths.json').read_text()); gp=[x['id'] for x in g['paths']]
if len(gp)!=len(set(gp)): fail('duplicate Golden Path id')
pat=re.compile(r'\[[^\]]+\]\(([^)]+)\)'); bad=[]
for md in ROOT.rglob('*.md'):
    for link in pat.findall(md.read_text()):
        if link.startswith(('http://','https://','#','mailto:')): continue
        link=link.split('#',1)[0]
        if link and not (md.parent/link).resolve().exists(): bad.append(f'{md.relative_to(ROOT)} -> {link}')
if bad: fail('broken links:\n'+'\n'.join(bad[:20]))
print(f'OK: {len(ids)} technologies, {len(gp)} Golden Paths, {len(list(ROOT.rglob("*.md")))} markdown files.')
