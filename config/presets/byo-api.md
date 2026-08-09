---
# byo-api — for someone with an API key who wants research off their session.
#
# Research and verification go out to your own Anthropic key; intake, planning and
# synthesis stay on the host session. The point is not cost — it is that a six-source
# research run no longer eats the context window you are working in.
#
# You must export the key named in key_env before this preset does anything. The key
# itself is never stored in config; only the name of the variable holding it.
# Set a budget you are comfortable with: on_exhausted: queue defers the rest to
# tomorrow rather than failing the run.

engines:
  default: host
  api:
    provider: anthropic
    model: claude-sonnet-5
    key_env: ANTHROPIC_API_KEY
  local: {enabled: false}
  budget:
    advisor_usd_per_day: 0.00
    api_usd_per_day: 1.00
    on_exhausted: queue

pipeline:
  intake:     {engine: host}
  plan:       {engine: host}
  research:   {engine: api, fallback: host}
  synthesize: {engine: host}
  verify:     {engine: api, fallback: host}
  link:       {engine: none}

check:
  ai_pass: false
  drain_advisor_queue: false
---
