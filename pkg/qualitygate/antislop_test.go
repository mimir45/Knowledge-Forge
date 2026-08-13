package qualitygate

import (
	"strings"
	"testing"
)

// TestBannedPhrasesMatchesShippedFile pins the parser against the literal shipped
// writing-rules.md content, per references/writing-rules.md's own doc comment: the list
// is meant to grow without a recompile, but a parser regression that silently returns zero
// phrases must fail loudly, not just let antislopGate stop catching anything.
func TestBannedPhrasesMatchesShippedFile(t *testing.T) {
	got := bannedPhrases()
	want := []string{
		"in today's fast-paced world",
		"it's important to note that",
		"it is important to note that",
		"let's dive in",
		"in conclusion",
		"game changer",
		"game-changer",
		"cutting-edge",
		"cutting edge",
		"leverage",
		"utilize",
		"delve into",
		"unlock the power of",
		"unlock the potential of",
		"seamlessly",
		"robust and scalable",
		"in the realm of",
		"at the end of the day",
		"when it comes to",
		"it goes without saying",
		"needless to say",
		"as an ai language model",
	}
	if len(got) != len(want) {
		t.Fatalf("bannedPhrases() = %d entries, want %d: %v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("bannedPhrases()[%d] = %q, want %q", i, got[i], w)
		}
	}
}

func TestAntislopGateBannedPhraseInBodyFails(t *testing.T) {
	draft := noteFrom(t, mutate(t, "# Kafka consumer group rebalancing\n",
		"# Kafka consumer group rebalancing\n\nThis is a game changer.\n"),
		"notes/concept/x.md")
	o := antislopGate(testConfig(t), draft)
	if o.Verdict != Fail || o.Remedy != RewritePass {
		t.Fatalf("antislopGate = %+v, want Fail/RewritePass", o)
	}
}

// TestAntislopGateBannedPhraseInsideCodeFenceIsIgnored mirrors vault.Wikilinks stripping
// fences before scanning: a variable literally named "leverage" is not filler prose.
func TestAntislopGateBannedPhraseInsideCodeFenceIsIgnored(t *testing.T) {
	src := mutate(t, "# Kafka consumer group rebalancing\n",
		"# Kafka consumer group rebalancing\n\n```java\nint leverage = 1;\n```\n")
	draft := noteFrom(t, src, "notes/concept/x.md")
	o := antislopGate(testConfig(t), draft)
	if o.Verdict != Pass {
		t.Fatalf("antislopGate = %+v, want Pass (banned word inside a fence must not count)", o)
	}
}

func TestAntislopGateHowtoWithoutCodeBlockFails(t *testing.T) {
	src := mutate(t, "type: concept", "type: howto")
	draft := noteFrom(t, src, "notes/howto/x.md")
	o := antislopGate(testConfig(t), draft)
	if o.Verdict != Fail || o.Remedy != RewritePass {
		t.Fatalf("antislopGate = %+v, want Fail/RewritePass (howto needs a fenced block)", o)
	}
}

func TestAntislopGateHowtoWithCodeBlockPasses(t *testing.T) {
	src := mutate(t, "type: concept", "type: howto")
	src = strings.Replace(src, "# Kafka consumer group rebalancing\n",
		"# Kafka consumer group rebalancing\n\n```bash\necho hi\n```\n", 1)
	draft := noteFrom(t, src, "notes/howto/x.md")
	o := antislopGate(testConfig(t), draft)
	if o.Verdict != Pass {
		t.Fatalf("antislopGate = %+v, want Pass", o)
	}
}
