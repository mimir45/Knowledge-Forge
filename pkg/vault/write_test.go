package vault

import (
	"os"
	"strings"
	"testing"
)

func TestSetListWritesTheKeyAndPreservesTheRest(t *testing.T) {
	n := noteFrom(t, goodNote)
	if err := SetList(n, schema(t), "engine_trail", []string{"verify=advisor"}); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(n.Path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, "engine_trail:") || !strings.Contains(got, "verify=advisor") {
		t.Errorf("engine_trail entry missing from:\n%s", got)
	}
	if !strings.Contains(got, `title: "Kafka consumer group rebalancing"`) {
		t.Errorf("an unrelated key changed:\n%s", got)
	}
	if !strings.HasSuffix(got, "# Kafka consumer group rebalancing\n") {
		t.Errorf("body was not preserved byte-for-byte:\n%s", got)
	}
}

func TestSetListOnAnAbsentKeyAddsIt(t *testing.T) {
	n := noteFrom(t, goodNote)
	if n.FM.Has("engine_trail") {
		t.Fatal("fixture already has engine_trail; pick a key this test can prove was added")
	}
	if err := SetList(n, schema(t), "engine_trail", []string{"recall=none"}); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(n.Path)
	if !strings.Contains(string(out), "recall=none") {
		t.Errorf("SetList did not add the new key:\n%s", out)
	}
}

func TestSetListNoFrontmatterIsAnError(t *testing.T) {
	n := noteFrom(t, "# No frontmatter\n\nbody\n")
	if err := SetList(n, schema(t), "engine_trail", []string{"index=none"}); err != ErrNoFM {
		t.Errorf("err = %v, want ErrNoFM", err)
	}
}
