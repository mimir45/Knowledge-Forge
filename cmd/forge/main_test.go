package main

import (
	"strings"
	"testing"
)

// TestUsageListsEveryCommand pins the two lists in main.go against each other. Nothing
// did until now, and the failure mode is quiet in exactly the wrong direction: a command
// registered but never mentioned in the usage block is a feature that ships undiscovered.
// Two commands were added in the same edit that introduced this test, which is the moment
// the gap was worth closing.
func TestUsageListsEveryCommand(t *testing.T) {
	body := usageBody(t)
	for name := range commands {
		if !strings.Contains(body, "\n  "+name+" ") {
			t.Errorf("command %q is registered but absent from the usage block", name)
		}
	}
}

// TestUsageMentionsNoUnknownCommand is the other direction: a renamed command that left
// its old line behind sends the reader to something that no longer runs.
func TestUsageMentionsNoUnknownCommand(t *testing.T) {
	seen := 0
	for _, line := range strings.Split(usageBody(t), "\n") {
		name, _, ok := strings.Cut(strings.TrimSpace(line), " ")
		if !ok || !strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "   ") {
			continue // a continuation line, indented deeper than a command line
		}
		seen++
		if _, known := commands[name]; !known {
			t.Errorf("usage lists %q, which is not a registered command", name)
		}
	}
	// Without this the test passes trivially if the indentation rule above ever stops
	// matching any line at all — a green test that inspects nothing.
	if seen != len(commands) {
		t.Errorf("parsed %d command lines out of %d registered commands", seen, len(commands))
	}
}

// usageBody isolates the command list from the header and footer, so the prose around it
// is not scanned for command names.
func usageBody(t *testing.T) string {
	t.Helper()
	_, after, ok := strings.Cut(usage, "commands:\n")
	if !ok {
		t.Fatal("usage has no `commands:` block")
	}
	body, _, ok := strings.Cut(after, "\nrun \"forge")
	if !ok {
		t.Fatal("usage's command block has no closing line")
	}
	return "\n" + body
}
