// Package config loads AUDIT §8.4 D-2's four-layer configuration chain and exposes it
// as one typed value. The schema is D-7's union: ADDENDUM §E's blocks plus the DESIGN
// §10 keys §E never restates. Where the two overlap §E wins; where §E is merely silent,
// §10 survives.
//
// This package is a leaf on purpose. It does not import pkg/vault, even though the
// config file is frontmatter-only markdown and pkg/vault has a frontmatter parser: the
// dependency has to run the other way round eventually (paths, freshness_days and the
// recall thresholds are vault-shaped settings), and a cycle would be the result. The
// forty lines of splitting duplicated in frontmatter.go are cheaper than that cycle.
package config

// Config is the merged chain. Every field is the union schema; nothing here is optional
// in the Go sense, because the packaged layer always supplies a value.
type Config struct {
	VaultPath string `yaml:"vault_path"`
	RepoPath  string `yaml:"repo_path"`

	Paths         Paths            `yaml:"paths"`
	Trigger       Trigger          `yaml:"trigger"`
	Recall        Recall           `yaml:"recall"`
	FreshnessDays map[string]int   `yaml:"freshness_days"`
	Engines       Engines          `yaml:"engines"`
	Pipeline      map[string]Stage `yaml:"pipeline"`
	Research      Research         `yaml:"research"`
	Verify        Verify           `yaml:"verify"`
	Write         Write            `yaml:"write"`
	Static        Static           `yaml:"static"`
	Check         Check            `yaml:"check"`
	Garden        Garden           `yaml:"garden"`
	Dataset       Dataset          `yaml:"dataset"`
	Telemetry     Telemetry        `yaml:"telemetry"`

	// Layers are the files that contributed, lowest precedence first. Carried so an
	// error can name the file a bad value came from, and so `forge init` can report
	// what it is about to shadow.
	Layers []string `yaml:"-"`
}

// Paths is DESIGN §10's vault topology, relative to VaultPath.
type Paths struct {
	Notes   string `yaml:"notes"`
	MOC     string `yaml:"moc"`
	Inbox   string `yaml:"inbox"`
	Archive string `yaml:"archive"`
	Index   string `yaml:"index"`
}

// Trigger is DESIGN §10: ask | auto | manual.
type Trigger struct {
	Mode string `yaml:"mode"`
}

// Recall is DESIGN §5.3's decision tree, moved here by AUDIT §8.4 D-7. Neighbour is not
// in §10 — it is pkg/recall's third threshold, and leaving it compiled in while the
// other two moved would be exactly the "threshold in two places" doc.go warns about.
type Recall struct {
	Strategy          string  `yaml:"strategy"`
	AnswerThreshold   float64 `yaml:"answer_threshold"`
	UpdateThreshold   float64 `yaml:"update_threshold"`
	NeighbourMinScore float64 `yaml:"neighbour_min_score"`
}

// Engines is ADDENDUM §E's engine block: the four tiers and what each costs.
type Engines struct {
	Default string  `yaml:"default"`
	API     API     `yaml:"api"`
	Advisor Advisor `yaml:"advisor"`
	Local   Local   `yaml:"local"`
	Budget  Budget  `yaml:"budget"`
	Routing Routing `yaml:"routing"`
}

type API struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	KeyEnv   string `yaml:"key_env"`
	BaseURL  string `yaml:"base_url"`
}

type Advisor struct {
	Model string `yaml:"model"`
	Mode  string `yaml:"mode"`
}

// Local is engines.local — a routing target, not a fifth Engine implementation. Phase 3b's
// select.go resolves the "local" alias to the api backend pointed at BaseURL.
type Local struct {
	Enabled bool   `yaml:"enabled"`
	Model   string `yaml:"model"`
	BaseURL string `yaml:"base_url"`
}

// Budget is §E's block. OnExhausted defaults to queue (D-5: §E beats §A.4's degrade).
// The counters behind it live in SQLite and survive forge reindex (D-8) — the one
// documented exception to "SQLite is purely derived". Phase 3b builds that.
type Budget struct {
	AdvisorUSDPerDay float64 `yaml:"advisor_usd_per_day"`
	APIUSDPerDay     float64 `yaml:"api_usd_per_day"`
	OnExhausted      string  `yaml:"on_exhausted"`
}

type Routing struct {
	AdvisorWhen AdvisorWhen `yaml:"advisor_when"`
}

// AdvisorWhen is the wider of the two shapes the docs give. §E drops §A.4's stack_in;
// D-5's carry-over says the union keeps it, because dropping it silently narrows when
// the expensive critique tier fires on security, auth and payments notes.
type AdvisorWhen struct {
	Type            []string `yaml:"type"`
	ConfidenceBelow string   `yaml:"confidence_below"`
	StackIn         []string `yaml:"stack_in"`
	UserFlag        string   `yaml:"user_flag"`
}

// Stage is one pipeline stage's engine assignment. Fallback and Then are §E's chain for
// verify: try the advisor, fall back to local, then host.
type Stage struct {
	Engine   string `yaml:"engine"`
	Fallback string `yaml:"fallback,omitempty"`
	Then     string `yaml:"then,omitempty"`
}

type Research struct {
	MaxSources   int      `yaml:"max_sources"`
	Prefer       []string `yaml:"prefer"`
	AllowDomains []string `yaml:"allow_domains"`
	DenyDomains  []string `yaml:"deny_domains"`
	UseDocsMCP   bool     `yaml:"use_docs_mcp"`
	ScanCodebase bool     `yaml:"scan_codebase"`
}

// Verify is verify-time policy — Phase 4's gate stage. DuplicateThreshold is deliberately
// its own field, not a read of Check.DuplicateThreshold: a user lowering the weekly
// report's threshold to see more pairs must not silently change what the write-time gate
// trips on. See references/duplicate-spec.md §6.
type Verify struct {
	RunCode            string   `yaml:"run_code"`
	RequireCitationFor []string `yaml:"require_citation_for"`
	DuplicateThreshold float64  `yaml:"duplicate_threshold"`
}

type Write struct {
	Language     string `yaml:"language"`
	MaxNoteWords int    `yaml:"max_note_words"`
	Diagrams     string `yaml:"diagrams"`
}

// Static is ADDENDUM §E's T0 block — everything the zero-model-call core does.
type Static struct {
	CodeIndex  bool      `yaml:"code_index"`
	Languages  []string  `yaml:"languages"`
	GitSignals bool      `yaml:"git_signals"`
	Drift      Drift     `yaml:"drift"`
	LinkCheck  LinkCheck `yaml:"linkcheck"`
	LogBack    LogBack   `yaml:"logback"`

	// CacheTTLDays is Phase 5's forge cache-source TTL for .forge/cache/<hash>.md
	// entries. Zero means unset, not "expire immediately" — the command-level default
	// (30) is applied at the call site, matching Check.ChurnDays' own pattern rather
	// than baking a magic number into the config chain's zero value.
	CacheTTLDays int `yaml:"cache_ttl_days"`
}

// Drift is §B.6. Trigger stays git — the invariant is that drift never runs on file save
// and never reads the uncommitted tree, so this key exists to be read, not to be widened.
type Drift struct {
	Enabled               bool   `yaml:"enabled"`
	Trigger               string `yaml:"trigger"`
	Branch                string `yaml:"branch"`
	AutoRepairLineNumbers bool   `yaml:"auto_repair_line_numbers"`
	OnBroken              string `yaml:"on_broken"`
	OnRestored            string `yaml:"on_restored"`
}

type LinkCheck struct {
	Enabled  bool `yaml:"enabled"`
	TimeoutS int  `yaml:"timeout_s"`
}

// LogBack is §B.7. InlineMarkers is opt-in and defaults false: the invariant is that
// log-back never modifies code semantics, comments and separate files only.
type LogBack struct {
	KnowledgeMap     bool `yaml:"knowledge_map"`
	ClaudeMDFragment bool `yaml:"claude_md_fragment"`
	InlineMarkers    bool `yaml:"inline_markers"`
}

// Check is §C's weekly checker. ChurnDays and DuplicateThreshold are not in §E — they
// are the two numbers cmd/forge/check.go and pkg/similarity compiled in, moved here by
// this phase's step 1.
type Check struct {
	Enabled            bool     `yaml:"enabled"`
	Schedule           string   `yaml:"schedule"`
	AIPass             bool     `yaml:"ai_pass"`
	Reports            []string `yaml:"reports"`
	DrainAdvisorQueue  bool     `yaml:"drain_advisor_queue"`
	ChurnDays          int      `yaml:"churn_days"`
	DuplicateThreshold float64  `yaml:"duplicate_threshold"`
}

// Garden is DESIGN §10. It overlaps Check.Schedule and is kept because §E is silent on
// it; the two are reconciled in Phase 5, where the scheduler actually exists.
type Garden struct {
	Enabled  bool   `yaml:"enabled"`
	Schedule string `yaml:"schedule"`
}

// Dataset is §D. Capture names the tiers d1…d5, not booleans, so a new tier is a list
// entry rather than a schema change.
type Dataset struct {
	Enabled           bool     `yaml:"enabled"`
	Capture           []string `yaml:"capture"`
	AnonymizeOnExport bool     `yaml:"anonymize_on_export"`
}

// Telemetry is DESIGN §10. The invariant it configures: topic and hash only, never raw
// question text, code, or file contents — scope: team does not relax that.
type Telemetry struct {
	Enabled bool   `yaml:"enabled"`
	Scope   string `yaml:"scope"`
}
