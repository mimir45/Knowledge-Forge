---
# offline — for someone who wants the vault machinery and nothing that phones home.
#
# Every stage runs on the Go static core. No model calls, no network, no API key, no
# spend. You get: dedup and recall, drift detection, the nine reports, link graph,
# duplicate detection, slugging and validation. You do not get: research, synthesis,
# or verification — this preset will not write a note for you, it will tell you whether
# one already exists and whether the ones you have still match your code.
#
# Also the right choice on an air-gapped machine or a repo whose contents may not leave
# the building.

engines:
  default: none
  local: {enabled: false}
  budget: {advisor_usd_per_day: 0.00, api_usd_per_day: 0.00}

pipeline:
  intake:     {engine: none}
  plan:       {engine: none}
  research:   {engine: none}
  synthesize: {engine: none}
  verify:     {engine: none}
  link:       {engine: none}

research:
  use_docs_mcp: false

static:
  linkcheck: {enabled: false}   # the one part of the static core that makes requests

check:
  ai_pass: false
  drain_advisor_queue: false
---
