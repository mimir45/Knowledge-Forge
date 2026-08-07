# Knowledge Forge B2B — "No Context Loss"

Third document. Read `KNOWLEDGE-FORGE-DESIGN.md` then `KNOWLEDGE-FORGE-ADDENDUM.md`
first — this builds directly on the engine tiers (§A) and the static core (§B).

> **Rev 2:** backend language is no longer assumed to be Spring Boot. The T0 core is
> Go (STACK ADR-001), and the server-side default **leans Go for direct library reuse
> of the core** — final call deferred to B1 (STACK ADR-002). Spring references below
> are retained as the alternative branch.

**The product in one line:** engineering knowledge is created in Slack threads, PR
reviews, incident calls and email, and it dies there. This ingests those sources,
extracts the durable parts into a validated wiki, and serves it back to every
developer's AI agent as an MCP server.

**Category name: "no context loss."** It's a good name — keep it. It describes a pain
every engineering manager recognizes without needing the word "AI" in it.

---

## 1. Why this is a different product, not a bigger version of the OSS one

| | OSS (single dev) | B2B |
|---|---|---|
| Input | one dev's questions | Slack, Teams, Jira, Confluence, email, PRs, incidents |
| Volume | ~50 notes/month | ~50k messages/month, ~0.5% durable |
| Retrieval | lexical over ~500 structured notes | hybrid over ~500k heterogeneous unstructured items |
| Hard problem | dedup | **noise ratio and permissions** |
| Buyer's pain | "I re-explain things to my agent" | "we lost the reason we did that, and Sasha left" |

The retrieval architecture genuinely flips here (§4). The extraction pipeline is the
same one, pointed at a firehose.

---

## 2. The core engineering problem: signal ratio

A 200-engineer Slack produces roughly 50k messages/month. Maybe 200–300 contain
knowledge worth keeping. **You cannot run a model over the firehose** — that's the
mistake that kills these products, on cost and on noise.

So the pipeline is inverted relative to the OSS version: a wide deterministic filter
first, a narrow model second.

```
  50,000 msgs/mo
        │
        ▼   T0 — deterministic, $0, runs continuously
  ┌─────────────────────────────────────────────────────┐
  │ STAGE 1: structural filters                          │
  │  channel allowlist · thread depth ≥4 · participants  │
  │  ≥3 · contains code block / stack trace / config     │
  │  · has ✅ or resolution reaction · linked from a      │
  │  ticket or PR · long reply to a question             │
  └─────────────────────────────────────────────────────┘
        │  ~2,500 threads
        ▼   T0 — lexical scoring
  ┌─────────────────────────────────────────────────────┐
  │ STAGE 2: knowledge-marker lexicon + scoring          │
  │  "root cause" "turns out" "we decided" "the reason"  │
  │  "gotcha" "don't do" "workaround" "because"          │
  │  × entity match against code index (§B.5) ──────────▶│
  │  a thread naming a real symbol/service scores high   │
  └─────────────────────────────────────────────────────┘
        │  ~600 candidates
        ▼   T0 — recall against existing wiki
  ┌─────────────────────────────────────────────────────┐
  │ STAGE 3: dedup — does a note already cover this?     │
  │  ≥0.85 → attach as evidence to existing note, done   │
  │  0.55–0.85 → queue as UPDATE                         │
  └─────────────────────────────────────────────────────┘
        │  ~250 to a model  (0.5% of input — a 200× cost reduction)
        ▼   T1/T2 — extraction
  ┌─────────────────────────────────────────────────────┐
  │ STAGE 4: extract → note draft → gates → ACL → review │
  └─────────────────────────────────────────────────────┘
        │  ~60 notes/mo published, ~190 merges into existing notes
```

**Stage 2's entity match against the code index is the differentiator.** Generic
"is this important?" classification is weak and noisy. "This thread names
`RefundOrchestrator`, which is a real class with 9 commits in 90 days and zero
documentation" is a strong, cheap, deterministic signal that no pure-LLM competitor
computes, because they never built §B.5.

Everything above Stage 4 is the T0 static core from the addendum, reused. That's the
architectural payoff of having engineered it properly.

---

## 3. Sources and the ingestion layer

Consumed via MCP, so each source is a connector rather than a bespoke integration:

| Source | What's extracted | Signal quality |
|---|---|---|
| **Slack / Teams** | resolved threads in eng channels, incident channels | ★★★ highest — where reasoning actually happens |
| **GitHub / GitLab** | PR review comments explaining *why*, not *what*; commit bodies | ★★★ dense, already code-linked |
| **Atlassian (Jira)** | ticket resolution comments, "won't fix" reasoning | ★★ |
| **Atlassian (Confluence)** | existing docs — **as ground truth to dedup against, not to re-ingest** | ★★ |
| **Incidents** (PagerDuty, Datadog) | postmortems, timeline annotations | ★★★ highest value per item |
| **Email** (Gmail/Outlook) | vendor decisions, external architecture threads | ★ noisy, narrow allowlist only |
| **Linear / Asana / ClickUp / Monday** | decision records in ticket discussion | ★★ |
| **Notion** | same role as Confluence | ★★ |

**Start with two: Slack + GitHub PR comments.** They're the highest signal density and
together they cover most of what gets lost. Every additional source multiplies the
permissions surface (§5) for diminishing returns. A pilot that does two sources
excellently beats one that does eight badly.

**Ingestion mechanics:** incremental cursor per source in Postgres, backfill on
connect (bounded — last 90 days), then poll. Normalize everything to one internal
`Event {source, id, thread_id, ts, author_hash, text, refs[], acl[]}` shape at the
boundary, so the pipeline never knows which source an item came from.

---

## 4. Retrieval: where vector databases actually belong

You mentioned vector DBs — and at this scale you're right, but the placement matters.
In the OSS doc I argued *against* embeddings. That argument doesn't survive contact
with Slack ingestion, and it's worth being precise about why:

| | Vault (notes) | Ingestion corpus (messages, tickets, email) |
|---|---|---|
| Size | ~1k items | ~500k+ items |
| Structure | strict frontmatter, controlled vocabulary | none |
| Vocabulary | canonical ("idempotency key") | colloquial ("that dedup thing we did") |
| Authority | curated, verified | raw, contradictory, ephemeral |
| **Retrieval** | **lexical + frontmatter filters** | **hybrid: BM25 + dense, RRF-fused** |

So: **two stores, two strategies.**

```
┌──────────────────────────────────────────────────────────────┐
│ SOURCE OF TRUTH — git repo of markdown notes                  │
│ reviewed via PR, human-readable, portable, survives you       │
└───────────────────────┬──────────────────────────────────────┘
                        │ indexed into (derived, rebuildable)
                        ▼
┌──────────────────────────────────────────────────────────────┐
│ Postgres + pgvector                                           │
│  notes_idx      : embeddings + frontmatter columns (filters)  │
│  events_idx     : embeddings over the raw ingestion corpus    │
│  BM25 via tsvector; fuse with RRF                             │
└──────────────────────────────────────────────────────────────┘
```

Keeping markdown-in-git as the source of truth preserves the founding principle
("plain markdown or it didn't happen") at org scale, and it buys three things a
vector-DB-as-truth design can't have: PR review on knowledge changes, full history,
and a trivial exit path. **The index is always rebuildable from the repo.** Say that
in the security review and half the objections evaporate.

**Store choice — opinionated:**

- **pgvector** if they already run Postgres, which a Spring Boot shop does. One less
  system to operate, transactional consistency with your metadata, and `Spring AI`'s
  `VectorStore` abstraction supports it directly. This is the right default and it's
  also the one you can actually build well.
- **Qdrant** if you outgrow it — better filtered-search performance and payload
  filtering at multi-million scale. Keep the `VectorStore` interface so it's a swap.
- **Not Pinecone** for this: managed-only conflicts with the self-hosting requirement
  that enterprise buyers of an internal-knowledge product will insist on.

**Retrieval flow when a dev asks a question:**

```
question
  ├─▶ lexical + frontmatter recall over notes  (T0, exact, cheap)
  │      └── hit ≥0.85 → answer from the note. done. no vectors touched.
  └─▶ miss → hybrid search over notes_idx + events_idx
             → ACL filter (§5) → rerank → synthesize with citations
             → offer: "no note covers this. draft one?"
```

Note the ordering: **cheap deterministic path first, vectors only on miss.** Same
principle as the OSS version, one layer up. Most questions never reach the vector
store, which is both faster and cheaper than an embed-everything design.

---

## 5. Permissions — the thing that kills these products

A note synthesized from a private channel that surfaces in a public search is a
company-ending bug for a product like this. Design it first, not later.

**Model:** every note carries the provenance of every source that contributed to it.

```yaml
sources:
  - {system: slack, channel: C0192, scope: private, acl_group: eng-payments}
  - {system: github, repo: order-service, scope: internal}
visibility: eng-payments        # = INTERSECTION of all source ACLs
```

Rules:

1. **Visibility is the intersection, never the union.** One private source makes the
   whole note private. Non-negotiable, and it must be enforced in the store, not in
   the prompt.
2. **Mixed-scope notes get split, not downgraded.** If a thread mixes public
   architecture reasoning with a private incident detail, extract two notes.
3. **DMs and private group chats: never ingested.** Not configurable. It's the
   difference between a knowledge tool and surveillance, and it's the first question
   any works council or employee-rep body will ask.
4. **ACL re-sync on a schedule.** Someone leaves `eng-payments` → their access to
   derived notes changes. Stale ACLs are a slow-motion breach.
5. **Author attribution is opt-in and pseudonymous by default.** "Decided in
   #eng-payments, 2026-03" not "Sasha said." Attribution turns the tool into a
   performance-review input, which poisons adoption instantly.

**Legal, briefly:** EU works councils treat message-content processing as employee
monitoring — expect a consultation requirement, and design so a customer can enable
per-channel rather than org-wide. GDPR erasure against a git-tracked wiki is a real
design problem: solve it by storing `author_hash` only and never PII in note bodies,
so erasure is a key-deletion rather than a history rewrite.

---

## 6. Serving it back: the loop that closes

**The product both consumes MCP servers and exposes one.**

```
  Slack ─┐                                    ┌─▶ dev's Claude Code
  GitHub ─┼─▶ [ Knowledge Forge Server ] ─────┼─▶ dev's IDE agent
  Jira ──┘         (MCP server)               └─▶ Slack bot (/ask)
```

Tools exposed by the server: `search_knowledge`, `get_note`, `whats_undocumented`,
`explain_module`, `who_knows_about`. Every dev's agent session now has the org's
accumulated reasoning available as a tool call, ACL-filtered to that user.

This is the actual "no context loss" delivery mechanism, and it's what makes the
product sticky: once agents in an org depend on that MCP server, removing it is a
visible regression in everyone's daily workflow.

**Land-and-expand path:** OSS plugin (free, individual) → team pilot (2 sources, one
squad) → org (all sources + MCP server + reports). The OSS version is the top of the
funnel and the credibility proof — which is why it has to ship first and be genuinely
good on its own.

---

## 7. What the buyer actually pays for

Not "AI notes." Three reports, all of which are mostly T0:

1. **Doc-debt heatmap** — high-churn, high-complexity, high-question-volume, zero-doc
   modules, ranked by cost. *"Your team asked about the payments service 43 times last
   month. There is no documentation. `RefundOrchestrator` changed 9 times."*
2. **Onboarding paths** — generated ramp-up sequences per module from the wiki + code
   map. Measurable against time-to-first-PR, which is a metric managers already track.
3. **Knowledge risk / bus factor** — modules where one person authored all the
   knowledge and all the code. This is the report that gets budget approved, because
   it quantifies a risk everyone already feels.

Plus the §D fine-tuning corpus, which for an org is the strongest version of the
argument: *your engineers' reasoning about your code, in instruction format, that you
own and can fine-tune on. We never see it.*

**Pricing shape:** per-seat for the MCP server access (that's the daily-use value),
plus a platform fee per connected source. Avoid per-message pricing — it punishes the
customer for the ingestion volume that makes the product better.

---

## 8. Build order

Do not start this until the OSS version has real users. The sequencing is the plan.

| Stage | Work | Gate to proceed |
|---|---|---|
| **B0** | OSS v2.0 shipped, 30 days of your own usage, `RESULTS.md` published | ≥3 outside users reporting value |
| **B1** | Decide ADR-002 (Go default, Spring the alternative), stand up the service; Postgres + pgvector; markdown-in-git store. **Reuse the Go T0 core as a library.** | your own vault runs through the server |
| **B2** | Slack + GitHub ingestion, Stages 1–3 filters only (no LLM). Ship the doc-debt report. | filter precision measured on real data |
| **B3** | Stage 4 extraction + ACL model + review queue | a squad accepts >50% of drafts |
| **B4** | MCP server, ACL-filtered serving | devs use it daily unprompted |
| **B5** | Reports, onboarding paths, dataset export | first paid pilot |

**B2 is a shippable product with no LLM in it at all.** A doc-debt heatmap built from
Slack question volume + git churn + code index, with zero AI, zero data leaving the
network, and a trivial security review. That's a much easier first enterprise sale
than anything involving model inference over company messages — and it gets you the
corpus and the deployed footprint you need for B3.

**Stack note (rev 2):** B1–B5 is a backend service with Postgres, pgvector, async
ingestion workers, and MCP endpoints. Default is **Go** — the decisive argument is
importing the T0 core (drift, code index, similarity, filters) as a library instead
of reimplementing or shelling out; drift is the subtlest logic in the system and must
not exist twice. Java + Spring (virtual threads, Spring AI `VectorStore`) remains the
documented alternative if a customer environment or your velocity demands it — final
decision at B1 with a working Go core as evidence (STACK ADR-002). Either way it's a
backend systems project on a CV, not a plugin.

---

## 9. Risks specific to B2B

| Risk | Severity | Mitigation |
|---|---|---|
| Ingesting private/DM content | **fatal** | never ingest DMs; intersection ACLs; per-channel opt-in |
| Extraction noise floods the wiki | high | review queue, publish nothing unreviewed in year one, precision over recall |
| Works council / GDPR blocks deployment | high | self-host, author hashing, no PII in bodies, per-channel enablement |
| Vendor lock-in objection | med | markdown-in-git means the customer keeps everything on exit — lead with this |
| Building B2B before OSS traction | **high** | B0 gate is a hard gate. Servers built before demand are the standard failure mode here. |
| Slack API rate limits / cost at scale | med | incremental cursors, 90-day bounded backfill, T0 filter before any storage |

---

## 10. What to decide now

1. **Two sources or eight?** I'd commit to Slack + GitHub PR comments for B2–B3.
   Every extra connector multiplies the ACL surface for marginal signal.
2. **Is B2 (the no-LLM doc-debt product) a separate SKU?** It has a much shorter
   sales cycle and a trivial security review. It may be the actual wedge, with the
   AI layer as the upsell rather than the product.
3. **Self-host only, or managed?** Enterprise buyers of an internal-knowledge product
   will push hard for self-host. That constrains the vector store choice (pgvector /
   Qdrant, not Pinecone) — decide before you write code against an SDK.

---

## Note on your connectors in this session

You named Slack, Teams, Atlassian and mail — several of those MCP servers are present
in your setup but **not currently authorized** (Slack, Atlassian, GitHub, Notion,
Linear, Asana, PagerDuty, Datadog among them). I couldn't inspect their tool surfaces
to ground the ingestion design in the actual APIs. Authorize them via your claude.ai
connector settings (or `/mcp` in an interactive Claude Code session) and I can rework
§3 against the real tool signatures — particularly which ones expose thread-level
reads with reactions and ACL metadata, which is what Stage 1's filters depend on.
