---
# minimal — for someone who wants notes and dedup, and nothing else running.
#
# Turns off every background analysis: no code index, no git signals, no drift, no link
# checking, no weekly report, no dataset capture, no telemetry. What remains is the part
# that pays for itself on day one — recall (does this note already exist?), slugging,
# schema validation, and the note index.
#
# Good on a laptop where you resent every millisecond, on a vault with no associated
# code repo at all, and as a starting point: turn things back on one at a time when you
# want them, rather than turning them off after they surprise you.

static:
  code_index: false
  languages: []
  git_signals: false
  drift: {enabled: false}
  linkcheck: {enabled: false}
  logback: {knowledge_map: false, claude_md_fragment: false, inline_markers: false}

check:
  enabled: false
  reports: []

garden:
  enabled: false

dataset:
  enabled: false
  capture: []

telemetry:
  enabled: false

research:
  max_sources: 3
  use_docs_mcp: false
  scan_codebase: false
---
