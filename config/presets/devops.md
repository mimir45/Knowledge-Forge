---
# devops — for someone whose "code" is mostly YAML, Terraform and shell.
#
# Read the tradeoff before picking this: `pkg/codeindex` ships tree-sitter grammars for
# Java and TypeScript only, so AST-level drift does not apply to a Helm chart or a
# Terraform module. `code_index: false` says so honestly rather than pretending.
#
# What still works, and is most of the value here: git signals (churn, ownership,
# co-change coupling), file-level drift on cited paths, the link graph, duplicate
# detection, and the incident note type. A note citing `infra/prod/values.yaml` is still
# demoted when that file changes — it is the *shape* of the change that goes unanalysed.
#
# incident notes never expire on a timer: a postmortem is a record of what happened, and
# what happened does not go stale. decision notes likewise — they get superseded.

static:
  code_index: false
  languages: []
  git_signals: true
  drift: {enabled: true, auto_repair_line_numbers: false}

freshness_days:
  howto: 180
  api: 90
  incident: 0
  decision: 0

research:
  prefer: [official-docs, source-code, rfc]

verify:
  run_code: ask            # a devops snippet that "just runs" is how prod gets changed

write:
  diagrams: ascii          # renders in a terminal and in a runbook
---
