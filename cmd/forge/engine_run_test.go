package main

import (
	"strings"
	"testing"

	"knowledge-forge/pkg/vault"
)

// queueNote is on_exhausted:queue's whole job: stamp pending_advisor:true and nothing
// else about the note. This is the write path SetScalars' doc-comment now names.
func TestQueueNoteStampsPendingAdvisor(t *testing.T) {
	root := fixtureCopy(t)
	rel := "concepts/hibernate.md"
	if err := queueNote(root, rel); err != nil {
		t.Fatal(err)
	}
	n, err := vault.Load(root+"/"+rel, rel)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(n.FM.Str("pending_advisor"), "true") {
		t.Errorf("pending_advisor = %q, want true", n.FM.Str("pending_advisor"))
	}
}
