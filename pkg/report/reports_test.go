package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"knowledge-forge/pkg/drift"
	"knowledge-forge/pkg/gitsig"
	"knowledge-forge/pkg/graph"
	"knowledge-forge/pkg/linkcheck"
	"knowledge-forge/pkg/similarity"
)

var at = time.Date(2026, 8, 9, 14, 30, 0, 0, time.UTC)

// each returns every report under one name, built from the same fixtures, so the
// cross-cutting properties can be asserted once rather than nine times.
func each() map[string][]byte {
	return map[string][]byte{
		"coverage.md":     RenderCoverage(coverageFixture()),
		"staleness.md":    RenderStaleness(stalenessFixture(nil)),
		"duplicates.md":   RenderDuplicates(duplicatesFixture()),
		"orphans.md":      RenderOrphans(orphansFixture()),
		"gaps.md":         RenderGaps(GapsInput{Now: at}),
		"graph-health.md": RenderGraphHealth(graphFixture()),
		"churn.md":        RenderChurn(churnFixture()),
		"deadlinks.md":    RenderDeadlinks(deadlinksFixture()),
		"drift.md":        RenderDrift(driftFixture()),
		"cost.md":         RenderCost(costFixture()),
	}
}

// The property that matters most for files committed to a git repo: identical input on the
// same at must produce identical bytes. A clock in a header would make every weekly run a
// nine-file diff of nothing, and a diff that is always noise is a diff nobody reads.
func TestEveryReportIsByteIdenticalOnRerun(t *testing.T) {
	first := each()
	for i := 0; i < 10; i++ {
		for name, body := range each() {
			if !bytes.Equal(body, first[name]) {
				t.Fatalf("%s changed between runs — map order or a clock leaked in", name)
			}
		}
	}
}

func TestEveryReportHasADateNotATimestamp(t *testing.T) {
	for name, body := range each() {
		line, _, _ := strings.Cut(string(body), "\n")
		if !strings.Contains(line, "2026-08-09") {
			t.Errorf("%s header = %q, want the date", name, line)
		}
		if strings.Contains(line, "14:30") {
			t.Errorf("%s header carries a clock: %q", name, line)
		}
	}
}

// A report is read by a human and grepped by an agent; neither can do anything with a file
// that renders nothing when the vault is clean.
func TestEmptyReportsStillSayWhatTheyChecked(t *testing.T) {
	for name, body := range each() {
		if len(body) < 60 {
			t.Errorf("%s = %q, too short to be readable", name, body)
		}
	}
}

// --- drift ------------------------------------------------------------------------

func driftFixture() DriftInput {
	return DriftInput{
		Findings: []drift.Finding{
			{Note: "notes/concept/a.md", Ref: "x.go:12", Verdict: drift.Broken, Reason: "file gone"},
			{Note: "notes/concept/a.md", Ref: "y.go:3", Verdict: drift.Suspect, Reason: "body changed"},
			{Note: "notes/howto/b.md", Ref: "z.go:1", Verdict: drift.OK},
			{Note: "notes/howto/c.md", Ref: "q.go:9", Verdict: drift.Skipped, Reason: "repo absent"},
		},
		Slugs: map[string]string{"notes/concept/a.md": "a", "notes/howto/b.md": "b"},
		Now:   at,
	}
}

// The headline counts notes, not references: a note with nine broken citations is one note
// to fix, and that is the question the user actually asked. The fixture's a.md carries one
// broken and one suspect reference, so summing the two verdict lists would report two notes
// where there is one — the real vault has a note in exactly that state.
func TestDriftCountsNotesNotReferences(t *testing.T) {
	got := string(RenderDrift(driftFixture()))
	if !strings.Contains(got, "**1 note reference") {
		t.Errorf("summary should count one affected note, got:\n%s", firstLines(got, 4))
	}
	if !strings.Contains(got, "1 broken, 1 suspect") {
		t.Error("per-verdict counts should still show the note under both")
	}
	if !strings.Contains(got, "Checked 4 citations") {
		t.Error("the citation total should count references, not notes")
	}
	for _, want := range []string{"Broken", "Suspect", "Not checked — 1"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing section %q", want)
		}
	}
}

// Broken and suspect must never be merged: only broken costs a note its confidence.
func TestDriftKeepsBrokenAndSuspectApart(t *testing.T) {
	got := string(RenderDrift(driftFixture()))
	broken := strings.Index(got, "## Broken")
	suspect := strings.Index(got, "## Suspect")
	if broken < 0 || suspect < 0 || broken > suspect {
		t.Errorf("broken must be listed first; broken=%d suspect=%d", broken, suspect)
	}
}

// --- duplicates -------------------------------------------------------------------

func duplicatesFixture() DuplicatesInput {
	return DuplicatesInput{
		Pairs:     []similarity.Pair{{A: "concepts/soft-delete.md", B: "concepts/soft-deletion.md", Score: 0.575}},
		Threshold: similarity.DuplicateThreshold,
		Compared:  1142,
		Slugs:     map[string]string{"concepts/soft-delete.md": "soft-delete"},
		Types:     map[string]string{"concepts/soft-delete.md": "concept"},
		Now:       at,
	}
}

// duplicate-spec.md §5: the header states outright that nothing reaches §B.4's 0.85. A
// reader who knows the spec would otherwise assume the documented threshold was met.
func TestDuplicatesHeaderAdmitsTheSpecThresholdIsUnmet(t *testing.T) {
	got := string(RenderDuplicates(duplicatesFixture()))
	if !strings.Contains(got, "0.85") {
		t.Error("header does not mention the spec's 0.85")
	}
	if !strings.Contains(got, "same type") {
		t.Error("method section does not state the same-type restriction")
	}
}

// A note with no slug must still be nameable — it is often exactly the invalid note the
// report is complaining about.
func TestUnslugedNotesRenderAsPaths(t *testing.T) {
	got := string(RenderDuplicates(duplicatesFixture()))
	if !strings.Contains(got, "`concepts/soft-deletion.md`") {
		t.Errorf("the unslugged note vanished:\n%s", got)
	}
}

// --- staleness --------------------------------------------------------------------

func stalenessFixture(asks map[string]int) StalenessInput {
	return StalenessInput{
		Entries: []Entry{
			{Rel: "notes/concept/old.md", Slug: "old", FreshnessDays: 30,
				Verified: at.AddDate(0, 0, -100)},
			{Rel: "notes/concept/asked.md", Slug: "asked", FreshnessDays: 30,
				Verified: at.AddDate(0, 0, -40)},
			{Rel: "notes/concept/fresh.md", Slug: "fresh", FreshnessDays: 30,
				Verified: at.AddDate(0, 0, -1)},
			{Rel: "notes/concept/unmeasured.md", Slug: "unmeasured"},
		},
		Asks: asks,
		Now:  at,
	}
}

// With no ask data the product collapses to zero for every note, so ranking falls back to
// days overdue — and the header has to admit it rather than implying §B.4's ranking ran.
func TestStalenessFallsBackWhenNoAsksExist(t *testing.T) {
	got := string(RenderStaleness(stalenessFixture(nil)))
	if !strings.Contains(got, "Ranked by days overdue") {
		t.Error("header does not disclose the fallback")
	}
	if i, j := strings.Index(got, "[[old]]"), strings.Index(got, "[[asked]]"); i < 0 || i > j {
		t.Error("the most overdue note is not first")
	}
}

// The fallback must be conditional, not baked in: Phase 4's capture data lands and the
// ranking has to start using it without another code change.
func TestStalenessUsesTheProductAsSoonAsAsksExist(t *testing.T) {
	got := string(RenderStaleness(stalenessFixture(map[string]int{"asked": 10})))
	if !strings.Contains(got, "ask frequency x days overdue") {
		t.Error("header still claims the fallback with ask data present")
	}
	// asked: 10 x 10 = 100 beats old: 0 x 70 = 0.
	if i, j := strings.Index(got, "[[asked]]"), strings.Index(got, "[[old]]"); i < 0 || i > j {
		t.Errorf("weighted ranking did not reorder:\n%s", got)
	}
}

// A note with no freshness budget is unmeasured, not stale. Counting it would make the
// number meaningless the moment a note omits the field.
func TestUnmeasuredNotesAreNotStale(t *testing.T) {
	if got := string(RenderStaleness(stalenessFixture(nil))); strings.Contains(got, "[[unmeasured]]") {
		t.Error("a note with no freshness budget was reported as stale")
	}
}

// --- graph ------------------------------------------------------------------------

func orphansFixture() OrphansInput {
	return OrphansInput{
		Orphans: []string{"notes/concept/lonely.md"},
		Total:   10,
		Slugs:   map[string]string{"notes/concept/lonely.md": "lonely"},
		Types:   map[string]string{"notes/concept/lonely.md": "concept"},
		Now:     at,
	}
}

func graphFixture() GraphHealthInput {
	return GraphHealthInput{
		Components: []graph.Component{
			{Members: []string{"index.md", "a.md"}, Roots: []string{"index.md"}},
			{Members: []string{"x.md", "y.md"}},
		},
		Hubs:  []string{"index.md"},
		Total: 4,
		Now:   at,
	}
}

// The cluster with no way in is the finding graph-health.md exists for — none of its
// members is an orphan, so orphans.md cannot see it.
func TestGraphHealthNamesUnreachableClusters(t *testing.T) {
	got := string(RenderGraphHealth(graphFixture()))
	if !strings.Contains(got, "Unreachable clusters — 1") {
		t.Errorf("the rootless cluster was not reported:\n%s", got)
	}
	if !strings.Contains(got, "50%") {
		t.Error("largest-component share missing")
	}
}

// --- churn / deadlinks / coverage --------------------------------------------------

func churnFixture() ChurnInput {
	return ChurnInput{
		Stats: gitsig.Analyze([]gitsig.Commit{
			{Author: "Ada", Files: []string{"notes/concept/a.md", "notes/concept/b.md"}},
			{Author: "Ada", Files: []string{"notes/concept/a.md", "notes/concept/b.md"}},
			{Author: "Grace", Files: []string{"notes/concept/a.md"}},
		}),
		Slugs:  map[string]string{"notes/concept/a.md": "a", "notes/concept/b.md": "b"},
		Months: 6,
		Now:    at,
	}
}

func deadlinksFixture() DeadlinksInput {
	return DeadlinksInput{
		Citations: []Citation{
			{Status: linkcheck.Status{URL: "https://ok.example/", Verdict: linkcheck.Alive, Code: 200}},
			{Status: linkcheck.Status{URL: "https://gone.example/", Verdict: linkcheck.Dead, Code: 404},
				Notes: []string{"notes/concept/a.md"}},
			{Status: linkcheck.Status{URL: "https://down.example/", Verdict: linkcheck.Unreachable,
				Detail: "dial tcp: timeout"}},
		},
		Slugs: map[string]string{"notes/concept/a.md": "a"},
		Now:   at,
	}
}

// Folding unreachable into dead would let one offline run claim the whole vault rotted.
func TestDeadlinksCountsUnreachableSeparately(t *testing.T) {
	got := string(RenderDeadlinks(deadlinksFixture()))
	if !strings.Contains(got, "**1 dead, 1 unreachable, 1 alive**") {
		t.Errorf("summary conflates the verdicts:\n%s", firstLines(got, 4))
	}
	if !strings.Contains(got, "cited by [[a]]") {
		t.Error("a dead link does not name the note citing it")
	}
}

func coverageFixture() CoverageInput {
	return CoverageInput{
		Entries: []Entry{
			{Slug: "a", Type: "concept", Stack: []string{"go"}},
			{Slug: "b", Type: "concept", Stack: []string{"go", "sqlite"}},
		},
		Vocabulary: []string{"go", "sqlite", "kafka", "keycloak"},
		Types:      []string{"concept", "howto", "pitfall"},
		Now:        at,
	}
}

// Only the schema vocabulary can name an absence; counting observed stacks alone can
// never report a stack with zero notes.
func TestCoverageNamesStacksWithNoNotes(t *testing.T) {
	got := string(RenderCoverage(coverageFixture()))
	if !strings.Contains(got, "No notes — 2") {
		t.Errorf("uncovered stacks not reported:\n%s", got)
	}
	for _, want := range []string{"kafka", "keycloak"} {
		if !strings.Contains(got, "- "+want) {
			t.Errorf("missing uncovered stack %q", want)
		}
	}
	// A type in the vocabulary with no notes must still appear, at zero.
	if !strings.Contains(got, "**howto** — 0") {
		t.Error("a type with no notes vanished from the mix")
	}
}

// --- gaps -------------------------------------------------------------------------

// The two empty states mean opposite things: no data at all versus data showing nothing
// missing. Conflating them would make an unwired pipeline look like a healthy vault.
func TestGapsDistinguishesNoDataFromNoGaps(t *testing.T) {
	noLog := string(RenderGaps(GapsInput{Now: at}))
	if !strings.Contains(noLog, "No ask log yet") || !strings.Contains(noLog, "coverage.md") {
		t.Errorf("the no-data state does not explain itself:\n%s", noLog)
	}
	withLog := string(RenderGaps(GapsInput{Asks: []Ask{{Topic: "x", Count: 5, Written: true}}, Now: at}))
	if strings.Contains(withLog, "No ask log yet") {
		t.Error("a populated log still reports as unwired")
	}
}

// Asked once is a passing curiosity; §B.4's threshold is two.
func TestGapsIgnoresSingleAsks(t *testing.T) {
	in := GapsInput{Asks: []Ask{{Topic: "once", Count: 1}, {Topic: "twice", Count: 2}}, Now: at}
	got := string(RenderGaps(in))
	if strings.Contains(got, "once") || !strings.Contains(got, "twice") {
		t.Errorf("wrong topics survived the threshold:\n%s", got)
	}
}

// --- codebase ---------------------------------------------------------------------

// §B.5's last line: high churn, real size, zero notes, ranked by churn.
func TestCodebaseRanksUndocumentedChurnFirst(t *testing.T) {
	got := string(RenderCodebase(CodebaseInput{
		Repo: "leprecoin", Days: 90,
		Groups: []CodeGroup{{Name: "payments", Files: 14, Commits: 30, Notes: []string{"idempotency"}}},
		Uncovered: []Uncovered{
			{Symbol: "Small", Path: "a.go", LOC: 10, Commits: 2},
			{Symbol: "RefundOrchestrator", Path: "b.go", LOC: 312, Commits: 9},
		},
		Now: at,
	}))
	if i, j := strings.Index(got, "RefundOrchestrator"), strings.Index(got, "Small"); i < 0 || i > j {
		t.Errorf("ranked by size or name instead of churn:\n%s", got)
	}
	if !strings.Contains(got, "[[idempotency]]") {
		t.Error("module notes are not linked, so the MOC de-orphans nothing")
	}
}

// TestUncoveredTiesBreakOnPath: a symbol name is not unique in a tree — the vault's own
// `Builder` citation matches 44 declarations — so two entries can agree on churn, size and
// name and still be different files. sort.Slice is not stable, so an unbroken tie would put
// map iteration order into the rendered report.
func TestUncoveredTiesBreakOnPath(t *testing.T) {
	in := CodebaseInput{Repo: "food", Days: 90, Now: at, Uncovered: []Uncovered{
		{Symbol: "Builder", Path: "z/Order.java", LOC: 40, Commits: 3},
		{Symbol: "Builder", Path: "a/Customer.java", LOC: 40, Commits: 3},
		{Symbol: "Builder", Path: "m/Product.java", LOC: 40, Commits: 3},
	}}
	first := string(RenderCodebase(in))
	for i := 0; i < 30; i++ {
		in.Uncovered[0], in.Uncovered[2] = in.Uncovered[2], in.Uncovered[0]
		if got := string(RenderCodebase(in)); got != first {
			t.Fatalf("run %d reordered same-name symbols:\n%s", i, got)
		}
	}
	if strings.Index(first, "a/Customer.java") > strings.Index(first, "z/Order.java") {
		t.Errorf("tied symbols not ordered by path:\n%s", first)
	}
}

// TestDriftFindingTiesBreakOnReason: one note can cite one ref twice and collect two
// reasons, and the ref alone does not separate them.
func TestDriftFindingTiesBreakOnReason(t *testing.T) {
	in := DriftInput{Now: at, Slugs: map[string]string{"notes/concept/a.md": "a"},
		Findings: []drift.Finding{
			{Note: "notes/concept/a.md", Ref: "`X`", Verdict: drift.Broken, Reason: "zeta"},
			{Note: "notes/concept/a.md", Ref: "`X`", Verdict: drift.Broken, Reason: "alpha"},
		}}
	first := string(RenderDrift(in))
	for i := 0; i < 30; i++ {
		in.Findings[0], in.Findings[1] = in.Findings[1], in.Findings[0]
		if got := string(RenderDrift(in)); got != first {
			t.Fatalf("run %d reordered same-ref findings:\n%s", i, got)
		}
	}
	if strings.Index(first, "alpha") > strings.Index(first, "zeta") {
		t.Errorf("tied findings not ordered by reason:\n%s", first)
	}
}

// --- cost ---------------------------------------------------------------------------

func costFixture() CostInput {
	return CostInput{
		SpentToday:  map[string]float64{"api": 0.42, "advisor": 0},
		CapPerDay:   map[string]float64{"api": 1.00, "advisor": 2.00},
		OnExhausted: "queue",
		StageEngine: map[string]string{"recall": "none", "research": "api", "write": "none"},
		QueuedNotes: 2,
		Now:         at,
	}
}

// The three sections answer different questions; a caller grepping for one must find it
// under its own heading, not folded into the summary line.
func TestCostSeparatesSpendEngineAndQueue(t *testing.T) {
	got := string(RenderCost(costFixture()))
	for _, want := range []string{"## Spend today", "## Per-stage engine", "## Queue"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing section %q:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "recall** — none") {
		t.Error("the T0 lock is not visible in the per-stage section")
	}
	if !strings.Contains(got, "2 notes waiting") {
		t.Errorf("queue count missing:\n%s", got)
	}
}

// A tier with a cap and zero spend must still be listed — "no line" and "spent nothing"
// are different facts, and only one of them is worth an "all clear".
func TestCostListsUnspentMeteredTiers(t *testing.T) {
	got := string(RenderCost(costFixture()))
	if !strings.Contains(got, "advisor** — $0.00 of $2.00") {
		t.Errorf("unspent advisor tier vanished:\n%s", got)
	}
}

// A cap of 0 makes pkg/engine's availableMetered report the tier exhausted (remaining =
// cap - spent <= 0), so cost.md must call it unavailable — offline and claude-only both
// ship cap 0 on tiers they never route to, and "$0.00 of $0.00" reads as merely maxed out.
func TestCostZeroCapReadsUnavailable(t *testing.T) {
	in := costFixture()
	in.CapPerDay["advisor"] = 0
	got := string(RenderCost(in))
	if !strings.Contains(got, "advisor** — $0.00 spent, cap $0.00: unavailable") {
		t.Errorf("zero cap not labeled unavailable:\n%s", got)
	}
	if strings.Contains(got, "advisor** — $0.00 of $0.00") {
		t.Errorf("zero cap still reads as merely maxed out:\n%s", got)
	}
}

func firstLines(s string, n int) string {
	parts := strings.SplitN(s, "\n", n+1)
	return strings.Join(head(parts, n), "\n")
}
