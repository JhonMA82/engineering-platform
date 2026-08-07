#!/usr/bin/env python3
import argparse,json
from datetime import date
from pathlib import Path
p=argparse.ArgumentParser(); p.add_argument('--name',required=True); p.add_argument('--golden-path',required=True); p.add_argument('--output',default='.engineering/project.json'); a=p.parse_args()
d={'schema_version':1,'platform_version':'0.1.0','project':{'name':a.name},'golden_path':{'id':a.golden_path,'version':'0.1.0'},'starters':{},'features':{},'decisions':[],'generated_at':str(date.today())}
o=Path(a.output); o.parent.mkdir(parents=True,exist_ok=True); o.write_text(json.dumps(d,indent=2)+'\n'); print(o)
