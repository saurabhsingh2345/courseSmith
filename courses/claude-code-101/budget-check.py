import json,sys,re,glob,os
WPM=174
IDEAL={'session':30,'doc':28,'wiring':26,'waypoint':20,'titlecard':18}
CEIL={'session':9,'doc':8,'wiring':8,'waypoint':3,'titlecard':3}
HEDGE=re.compile(r'(?i)\b(typically|roughly|generally|usually|often|depending on|in many cases|may vary|can vary|tends? to|approximately|somewhat|relatively)\b')
def bounds(tw,ceil,ideal):
    if tw<=0: return 2,ceil
    sug=max((tw+ideal//2)//ideal,2); sug=max(sug,-(-tw//60)); sug=min(sug,ceil)
    mn=min(max(sug-1,2),ceil); mx=min(max(sug+2,mn),ceil)
    return mn,mx
tot_t=0; tot_w=0; bad=0
files=sys.argv[1:]
segs=[]
for f in files:
    segs+=json.load(open(f))
for s in segs:
    t=s['target_sec']; tpl=s['template']; beats=s['plan']['beats']
    w=sum(len(b['narration'].split()) for b in beats)
    tw=t*WPM//60; lo,hi=tw*75//100,tw*155//100
    mn,mx=bounds(tw,CEIL[tpl],IDEAL[tpl])
    tot_t+=t; tot_w+=w
    flags=[]
    if not (lo<=w<=hi): flags.append(f'WORDS {w} not in {lo}-{hi} (aim {tw})')
    if not (mn<=len(beats)<=mx): flags.append(f'BEATS {len(beats)} not in {mn}-{mx}')
    for b in beats:
        n=len(b['narration'].split())
        floor=6 if b is beats[0] or b is beats[-1] else 10
        if n>60: flags.append(f'beat {b["id"]} {n}w >60')
        if n<floor: flags.append(f'beat {b["id"]} {n}w <{floor}')
        if ';' in b['narration']: flags.append(f'beat {b["id"]} SEMICOLON')
        if HEDGE.search(b['narration']): flags.append(f'beat {b["id"]} HEDGE: {HEDGE.search(b["narration"]).group()}')
    if flags:
        bad+=1
        print(f'✗ {s["id"]:14} {tpl:10} {t:4}s  '+' | '.join(flags))
    else:
        print(f'✓ {s["id"]:14} {tpl:10} {t:4}s  {w:4}w  {len(beats)} beats')
print(f'\n{len(segs)} segments · target {tot_t}s ({tot_t//60}m{tot_t%60:02d}s) · {tot_w} words · {bad} with problems')
