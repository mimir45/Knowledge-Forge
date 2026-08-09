---
# claude-only — the default, and the right answer for almost everyone.
#
# For someone running inside Claude Code who does not want a second bill. Every stage
# that needs a model uses the session you are already paying for (`engine: host`); no
# API key, no per-day budget, nothing to configure. The static core still runs locally.
#
# The tradeoff: work happens in your session's context window, so a long research run
# competes with whatever else you are doing. If that starts to bite, move research to
# byo-api.

engines:
  default: host
  api: {provider: "", model: "", key_env: ""}
  local: {enabled: false}
  budget: {advisor_usd_per_day: 0.00, api_usd_per_day: 0.00}

pipeline:
  intake:     {engine: host}
  plan:       {engine: host}
  research:   {engine: host}
  synthesize: {engine: host}
  verify:     {engine: host}
  link:       {engine: none}

check:
  ai_pass: false
  drain_advisor_queue: false
---
