#!/usr/bin/env python3
"""Print the documentation body for a Freshdesk API anchor."""
import re, html, sys

D = "/private/tmp/claude-501/-Users-leca-src-terraform-provider-freshdesk/fd80c7ec-4f49-416a-8a36-46a76a725dd3/scratchpad/docs"
S = open(f"{D}/api.html", encoding="utf-8", errors="replace").read()
S = re.sub(r'<(script|style)[^>]*>.*?</\1>', '', S, flags=re.S)

# Content sections are <section id="anchor"> or <div id="anchor"> outside the nav.
# Content sections are <div id='anchor' class='scroll-spy'>; the nav uses
# double-quoted hrefs, so match the single-quoted scroll-spy form specifically.
ANCHORS = [(m.start(), m.group(1)) for m in
           re.finditer(r'''<div id=['"]([a-zA-Z0-9_-]+)['"] class=['"]scroll-spy['"]''', S)]

def clean(frag):
    frag = re.sub(r'</t[dh]>', ' | ', frag)
    frag = re.sub(r'</tr>', '\n', frag)
    frag = re.sub(r'</(p|div|li|h[1-6]|pre|section)>', '\n', frag)
    t = html.unescape(re.sub(r'<[^>]+>', '', frag))
    t = re.sub(r'[ \t]+', ' ', t)
    t = re.sub(r'\n *', '\n', t)
    return re.sub(r'\n{3,}', '\n\n', t).strip()

def body(anchor, limit=14000):
    hits = [i for i, (p, a) in enumerate(ANCHORS) if a == anchor]
    if not hits:
        return f"!! anchor #{anchor} not found"
    # Prefer the occurrence with the most content after it (skips nav stubs).
    best, bestlen = None, -1
    for i in hits:
        start = ANCHORS[i][0]
        end = ANCHORS[i + 1][0] if i + 1 < len(ANCHORS) else len(S)
        # extend across following anchors until the next *documented* section
        if end - start > bestlen:
            best, bestlen = (start, end), end - start
    return clean(S[best[0]:best[1]])[:limit]

if __name__ == "__main__":
    for a in sys.argv[1:]:
        print(f"\n{'='*78}\n#{a}\n{'='*78}")
        print(body(a))
