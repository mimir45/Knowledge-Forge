package codeindex

import "testing"

const javaSrc = `package com.food.order;

public class OrderConsumer {
    private final Repo repo;

    public void receive(String id) {
        repo.save(id);
    }
}
`

const tsxSrc = `export function useAccounts() {
  return 1;
}

export class AccountsLoader {
  render() {
    return null;
  }
}
`

// The idiom the vault actually cites. `export function` is rare in the TypeScript
// corpus; `const X: FC = () => {}` is how components and hooks are written, and until
// the extractor recorded it every leprecoin citation resolved to nothing.
const arrowSrc = `const LoginPage: FC = (): ReactElement => {
  return null;
};

const Title = styled(Typography)` + "`text-align: center;`" + `;

export { LoginPage };
`

func TestParseArrowConstDeclarations(t *testing.T) {
	requireCgo(t)
	f, err := Parse("src/widgets/LoginPage/index.tsx", []byte(arrowSrc))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := f.Lookup("LoginPage"); !ok {
		t.Errorf("missing LoginPage in %+v", f.Symbols)
	}
	// A styled() call is a value, not a declaration a note cites by behaviour.
	if _, ok := f.Lookup("Title"); ok {
		t.Errorf("recorded a non-function const: %+v", f.Symbols)
	}
}

func TestParseJavaQualifiesMembers(t *testing.T) {
	requireCgo(t)
	f, err := Parse("src/main/java/com/food/order/OrderConsumer.java", []byte(javaSrc))
	if err != nil {
		t.Fatal(err)
	}
	if f.Lang != "java" {
		t.Fatalf("lang = %q", f.Lang)
	}
	s, ok := f.Lookup("OrderConsumer.receive")
	if !ok {
		t.Fatalf("symbols = %+v, want OrderConsumer.receive", f.Symbols)
	}
	if s.Kind != "method" || s.Start != 6 {
		t.Errorf("symbol = %+v, want method starting line 6", s)
	}
	// A note that cites the bare member name must still find it.
	if _, ok := f.Lookup("receive"); !ok {
		t.Error("bare member lookup failed")
	}
}

func TestParseTypeScript(t *testing.T) {
	requireCgo(t)
	f, err := Parse("src/hooks/useAccounts.tsx", []byte(tsxSrc))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"useAccounts", "AccountsLoader", "AccountsLoader.render"} {
		if _, ok := f.Lookup(want); !ok {
			t.Errorf("missing %s in %+v", want, f.Symbols)
		}
	}
}

// A reformat must not read as a body change. SUSPECT means "the behaviour this note
// describes may have moved"; gofmt-equivalent churn does not qualify, and if it did
// every note in the vault would go SUSPECT on the first repo-wide format.
func TestBodyHashIgnoresWhitespace(t *testing.T) {
	requireCgo(t)
	a, _ := Parse("A.java", []byte(javaSrc))
	reformatted := "package com.food.order;\n\npublic class OrderConsumer {\n" +
		"  private final Repo repo;\n  public void receive(String id) { repo.save(id); }\n}\n"
	b, _ := Parse("A.java", []byte(reformatted))
	sa, _ := a.Lookup("OrderConsumer.receive")
	sb, _ := b.Lookup("OrderConsumer.receive")
	if sa.BodyHash != sb.BodyHash {
		t.Errorf("body hash changed on reformat: %s vs %s", sa.BodyHash, sb.BodyHash)
	}
	if sa.Start == sb.Start {
		t.Error("expected the line number to have moved, so auto-repair has a case to fix")
	}
}

func TestLangCoverage(t *testing.T) {
	for _, c := range [][2]string{
		{"A.java", "java"}, {"a.ts", "typescript"}, {"a.tsx", "typescript"},
		{"a.kt", ""}, {"a.md", ""}, {"pom.xml", ""},
	} {
		if got := Lang(c[0]); got != c[1] {
			t.Errorf("Lang(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func requireCgo(t *testing.T) {
	t.Helper()
	if !Available() {
		t.Skip("built without cgo")
	}
}

func BenchmarkParseJava(b *testing.B) {
	if !Available() {
		b.Skip("built without cgo")
	}
	src := []byte(javaSrc)
	for b.Loop() {
		_, _ = Parse("A.java", src)
	}
}
