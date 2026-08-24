//go:build cgo

package codeindex

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
)

// Available reports whether this build can parse source. True here, false in the
// no-cgo build, so callers can degrade with a clear message instead of a nil result.
func Available() bool { return true }

func grammar(lang string) *sitter.Language {
	switch lang {
	case "java":
		return java.GetLanguage()
	case "typescript":
		// The tsx grammar is a superset of typescript's and parses plain .ts as well,
		// so one grammar covers .ts/.tsx/.js/.jsx and one fewer C library ships.
		return tsx.GetLanguage()
	}
	return nil
}

// declKinds maps a tree-sitter node type to the Symbol.Kind we record. Anything not
// listed is walked through but not recorded: drift compares declarations, and a note
// does not cite an if-statement.
var declKinds = map[string]string{
	"class_declaration":          "class",
	"interface_declaration":      "interface",
	"enum_declaration":           "enum",
	"record_declaration":         "record",
	"method_declaration":         "method",
	"constructor_declaration":    "constructor",
	"method_definition":          "method",
	"function_declaration":       "function",
	"type_alias_declaration":     "type",
	"abstract_class_declaration": "class",
}

// arrowValues are the right-hand sides that make a `const` a declaration worth
// recording. `const Login = () => {}` is how nearly every component and hook in the
// TypeScript corpus is written, and a note citing `Login` is citing this.
var arrowValues = map[string]bool{
	"arrow_function": true, "function_expression": true, "function": true,
}

// kindOf classifies a node, or reports that it is not a declaration. It is the one
// place the extractor's rules live, which is what Extractor versions.
func kindOf(n *sitter.Node) (string, bool) {
	if k, ok := declKinds[n.Type()]; ok {
		return k, true
	}
	if n.Type() != "variable_declarator" {
		return "", false
	}
	v := n.ChildByFieldName("value")
	if v == nil || !arrowValues[v.Type()] {
		return "", false
	}
	return "function", true
}

// Parse extracts the declarations from one source file.
func Parse(path string, src []byte) (File, error) {
	lang := Lang(path)
	g := grammar(lang)
	if g == nil {
		return File{}, nil
	}
	p := sitter.NewParser()
	p.SetLanguage(g)
	tree, err := p.ParseCtx(context.Background(), nil, src)
	if err != nil {
		return File{}, err
	}
	defer tree.Close()
	f := File{Path: path, Lang: lang}
	walk(tree.RootNode(), src, "", &f)
	return f, nil
}

// walk descends the tree, recording every declaration and passing the enclosing type
// name down so members come out qualified.
func walk(n *sitter.Node, src []byte, prefix string, f *File) {
	scope := prefix
	if kind, ok := kindOf(n); ok {
		if name := nameOf(n, src); name != "" {
			scope = qualify(prefix, name)
			f.Symbols = append(f.Symbols, symbolOf(n, src, scope, kind))
		}
	}
	if imp, ok := importOf(n, src); ok {
		f.Imports = append(f.Imports, imp)
	}
	for i := 0; i < int(n.NamedChildCount()); i++ {
		walk(n.NamedChild(i), src, scope, f)
	}
}

// importPathKinds are the node types that can name something a file depends on. Kept
// separate from declKinds: those record what a file declares, this records what it
// imports, and DependsOn (cmd/forge) is the only reader of the latter.
var importPathKinds = map[string]bool{
	"import_declaration": true, // java: import [static] a.b.C[.*];
	"import_statement":   true, // typescript: import ... from '...'
	"export_statement":   true, // typescript re-export: export ... from '...'
}

// importOf extracts one import's raw target, or reports the node names no import at all
// (an `export class Foo {}` is an export_statement with no source to re-export from).
func importOf(n *sitter.Node, src []byte) (string, bool) {
	if !importPathKinds[n.Type()] {
		return "", false
	}
	if n.Type() == "import_declaration" {
		return javaImportPath(n, src)
	}
	return tsImportSource(n, src)
}

// javaImportPath reads the qualified name out of `import [static] <name>[.*];`. The
// grammar's scoped_identifier already excludes both "static" and a trailing ".*" — a
// wildcard import comes out as the bare package name, a class import as package+class,
// and a static member import as package+class+member, all of which resolveJavaImport
// (cmd/forge) handles by trimming from the right.
func javaImportPath(n *sitter.Node, src []byte) (string, bool) {
	for i := 0; i < int(n.NamedChildCount()); i++ {
		c := n.NamedChild(i)
		if c.Type() == "scoped_identifier" || c.Type() == "identifier" {
			return string(src[c.StartByte():c.EndByte()]), true
		}
	}
	return "", false
}

// tsImportSource reads the module specifier off import_statement/export_statement's
// "source" field, quotes stripped. A plain export (no re-export) has no source field.
func tsImportSource(n *sitter.Node, src []byte) (string, bool) {
	f := n.ChildByFieldName("source")
	if f == nil {
		return "", false
	}
	return strings.Trim(string(src[f.StartByte():f.EndByte()]), `"'`), true
}

func symbolOf(n *sitter.Node, src []byte, name, kind string) Symbol {
	body := n
	if b := n.ChildByFieldName("body"); b != nil {
		body = b
	}
	return Symbol{
		Name:     name,
		Kind:     kind,
		Start:    int(n.StartPoint().Row) + 1,
		End:      int(n.EndPoint().Row) + 1,
		BodyHash: hashBody(src[body.StartByte():body.EndByte()]),
	}
}

// nameOf reads the declaration's name. TypeScript's method_definition and Java's
// declarations all expose it as the "name" field; a computed or destructured name has
// none, and an unnamed declaration is not something a note can cite.
func nameOf(n *sitter.Node, src []byte) string {
	f := n.ChildByFieldName("name")
	if f == nil {
		return ""
	}
	return string(src[f.StartByte():f.EndByte()])
}

func qualify(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}
