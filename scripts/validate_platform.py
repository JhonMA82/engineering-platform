#!/usr/bin/env python3
from pathlib import Path
import json,re,sys
R=Path(__file__).resolve().parents[1]
required=['README.md','AGENTS.md','platform/catalog.json','platform/golden-paths.json','platform/feature-packs.json','docs/01-concepts/CONCEPTS_WITH_EXAMPLES.md','docs/13-examples/END_TO_END_SCHOOL_REQUESTS.md']
for f in required:
    if not (R/f).exists(): raise SystemExit(f'MISSING {f}')
for f in ['platform/catalog.json','platform/golden-paths.json','platform/feature-packs.json','skills/registry.json']:
    json.loads((R/f).read_text())
# local markdown links
rx=re.compile(r'\[[^\]]+\]\(([^)]+)\)')
bad=[]
for md in R.rglob('*.md'):
    for link in rx.findall(md.read_text(encoding='utf-8')):
        if link.startswith(('http://','https://','mailto:','#')): continue
        link=link.split('#',1)[0]
        if not link: continue
        if not (md.parent/link).resolve().exists(): bad.append(f'{md.relative_to(R)} -> {link}')
if bad: raise SystemExit('BROKEN LINKS\n'+'\n'.join(bad[:30]))
# junior coverage words
concept=(R/'docs/01-concepts/CONCEPTS_WITH_EXAMPLES.md').read_text()
for term in ['Golden Path','Feature Pack','Project Manifest','Harness','Skill','Guard','Quality Gate','Canonical Example','Eval','ADR','Knowledge Entry','Upgrade Recipe','Design System']:
    if term not in concept: raise SystemExit(f'MISSING CONCEPT {term}')
print(f'OK: {sum(1 for _ in R.rglob("*.md"))} markdown docs, examples and links valid.')
