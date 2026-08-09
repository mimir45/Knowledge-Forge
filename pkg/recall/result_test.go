package recall

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestResultPopulatesNeighboursOnCreateOnly(t *testing.T) {
	band := []Candidate{{Slug: "near", Score: 0.42}}
	create := DefaultThresholds.Result(Query{Question: "q"}, band)
	if create.Verdict != Create || len(create.Neighbours) != 1 {
		t.Errorf("CREATE: verdict=%s neighbours=%d, want CREATE/1",
			create.Verdict, len(create.Neighbours))
	}
	// Same band, but a strong top candidate. The caller was told to answer from the
	// vault; handing it link targets invites it to link them anyway.
	answer := DefaultThresholds.Result(Query{Question: "q"},
		append([]Candidate{{Slug: "hit", Score: 0.9}}, band...))
	if answer.Verdict != AnswerFromVault || len(answer.Neighbours) != 0 {
		t.Errorf("ANSWER: verdict=%s neighbours=%d, want ANSWER_FROM_VAULT/0",
			answer.Verdict, len(answer.Neighbours))
	}
}

func TestResultEmptyVaultIsCreateNotError(t *testing.T) {
	res := DefaultThresholds.Result(Query{Question: "how do goroutines work"}, nil)
	if res.Verdict != Create || res.TopScore != 0 {
		t.Errorf("verdict=%s top_score=%v, want CREATE/0", res.Verdict, res.TopScore)
	}
}

// Both arrays must marshal as `[]`, never `null`: the empty case is what every
// genuinely new topic hits, and a null would make each consumer special-case it.
func TestResultMarshalsEmptyArraysNotNull(t *testing.T) {
	out, err := json.Marshal(DefaultThresholds.Result(Query{Question: "q"}, nil))
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"candidates":[]`, `"neighbours":[]`} {
		if !strings.Contains(string(out), key) {
			t.Errorf("missing %s in %s", key, out)
		}
	}
}

// Channels carry the --explain breakdown and are deliberately not serialised; the
// JSON contract is the scores, not the internals of how they were reached.
func TestCandidateOmitsChannelsFromJSON(t *testing.T) {
	out, _ := json.Marshal(Candidate{Slug: "a", Channels: []Channel{{Name: "title"}}})
	if strings.Contains(string(out), "channels") {
		t.Errorf("channels leaked into the JSON contract: %s", out)
	}
}
