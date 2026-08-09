---
# max — all four tiers, for a vault whose notes are load-bearing.
#
# For someone whose notes get read by other people, or fed back into decisions months
# later, and who would rather pay a little to be told when a claim is wrong. Adds the
# advisor tier on top of byo-api: a stronger model critiques what was synthesised.
#
# The advisor is critique-only by contract — it returns disputed claims and a patch,
# never a rewrite. "Generate cheap, critique expensive."
#
# It fires selectively, not on every note: decisions and patterns, anything the writer
# marked low-confidence, anything touching security, auth or payments, and anything you
# ran with --deep. Budget it: two dollars a day is roughly a dozen critiques.

engines:
  default: host
  api:
    provider: anthropic
    model: claude-sonnet-5
    key_env: ANTHROPIC_API_KEY
  advisor:
    model: claude-opus-5
    mode: critique
  local: {enabled: false}
  budget:
    advisor_usd_per_day: 2.00
    api_usd_per_day: 1.00
    on_exhausted: queue
  routing:
    advisor_when:
      type: [decision, pattern]
      confidence_below: medium
      stack_in: [security, auth, payments]
      user_flag: "--deep"

pipeline:
  intake:     {engine: host}
  plan:       {engine: host}
  research:   {engine: api, fallback: host}
  synthesize: {engine: host}
  verify:     {engine: advisor, fallback: api, then: host}
  link:       {engine: none}

check:
  ai_pass: true
  drain_advisor_queue: true
---
