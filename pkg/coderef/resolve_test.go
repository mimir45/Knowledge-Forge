package coderef

import "testing"

func testRegistry() *Registry {
	return NewRegistry([]Repo{{
		Name: "food",
		Files: []string{
			"common-domain/src/main/java/com/food/domain/valueobject/Money.java",
			"order-service/src/main/java/com/food/order/domain/valueobject/Money.java",
			"order-service/src/main/java/com/food/order/OrderConsumer.java",
		},
	}, {
		Name:  "leprecoin",
		Files: []string{"src/app/page.tsx"},
	}})
}

// NF-4's exact failure, verbatim from the audit.
func TestResolveNF4Shorthand(t *testing.T) {
	res := testRegistry().Resolve(Ref{Kind: KindPath, Path: "common-domain/valueobject/Money.java"})
	if res.Status != Resolved ||
		res.RepoPath != "common-domain/src/main/java/com/food/domain/valueobject/Money.java" {
		t.Fatalf("res = %+v, want resolved to common-domain", res)
	}
	if res.Ref.Repo != "food" {
		t.Errorf("repo = %q, want food", res.Ref.Repo)
	}
}

// Both files match "domain/valueobject/Money.java" as a subsequence, but the
// common-domain path spends fewer unstated segments getting there.
func TestResolvePrefersTightestMatch(t *testing.T) {
	res := testRegistry().Resolve(Ref{Kind: KindPath, Path: "domain/valueobject/Money.java"})
	if res.Status != Resolved ||
		res.RepoPath != "common-domain/src/main/java/com/food/domain/valueobject/Money.java" {
		t.Fatalf("res = %+v, want the shorter common-domain path", res)
	}
}

// A bare filename that two modules define is Ambiguous, never a coin flip: reporting
// one of them would make drift confidently wrong about which file to watch.
func TestResolveAmbiguousBareFilename(t *testing.T) {
	res := testRegistry().Resolve(Ref{Kind: KindPath, Path: "Money.java"})
	if res.Status != Ambiguous || len(res.Ambiguity) != 2 {
		t.Fatalf("res = %+v, want ambiguous with 2 candidates", res)
	}
}

func TestResolveUnresolvedAndRepoScoping(t *testing.T) {
	rg := testRegistry()
	if res := rg.Resolve(Ref{Kind: KindPath, Path: "app/Missing.java"}); res.Status != Unresolved {
		t.Errorf("missing file: %+v, want unresolved", res)
	}
	// The same shorthand, scoped to one repo, stops being ambiguous.
	scoped := rg.Resolve(Ref{Kind: KindPath, Path: "Money.java", Repo: "food"})
	if scoped.Status != Ambiguous {
		t.Errorf("food-scoped Money.java: %+v, want still ambiguous", scoped)
	}
	if res := rg.Resolve(Ref{Kind: KindPath, Path: "page.tsx", Repo: "food"}); res.Status != Unresolved {
		t.Errorf("wrong-repo scope: %+v, want unresolved", res)
	}
}

// Symbol refs carry no path. Resolving them here would report every class name in the
// vault as a broken file reference; the symbol table in pkg/drift answers for them.
func TestResolveSymbolRefIsNotAPathLookup(t *testing.T) {
	if res := testRegistry().Resolve(Ref{Kind: KindSymbol, Symbol: "Nowhere"}); res.Status != Resolved {
		t.Errorf("symbol ref: %+v, want resolved (deferred to the symbol table)", res)
	}
}
