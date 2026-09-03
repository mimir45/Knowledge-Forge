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

// The idiom the vault actually cites.
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

// A reformat must not read as a body change.
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

const javaImportSrc = `package com.food.order;

import com.food.repo.Repo;
import static com.food.util.Const.MAX;
import com.food.repo.*;

public class OrderConsumer {
    public void receive(String id) {}
}
`

func TestParseJavaImports(t *testing.T) {
	requireCgo(t)
	f, err := Parse("src/main/java/com/food/order/OrderConsumer.java", []byte(javaImportSrc))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"com.food.repo.Repo", "com.food.util.Const.MAX", "com.food.repo"}
	if !slicesEqual(f.Imports, want) {
		t.Fatalf("Imports = %v, want %v", f.Imports, want)
	}
}

const tsImportSrc = `import { Button } from '../widgets/Button';
import Foo from './foo';
import bare from 'react';
export { Bar } from './bar';
export * from './baz';

export function LoginPage() {
  return Foo();
}
`

// A re-export (`export ... from`) is a real dependency edge — it is exactly how a
// barrel file works.
func TestParseTypeScriptImports(t *testing.T) {
	requireCgo(t)
	f, err := Parse("src/pages/LoginPage.tsx", []byte(tsImportSrc))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"../widgets/Button", "./foo", "react", "./bar", "./baz"}
	if !slicesEqual(f.Imports, want) {
		t.Fatalf("Imports = %v, want %v", f.Imports, want)
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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
