package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFencedBlocksExtractsContentWithoutMarkers(t *testing.T) {
	body := []byte("prose\n```java\nint x = 1;\n```\nmore prose\n`inline` ignored")
	got := FencedBlocks(body)
	if len(got) != 1 || string(got[0]) != "int x = 1;" {
		t.Errorf("FencedBlocks = %q, want one block %q", got, "int x = 1;")
	}
}

func TestFencedBlocksNoneReturnsEmpty(t *testing.T) {
	if got := FencedBlocks([]byte("no fences here")); len(got) != 0 {
		t.Errorf("got %v, want none", got)
	}
}

func TestWriteToInboxSetsLowConfidenceAndOpenQuestions(t *testing.T) {
	root := t.TempDir()
	n := noteFrom(t, goodNote)
	if err := WriteToInbox(root, n, schema(t), []string{"code gate failed: syntax error"}); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(filepath.Join(root, "_inbox", "kafka-consumer-group-rebalancing.md"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, "confidence: low") {
		t.Errorf("confidence not demoted:\n%s", got)
	}
	if !strings.Contains(got, "## Open questions") || !strings.Contains(got, "syntax error") {
		t.Errorf("open questions section missing:\n%s", got)
	}
}

func TestWriteToInboxNoSlugIsAnError(t *testing.T) {
	n := noteFrom(t, strings.Replace(goodNote, "slug: kafka-consumer-group-rebalancing\n", "", 1))
	if err := WriteToInbox(t.TempDir(), n, schema(t), nil); err == nil {
		t.Error("want an error for a slug-less draft, got nil")
	}
}
