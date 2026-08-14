package report

import (
	"strings"
	"testing"
)

func TestRenderKnowledgeMapSkipsUndocumentedGroups(t *testing.T) {
	groups := []CodeGroup{
		{Name: "src/app", Notes: []string{"a-note", "b-note"}},
		{Name: "src/vendor"}, // no notes: must not appear
	}
	got := string(RenderKnowledgeMap(groups))
	if !strings.Contains(got, "## src/app") {
		t.Errorf("missing documented group:\n%s", got)
	}
	if strings.Contains(got, "src/vendor") {
		t.Errorf("undocumented group leaked into the map:\n%s", got)
	}
	if !strings.Contains(got, "[[a-note]]") || !strings.Contains(got, "[[b-note]]") {
		t.Errorf("missing note links:\n%s", got)
	}
}

func TestRenderKnowledgeMapEmptyRepo(t *testing.T) {
	got := string(RenderKnowledgeMap(nil))
	if !strings.Contains(got, "No module") {
		t.Errorf("expected the empty-repo message, got:\n%s", got)
	}
}
