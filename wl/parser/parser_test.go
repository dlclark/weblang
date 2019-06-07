package parser

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"weblang/wl/ast"
	"weblang/wl/token"
)

var validFiles = []string{
	"parser.go",
	"parser_test.go",
	"error_test.go",
	"short_test.go",
}

/* These tests just parse the local files for Go

We should use testdata\*.wl files for this

func TestParse(t *testing.T) {
	for _, filename := range validFiles {
		_, err := ParseFile(token.NewFileSet(), filename, nil, DeclarationErrors)
		if err != nil {
			t.Fatalf("ParseFile(%s): %v", filename, err)
		}
	}
}

func nameFilter(filename string) bool {
	switch filename {
	case "parser.go", "interface.go", "parser_test.go":
		return true
	case "parser.go.orig":
		return true // permit but should be ignored by ParseDir
	}
	return false
}

func dirFilter(f os.FileInfo) bool { return nameFilter(f.Name()) }

func TestParseDir(t *testing.T) {
	path := "."
	pkgs, err := ParseDir(token.NewFileSet(), path, dirFilter, 0)
	if err != nil {
		t.Fatalf("ParseDir(%s): %v", path, err)
	}
	if n := len(pkgs); n != 1 {
		t.Errorf("got %d packages; want 1", n)
	}
	pkg := pkgs["parser"]
	if pkg == nil {
		t.Errorf(`package "parser" not found`)
		return
	}
	if n := len(pkg.Files); n != 3 {
		t.Errorf("got %d package files; want 3", n)
	}
	for filename := range pkg.Files {
		if !nameFilter(filename) {
			t.Errorf("unexpected package file: %s", filename)
		}
	}
}*/

func TestParseExpr(t *testing.T) {
	// just kicking the tires:
	// a valid arithmetic expression
	src := "a + b"
	x, err := ParseExpr(src)
	if err != nil {
		t.Errorf("ParseExpr(%q): %v", src, err)
	}
	// sanity check
	if _, ok := x.(*ast.BinaryExpr); !ok {
		t.Errorf("ParseExpr(%q): got %T, want *ast.BinaryExpr", src, x)
	}

	// a valid type expression
	src = "struct{x int}"
	x, err = ParseExpr(src)
	if err != nil {
		t.Errorf("ParseExpr(%q): %v", src, err)
	}
	// sanity check
	if _, ok := x.(*ast.StructType); !ok {
		t.Errorf("ParseExpr(%q): got %T, want *ast.StructType", src, x)
	}

	// an invalid expression
	src = "a + *"
	if _, err := ParseExpr(src); err == nil {
		t.Errorf("ParseExpr(%q): got no error", src)
	}

	// a valid expression followed by extra tokens is invalid
	src = "a[i] := x"
	if _, err := ParseExpr(src); err == nil {
		t.Errorf("ParseExpr(%q): got no error", src)
	}

	// a semicolon is not permitted unless automatically inserted
	src = "a + b\n"
	if _, err := ParseExpr(src); err != nil {
		t.Errorf("ParseExpr(%q): got error %s", src, err)
	}
	src = "a + b;"
	if _, err := ParseExpr(src); err == nil {
		t.Errorf("ParseExpr(%q): got no error", src)
	}

	// various other stuff following a valid expression
	const validExpr = "a + b"
	const anything = "dh3*#D)#_"
	for _, c := range "!)]};," {
		src := validExpr + string(c) + anything
		if _, err := ParseExpr(src); err == nil {
			t.Errorf("ParseExpr(%q): got no error", src)
		}
	}

	// ParseExpr must not crash
	for _, src := range valids {
		ParseExpr(src)
	}
}

func TestColonEqualsScope(t *testing.T) {
	f, err := ParseFile(token.NewFileSet(), "", `package p; func f() { x, y, z := x, y, z }`, 0)
	if err != nil {
		t.Fatal(err)
	}

	// RHS refers to undefined globals; LHS does not.
	as := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)
	for _, v := range as.Rhs {
		id := v.(*ast.Ident)
		if id.Obj != nil {
			t.Errorf("rhs %s has Obj, should not", id.Name)
		}
	}
	for _, v := range as.Lhs {
		id := v.(*ast.Ident)
		if id.Obj == nil {
			t.Errorf("lhs %s does not have Obj, should", id.Name)
		}
	}
}

func TestVarScope(t *testing.T) {
	f, err := ParseFile(token.NewFileSet(), "", `package p; func f() { var x, y, z = x, y, z }`, 0)
	if err != nil {
		t.Fatal(err)
	}

	// RHS refers to undefined globals; LHS does not.
	as := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.DeclStmt).Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
	for _, v := range as.Values {
		id := v.(*ast.Ident)
		if id.Obj != nil {
			t.Errorf("rhs %s has Obj, should not", id.Name)
		}
	}
	for _, id := range as.Names {
		if id.Obj == nil {
			t.Errorf("lhs %s does not have Obj, should", id.Name)
		}
	}
}

func TestObjects(t *testing.T) {
	const src = `
package p
import fmt "fmt"
const pi = 3.14
type T struct{}
var x int
func f() { L: }
`

	f, err := ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	objects := map[string]ast.ObjKind{
		"p":   ast.Bad, // not in a scope
		"fmt": ast.Bad, // not resolved yet
		"pi":  ast.Con,
		"T":   ast.Typ,
		"x":   ast.Var,
		"int": ast.Bad, // not resolved yet
		"f":   ast.Fun,
		"L":   ast.Lbl,
	}

	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			obj := ident.Obj
			if obj == nil {
				if objects[ident.Name] != ast.Bad {
					t.Errorf("no object for %s", ident.Name)
				}
				return true
			}
			if obj.Name != ident.Name {
				t.Errorf("names don't match: obj.Name = %s, ident.Name = %s", obj.Name, ident.Name)
			}
			kind := objects[ident.Name]
			if obj.Kind != kind {
				t.Errorf("%s: obj.Kind = %s; want %s", ident.Name, obj.Kind, kind)
			}
		}
		return true
	})
}

func TestUnresolved(t *testing.T) {
	f, err := ParseFile(token.NewFileSet(), "", `
package p
//
func f1a(int)
func f2a(byte, int, float)
func f3a(a, b int, c float)
func f4a(...complex)
func f5a(a s1a, b ...complex)
//
func f1b(int)
func f2b([]byte, (int), float)
func f3b(a, b int, c []float)
func f4b(...complex)
func f5b(a s1a, b ...[]complex)
//
type s1a struct { int }
type s2a struct { byte; int; s1a }
type s3a struct { a, b int; c float }
//
type s1b struct { int }
type s2b struct { byte; int; float }
type s3b struct { a, b s3b; c []float }
`, 0)
	if err != nil {
		t.Fatal(err)
	}

	want := "int " + // f1a
		"byte int float " + // f2a
		"int float " + // f3a
		"complex " + // f4a
		"complex " + // f5a
		//
		"int " + // f1b
		"byte int float " + // f2b
		"int float " + // f3b
		"complex " + // f4b
		"complex " + // f5b
		//
		"int " + // s1a
		"byte int " + // s2a
		"int float " + // s3a
		//
		"int " + // s1a
		"byte int float " + // s2a
		"float " // s3a

	// collect unresolved identifiers
	var buf bytes.Buffer
	for _, u := range f.Unresolved {
		buf.WriteString(u.Name)
		buf.WriteByte(' ')
	}
	got := buf.String()

	if got != want {
		t.Errorf("\ngot:  %s\nwant: %s", got, want)
	}
}

var imports = map[string]bool{
	`"a"`:        true,
	"`a`":        true,
	`"a/b"`:      true,
	`"a.b"`:      true,
	`"m\x61th"`:  true,
	`"greek/αβ"`: true,
	`""`:         false,

	// Each of these pairs tests both `` vs "" strings
	// and also use of invalid characters spelled out as
	// escape sequences and written directly.
	// For example `"\x00"` tests import "\x00"
	// while "`\x00`" tests import `<actual-NUL-byte>`.
	`"\x00"`:     false,
	"`\x00`":     false,
	`"\x7f"`:     false,
	"`\x7f`":     false,
	`"a!"`:       false,
	"`a!`":       false,
	`"a b"`:      false,
	"`a b`":      false,
	`"a\\b"`:     false,
	"`a\\b`":     false,
	"\"`a`\"":    false,
	"`\"a\"`":    false,
	`"\x80\x80"`: false,
	"`\x80\x80`": false,
	`"\xFFFD"`:   false,
	"`\xFFFD`":   false,
}

func TestImports(t *testing.T) {
	for path, isValid := range imports {
		src := fmt.Sprintf("package p; import %s", path)
		_, err := ParseFile(token.NewFileSet(), "", src, 0)
		switch {
		case err != nil && isValid:
			t.Errorf("ParseFile(%s): got %v; expected no error", src, err)
		case err == nil && !isValid:
			t.Errorf("ParseFile(%s): got no error; expected one", src)
		}
	}
}

func TestCommentGroups(t *testing.T) {
	f, err := ParseFile(token.NewFileSet(), "", `
package p /* 1a */ /* 1b */      /* 1c */ // 1d
/* 2a
*/
// 2b
const pi = 3.1415
/* 3a */ // 3b
/* 3c */ const e = 2.7182

// Example from issue 3139
func ExampleCount() {
	fmt.Println(strings.Count("cheese", "e"))
	fmt.Println(strings.Count("five", "")) // before & after each rune
	// Output:
	// 3
	// 5
}
`, ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	expected := [][]string{
		{"/* 1a */", "/* 1b */", "/* 1c */", "// 1d"},
		{"/* 2a\n*/", "// 2b"},
		{"/* 3a */", "// 3b", "/* 3c */"},
		{"// Example from issue 3139"},
		{"// before & after each rune"},
		{"// Output:", "// 3", "// 5"},
	}
	if len(f.Comments) != len(expected) {
		t.Fatalf("got %d comment groups; expected %d", len(f.Comments), len(expected))
	}
	for i, exp := range expected {
		got := f.Comments[i].List
		if len(got) != len(exp) {
			t.Errorf("got %d comments in group %d; expected %d", len(got), i, len(exp))
			continue
		}
		for j, exp := range exp {
			got := got[j].Text
			if got != exp {
				t.Errorf("got %q in group %d; expected %q", got, i, exp)
			}
		}
	}
}

func getField(file *ast.File, fieldname string) *ast.Field {
	parts := strings.Split(fieldname, ".")
	for _, d := range file.Decls {
		if d, ok := d.(*ast.GenDecl); ok && d.Tok == token.TYPE {
			for _, s := range d.Specs {
				if s, ok := s.(*ast.TypeSpec); ok && s.Name.Name == parts[0] {
					if s, ok := s.Type.(*ast.StructType); ok {
						for _, f := range s.Fields.List {
							for _, name := range f.Names {
								if name.Name == parts[1] {
									return f
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}

// Don't use ast.CommentGroup.Text() - we want to see exact comment text.
func commentText(c *ast.CommentGroup) string {
	var buf bytes.Buffer
	if c != nil {
		for _, c := range c.List {
			buf.WriteString(c.Text)
		}
	}
	return buf.String()
}

func checkFieldComments(t *testing.T, file *ast.File, fieldname, lead, line string) {
	f := getField(file, fieldname)
	if f == nil {
		t.Fatalf("field not found: %s", fieldname)
	}
	if got := commentText(f.Doc); got != lead {
		t.Errorf("got lead comment %q; expected %q", got, lead)
	}
	if got := commentText(f.Comment); got != line {
		t.Errorf("got line comment %q; expected %q", got, line)
	}
}

func TestLeadAndLineComments(t *testing.T) {
	f, err := ParseFile(token.NewFileSet(), "", `
package p
type T struct {
	/* F1 lead comment */
	//
	F1 int  /* F1 */ // line comment
	// F2 lead
	// comment
	F2 int  // F2 line comment
	// f3 lead comment
	f3 int  // f3 line comment
}
`, ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	checkFieldComments(t, f, "T.F1", "/* F1 lead comment *///", "/* F1 */// line comment")
	checkFieldComments(t, f, "T.F2", "// F2 lead// comment", "// F2 line comment")
	checkFieldComments(t, f, "T.f3", "// f3 lead comment", "// f3 line comment")
	ast.FileExports(f)
	checkFieldComments(t, f, "T.F1", "/* F1 lead comment *///", "/* F1 */// line comment")
	checkFieldComments(t, f, "T.F2", "// F2 lead// comment", "// F2 line comment")
	if getField(f, "T.f3") != nil {
		t.Error("not expected to find T.f3")
	}
}

// TestIssue9979 verifies that empty statements are contained within their enclosing blocks.
func TestIssue9979(t *testing.T) {
	for _, src := range []string{
		"package p; func f() {;}",
		"package p; func f() {L:}",
		"package p; func f() {L:;}",
		"package p; func f() {L:\n}",
		"package p; func f() {L:\n;}",
		"package p; func f() { ; }",
		"package p; func f() { L: }",
		"package p; func f() { L: ; }",
		"package p; func f() { L: \n}",
		"package p; func f() { L: \n; }",
	} {
		fset := token.NewFileSet()
		f, err := ParseFile(fset, "", src, 0)
		if err != nil {
			t.Fatal(err)
		}

		var pos, end token.Pos
		ast.Inspect(f, func(x ast.Node) bool {
			switch s := x.(type) {
			case *ast.BlockStmt:
				pos, end = s.Pos()+1, s.End()-1 // exclude "{", "}"
			case *ast.LabeledStmt:
				pos, end = s.Pos()+2, s.End() // exclude "L:"
			case *ast.EmptyStmt:
				// check containment
				if s.Pos() < pos || s.End() > end {
					t.Errorf("%s: %T[%d, %d] not inside [%d, %d]", src, s, s.Pos(), s.End(), pos, end)
				}
				// check semicolon
				offs := fset.Position(s.Pos()).Offset
				if ch := src[offs]; ch != ';' != s.Implicit {
					want := "want ';'"
					if s.Implicit {
						want = "but ';' is implicit"
					}
					t.Errorf("%s: found %q at offset %d; %s", src, ch, offs, want)
				}
			}
			return true
		})
	}
}

// TestIncompleteSelection ensures that an incomplete selector
// expression is parsed as a (blank) *ast.SelectorExpr, not a
// *ast.BadExpr.
func TestIncompleteSelection(t *testing.T) {
	for _, src := range []string{
		"package p; var _ = fmt.",             // at EOF
		"package p; var _ = fmt.\ntype X int", // not at EOF
	} {
		fset := token.NewFileSet()
		f, err := ParseFile(fset, "", src, 0)
		if err == nil {
			t.Errorf("ParseFile(%s) succeeded unexpectedly", src)
			continue
		}

		const wantErr = "expected selector or type assertion"
		if !strings.Contains(err.Error(), wantErr) {
			t.Errorf("ParseFile returned wrong error %q, want %q", err, wantErr)
		}

		var sel *ast.SelectorExpr
		ast.Inspect(f, func(n ast.Node) bool {
			if n, ok := n.(*ast.SelectorExpr); ok {
				sel = n
			}
			return true
		})
		if sel == nil {
			t.Error("found no *ast.SelectorExpr")
			continue
		}
		const wantSel = "&{fmt _}"
		if fmt.Sprint(sel) != wantSel {
			t.Errorf("found selector %s, want %s", sel, wantSel)
			continue
		}
	}
}

func TestLastLineComment(t *testing.T) {
	const src = `package main
type x int // comment
`
	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	comment := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Comment.List[0].Text
	if comment != "// comment" {
		t.Errorf("got %q, want %q", comment, "// comment")
	}
}

func TestCatchBasics(t *testing.T) {
	const src = `package main
	func f() {
		catch func(e error) {
			console.Log(e)
		}
		catch pkg.Test.SomeFunction(inp+1).Handler
	}`
	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	catch := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.CatchStmt)
	if want, got := "3:3", fset.Position(catch.Pos()).String(); want != got {
		t.Errorf("1 Pos() wanted %v got %v", want, got)
	}
	fun := catch.Fun.(*ast.FuncLit)
	if want, got := len(fun.Body.List), 1; want != got {
		t.Errorf("1 Fun Body Len, wanted %v, got %v", want, got)
	}

	catch = f.Decls[0].(*ast.FuncDecl).Body.List[1].(*ast.CatchStmt)

	if want, got := "6:3", fset.Position(catch.Pos()).String(); want != got {
		t.Errorf("2 Pos() wanted %v got %v", want, got)
	}
	fun2 := catch.Fun.(*ast.SelectorExpr)
	if want, got := fun2.Sel.Name, "Handler"; want != got {
		t.Errorf("2 Fun Body Len, wanted %v, got %v", want, got)
	}
}

func TestUnionSwitchBasic(t *testing.T) {
	const src = `package main
	func f() {
		switch v := val.(union) {
		case T:
			v.SomeT()
		case S:
			v.SomeS()
		default:
			// v?
		}
	}`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	union := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.UnionSwitchStmt)
	if want, got := "3:3", fset.Position(union.Pos()).String(); want != got {
		t.Errorf("Pos() wanted %v got %v", want, got)
	}
	if want, got := ast.SpecialTypeAssertUnion, union.Assign.(*ast.AssignStmt).Rhs[0].(*ast.TypeAssertExpr).Special; want != got {
		t.Errorf("Special wanted %v got %v", want, got)
	}
	if want, got := ast.SpecialTypeAssertUnion, union.Assign.(*ast.AssignStmt).Rhs[0].(*ast.TypeAssertExpr).Special; want != got {
		t.Errorf("Special wanted %v got %v", want, got)
	}
	if want, got := 3, len(union.Body.List); want != got {
		t.Errorf("Case statement count, wanted %v, got %v", want, got)
	}
	if want, got := 1, len(union.Body.List[0].(*ast.CaseClause).Body); want != got {
		t.Errorf("Case 1 statement count, wanted %v, got %v", want, got)
	}
	if want, got := "T", union.Body.List[0].(*ast.CaseClause).List[0].(*ast.Ident).Name; want != got {
		t.Errorf("Case 1 statement count, wanted %v, got %v", want, got)
	}
	if want, got := 1, len(union.Body.List[1].(*ast.CaseClause).Body); want != got {
		t.Errorf("Case 2 statement count, wanted %v, got %v", want, got)
	}
	if want, got := "S", union.Body.List[1].(*ast.CaseClause).List[0].(*ast.Ident).Name; want != got {
		t.Errorf("Case 2 statement count, wanted %v, got %v", want, got)
	}
	if want, got := 0, len(union.Body.List[2].(*ast.CaseClause).Body); want != got {
		t.Errorf("Case 2 body statement count, wanted %v, got %v", want, got)
	}
	if want, got := 0, len(union.Body.List[2].(*ast.CaseClause).List); want != got {
		t.Errorf("Case 2 list item count, wanted %v, got %v", want, got)
	}
}

func TestUnionTypeDefBasic(t *testing.T) {
	const src = `package main
	
	type A union {
		B int
		C []union {
			D struct { stuff string }
			E float
			F int
		}
	}`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	union := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.UnionType)
	if want, got := 2, len(union.SubTypes.List); want != got {
		t.Errorf("number of subTypes, want %v got %v", want, got)
	}
	if want, got := "B", union.SubTypes.List[0].Names[0].Name; want != got {
		t.Errorf("name of subType 1, want %v got %v", want, got)
	}
	if want, got := "int", union.SubTypes.List[0].Type.(*ast.Ident).Name; want != got {
		t.Errorf("type of subType 1, want %v got %v", want, got)
	}
	if want, got := "C", union.SubTypes.List[1].Names[0].Name; want != got {
		t.Errorf("name of subType 2, want %v got %v", want, got)
	}
	if want, got := 3, len(union.SubTypes.List[1].Type.(*ast.ArrayType).Elt.(*ast.UnionType).SubTypes.List); want != got {
		t.Errorf("number of subTypes for subType 2, want %v got %v", want, got)
	}
}

func TestLambdaBasic(t *testing.T) {
	const src = `package main
	
	func f() {
		l := fn(i) s.ret
	}`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	fn := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.LambdaLit)
	if want, got := 1, len(fn.Params.List); want != got {
		t.Errorf("number of params, want %v got %v", want, got)
	}
	if want, got := "i", fn.Params.List[0].Names[0].Name; want != got {
		t.Errorf("param name, want %v got %v", want, got)
	}
	if want, got := 1, len(fn.Body); want != got {
		t.Errorf("body len, want %v got %v", want, got)
	}
	if want, got := "ret", fn.Body[0].(*ast.SelectorExpr).Sel.Name; want != got {
		t.Errorf("body expr name, want %v got %v", want, got)
	}
}

func TestLambdaMultiAssign(t *testing.T) {
	const src = `package main
	
	func f() {
		l := fn(i) s.ret, 1
	}`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	asn := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)
	if want, got := "l", asn.Lhs[0].(*ast.Ident).Name; want != got {
		t.Errorf("assign lhs 1, want %v got %v", want, got)
	}
	fn := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.LambdaLit)
	if want, got := 1, len(fn.Params.List); want != got {
		t.Errorf("number of params, want %v got %v", want, got)
	}
	if want, got := "i", fn.Params.List[0].Names[0].Name; want != got {
		t.Errorf("param name, want %v got %v", want, got)
	}
	if want, got := 1, len(fn.Body); want != got {
		t.Errorf("body len, want %v got %v", want, got)
	}
	if want, got := "ret", fn.Body[0].(*ast.SelectorExpr).Sel.Name; want != got {
		t.Errorf("body expr name, want %v got %v", want, got)
	}

	lit := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[1].(*ast.BasicLit)
	if want, got := "1", lit.Value; want != got {
		t.Errorf("assignment second rhs, want %v got %v", want, got)
	}
}

func TestLambdaMultiReturn(t *testing.T) {
	const src = `package main
	
	func f() {
		l,r := fn(i) { 100, i, 50 },1
	}`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	asn := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)
	if want, got := "l", asn.Lhs[0].(*ast.Ident).Name; want != got {
		t.Errorf("assign lhs 1, want %v got %v", want, got)
	}
	if want, got := "r", asn.Lhs[1].(*ast.Ident).Name; want != got {
		t.Errorf("assign lhs 2, want %v got %v", want, got)
	}

	fn := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.LambdaLit)
	if want, got := 1, len(fn.Params.List); want != got {
		t.Fatalf("number of params, want %v got %v", want, got)
	}
	if want, got := "i", fn.Params.List[0].Names[0].Name; want != got {
		t.Errorf("param name, want %v got %v", want, got)
	}
	if want, got := 3, len(fn.Body); want != got {
		t.Fatalf("body len, want %v got %v", want, got)
	}
	if want, got := "100", fn.Body[0].(*ast.BasicLit).Value; want != got {
		t.Errorf("body expr 1, want %v got %v", want, got)
	}
	if want, got := "i", fn.Body[1].(*ast.Ident).Name; want != got {
		t.Errorf("body expr 2, want %v got %v", want, got)
	}
	if want, got := "50", fn.Body[2].(*ast.BasicLit).Value; want != got {
		t.Errorf("body expr 3, want %v got %v", want, got)
	}
	lit := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[1].(*ast.BasicLit)
	if want, got := "1", lit.Value; want != got {
		t.Errorf("assignment second rhs, want %v got %v", want, got)
	}
}

func TestParseEnumTypeBasic(t *testing.T) {
	const src = `package main
	
	type e enum { a=1;b;c	}
	type e enum { a,b,c=1,2,3	}
	type e enum int { a=iota;b;c }`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	//enum1 - type e enum { a=1;b;c	}
	enum := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.EnumType)
	if want, got := ast.Expr(nil), enum.Type; want != got {
		t.Errorf("enum 1 type want %v got %v", want, got)
	}
	if want, got := 3, len(enum.Specs); want != got {
		t.Errorf("enum 1 specs want %v got %v", want, got)
	}
	if want, got := "a", enum.Specs[0].(*ast.ValueSpec).Names[0].Name; want != got {
		t.Errorf("enum 1 spec 1 name want %v got %v", want, got)
	}
	if want, got := "1", enum.Specs[0].(*ast.ValueSpec).Values[0].(*ast.BasicLit).Value; want != got {
		t.Errorf("enum 1 spec 1 value want %v got %v", want, got)
	}
	if want, got := "b", enum.Specs[1].(*ast.ValueSpec).Names[0].Name; want != got {
		t.Errorf("enum 1 spec 2 name want %v got %v", want, got)
	}
	if want, got := 0, len(enum.Specs[1].(*ast.ValueSpec).Values); want != got {
		t.Errorf("enum 1 spec 2 value want %v got %v", want, got)
	}

	//enum2 - type e enum { a,b,c=1,2,3	}
	enum = f.Decls[1].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.EnumType)
	if want, got := ast.Expr(nil), enum.Type; want != got {
		t.Errorf("enum 2 type want %v got %v", want, got)
	}
	if want, got := 1, len(enum.Specs); want != got {
		t.Errorf("enum 2 specs want %v got %v", want, got)
	}
	if want, got := "a", enum.Specs[0].(*ast.ValueSpec).Names[0].Name; want != got {
		t.Errorf("enum 2 spec 1 name want %v got %v", want, got)
	}
	if want, got := "1", enum.Specs[0].(*ast.ValueSpec).Values[0].(*ast.BasicLit).Value; want != got {
		t.Errorf("enum 2 spec 1 value want %v got %v", want, got)
	}
	if want, got := "b", enum.Specs[0].(*ast.ValueSpec).Names[1].Name; want != got {
		t.Errorf("enum 2 spec 2 name want %v got %v", want, got)
	}
	if want, got := "2", enum.Specs[0].(*ast.ValueSpec).Values[1].(*ast.BasicLit).Value; want != got {
		t.Errorf("enum 2 spec 2 value want %v got %v", want, got)
	}

	//enum3 - type e enum int { a=iota;b;c }
	enum = f.Decls[2].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.EnumType)
	if want, got := "int", enum.Type.(*ast.Ident).Name; want != got {
		t.Errorf("enum 3 type want %v got %v", want, got)
	}
	if want, got := 3, len(enum.Specs); want != got {
		t.Errorf("enum 3 specs want %v got %v", want, got)
	}
	if want, got := "a", enum.Specs[0].(*ast.ValueSpec).Names[0].Name; want != got {
		t.Errorf("enum 3 spec 1 name want %v got %v", want, got)
	}
	if want, got := "iota", enum.Specs[0].(*ast.ValueSpec).Values[0].(*ast.Ident).Name; want != got {
		t.Errorf("enum 3 spec 1 value want %v got %v", want, got)
	}
	if want, got := "b", enum.Specs[1].(*ast.ValueSpec).Names[0].Name; want != got {
		t.Errorf("enum 3 spec 2 name want %v got %v", want, got)
	}
	if want, got := 0, len(enum.Specs[1].(*ast.ValueSpec).Values); want != got {
		t.Errorf("enum 3 spec 2 value want %v got %v", want, got)
	}
}

func TestGenericTypeDefBasic(t *testing.T) {
	const src = `package main
	
	type A struct<T,V io.Reader> {
		B T
		C string
		D int
	}`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	str := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)
	if want, got := 2, len(str.TypeParams.List); want != got {
		t.Errorf("number of fields, want %v got %v", want, got)
	}
	if want, got := 3, len(str.Fields.List); want != got {
		t.Errorf("number of fields, want %v got %v", want, got)
	}

	if want, got := "T", str.TypeParams.List[0].Names[0].Name; want != got {
		t.Errorf("name of typeparam 1, want %v got %v", want, got)
	}
	if want, got := ast.Expr(nil), str.TypeParams.List[0].Type; want != got {
		t.Errorf("type of typeparam 1, want %v got %v", want, got)
	}
	if want, got := "V", str.TypeParams.List[1].Names[0].Name; want != got {
		t.Errorf("name of typeparam 2, want %v got %v", want, got)
	}
	if want, got := "Reader", str.TypeParams.List[1].Type.(*ast.SelectorExpr).Sel.Name; want != got {
		t.Errorf("type of typeparam 2, want %v got %v", want, got)
	}

	if want, got := "B", str.Fields.List[0].Names[0].Name; want != got {
		t.Errorf("name of field 1, want %v got %v", want, got)
	}
	if want, got := "T", str.Fields.List[0].Type.(*ast.Ident).Name; want != got {
		t.Errorf("type of field 1, want %v got %v", want, got)
	}

	if want, got := "C", str.Fields.List[1].Names[0].Name; want != got {
		t.Errorf("type of field 2, want %v got %v", want, got)
	}
	if want, got := "string", str.Fields.List[1].Type.(*ast.Ident).Name; want != got {
		t.Errorf("type of field 2, want %v got %v", want, got)
	}

	if want, got := "D", str.Fields.List[2].Names[0].Name; want != got {
		t.Errorf("type of field 3, want %v got %v", want, got)
	}
	if want, got := "int", str.Fields.List[2].Type.(*ast.Ident).Name; want != got {
		t.Errorf("type of field 3, want %v got %v", want, got)
	}
}

func TestGenericInterfaceDefBasic(t *testing.T) {
	const src = `package main
	
	type A interface<T,V io.Reader> {
		m() T
		a V
		b io.Reader
	}`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	str := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.InterfaceType)
	if want, got := 2, len(str.TypeParams.List); want != got {
		t.Errorf("number of fields, want %v got %v", want, got)
	}
	if want, got := 3, len(str.Fields.List); want != got {
		t.Errorf("number of fields, want %v got %v", want, got)
	}

	if want, got := "T", str.TypeParams.List[0].Names[0].Name; want != got {
		t.Errorf("name of typeparam 1, want %v got %v", want, got)
	}
	if want, got := ast.Expr(nil), str.TypeParams.List[0].Type; want != got {
		t.Errorf("type of typeparam 1, want %v got %v", want, got)
	}
	if want, got := "V", str.TypeParams.List[1].Names[0].Name; want != got {
		t.Errorf("name of typeparam 2, want %v got %v", want, got)
	}
	if want, got := "Reader", str.TypeParams.List[1].Type.(*ast.SelectorExpr).Sel.Name; want != got {
		t.Errorf("type of typeparam 2, want %v got %v", want, got)
	}

	if want, got := "m", str.Fields.List[0].Names[0].Name; want != got {
		t.Errorf("name of field 1, want %v got %v", want, got)
	}
	if want, got := "T", str.Fields.List[0].Type.(*ast.FuncType).Results.List[0].Type.(*ast.Ident).Name; want != got {
		t.Errorf("type of field 1, want %v got %v", want, got)
	}

	if want, got := "a", str.Fields.List[1].Names[0].Name; want != got {
		t.Errorf("name of field 2, want %v got %v", want, got)
	}
	if want, got := "V", str.Fields.List[1].Type.(*ast.Ident).Name; want != got {
		t.Errorf("type of field 2, want %v got %v", want, got)
	}

	if want, got := "b", str.Fields.List[2].Names[0].Name; want != got {
		t.Errorf("type of field 3, want %v got %v", want, got)
	}
	if want, got := "Reader", str.Fields.List[2].Type.(*ast.SelectorExpr).Sel.Name; want != got {
		t.Errorf("type of field 3, want %v got %v", want, got)
	}
}

func TestGenericTypeUseBasic(t *testing.T) {
	const src = `package main
	
	func f<T>(in List<T>) T { 
		var a pkg.Test<int,pkg.Some<T>>
	}
	`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	fun := f.Decls[0].(*ast.FuncDecl)
	if want, got := 1, len(fun.Type.TypeParams.List); want != got {
		t.Errorf("func type param count, want %v got %v", want, got)
	}
	if want, got := 1, len(fun.Type.Params.List); want != got {
		t.Errorf("func param count, want %v got %v", want, got)
	}
	if want, got := "in", fun.Type.Params.List[0].Names[0].Name; want != got {
		t.Errorf("func param name, want %v got %v", want, got)
	}

	intype := fun.Type.Params.List[0].Type.(*ast.Ident)
	if want, got := "List", intype.Name; want != got {
		t.Errorf("func param type name, want %v got %v", want, got)
	}
	if want, got := 1, len(intype.TypeArgs); want != got {
		t.Errorf("func param type param count, want %v got %v", want, got)
	}
	if want, got := "T", intype.TypeArgs[0].(*ast.Ident).Name; want != got {
		t.Errorf("func param type param name, want %v got %v", want, got)
	}

	sel := fun.Body.List[0].(*ast.DeclStmt).Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Type.(*ast.SelectorExpr)
	if want, got := "pkg", sel.X.(*ast.Ident).Name; want != got {
		t.Errorf("body type package name, want %v got %v", want, got)
	}
	if want, got := "Test", sel.Sel.Name; want != got {
		t.Errorf("body type name, want %v got %v", want, got)
	}
	if want, got := 2, len(sel.Sel.TypeArgs); want != got {
		t.Errorf("body type param count, want %v got %v", want, got)
	}
	if want, got := "int", sel.Sel.TypeArgs[0].(*ast.Ident).Name; want != got {
		t.Errorf("body type param 1 type, want %v got %v", want, got)
	}
	if want, got := "Some", sel.Sel.TypeArgs[1].(*ast.SelectorExpr).Sel.Name; want != got {
		t.Errorf("body type param 2 type, want %v got %v", want, got)
	}
	if want, got := 1, len(sel.Sel.TypeArgs[1].(*ast.SelectorExpr).Sel.TypeArgs); want != got {
		t.Errorf("body type param 2 subtype, want %v got %v", want, got)
	}
}

func TestGenericTypeUseLiteral(t *testing.T) {
	const src = `package main
	
	var b t<struct{sup int}>
	`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	val := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
	if want, got := "b", val.Names[0].Name; want != got {
		t.Errorf("var1 name, want %v got %v", want, got)
	}
	if want, got := "t", val.Type.(*ast.Ident).Name; want != got {
		t.Errorf("var1 type name, want %v got %v", want, got)
	}

	if want, got := 1, len(val.Type.(*ast.Ident).TypeArgs); want != got {
		t.Errorf("val1 type param count, want %v got %v", want, got)
	}

	if want, got := 1, len(val.Type.(*ast.Ident).TypeArgs[0].(*ast.StructType).Fields.List); want != got {
		t.Errorf("val1 type param struct field count, want %v got %v", want, got)
	}
}

func TestGenericFuncUseBasic(t *testing.T) {
	const src = `package main
	
	func main() {
		f(<T<interface{blah int}>, int> in1, in2)
	}
	`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	call := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt).X.(*ast.CallExpr)
	if want, got := "f", call.Fun.(*ast.Ident).Name; want != got {
		t.Errorf("func name, want %v got %v", want, got)
	}
	if want, got := 2, len(call.TypeArgs); want != got {
		t.Errorf("func type param count, want %v got %v", want, got)
	}
	if want, got := "T", call.TypeArgs[0].(*ast.Ident).Name; want != got {
		t.Errorf("func type param 1 name, want %v got %v", want, got)
	}
	if want, got := 1, len(call.TypeArgs[0].(*ast.Ident).TypeArgs); want != got {
		t.Errorf("func type param 1 subtype count, want %v got %v", want, got)
	}
	if want, got := 1, call.TypeArgs[0].(*ast.Ident).TypeArgs[0].(*ast.InterfaceType).Fields.NumFields(); want != got {
		t.Errorf("func type param 1, subparam 1 interface field count, want %v got %v", want, got)
	}

}

var b = m{1, "test"}

type m struct {
	a int
	z string
}

func TestGenericStructLiteral(t *testing.T) {
	const src = `package main
	
	var b = m{<int,optional<string>> 1, "test" }
	type m struct<K,V> { a K; z V; }
	`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	val := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
	if want, got := "b", val.Names[0].Name; want != got {
		t.Errorf("var1 name, want %v got %v", want, got)
	}

	lit := val.Values[0].(*ast.CompositeLit)
	if want, got := "m", lit.Type.(*ast.Ident).Name; want != got {
		t.Errorf("var1 type name, want %v got %v", want, got)
	}
	if want, got := 2, len(lit.Elts); want != got {
		t.Errorf("var1 lit value count, want %v got %v", want, got)
	}
	if want, got := "1", lit.Elts[0].(*ast.BasicLit).Value; want != got {
		t.Errorf("var1 lit value 1, want %v got %v", want, got)
	}
	if want, got := "\"test\"", lit.Elts[1].(*ast.BasicLit).Value; want != got {
		t.Errorf("var1 lit value 2, want %v got %v", want, got)
	}
}

func TestGenericStructMapLiteral(t *testing.T) {
	const src = `package main
	
	var b = map{<int,string>
		 1 : "test",
	}
	`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	val := f.Decls[0].(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
	if want, got := "b", val.Names[0].Name; want != got {
		t.Errorf("var1 name, want %v got %v", want, got)
	}

	lit := val.Values[0].(*ast.CompositeLit)
	if want, got := "map", lit.Type.(*ast.Ident).Name; want != got {
		t.Errorf("var1 type name, want %v got %v", want, got)
	}
	if want, got := 1, len(lit.Elts); want != got {
		t.Errorf("var1 lit value count, want %v got %v", want, got)
	}
	if want, got := "1", lit.Elts[0].(*ast.KeyValueExpr).Key.(*ast.BasicLit).Value; want != got {
		t.Errorf("var1 lit key 1, want %v got %v", want, got)
	}
	if want, got := "\"test\"", lit.Elts[0].(*ast.KeyValueExpr).Value.(*ast.BasicLit).Value; want != got {
		t.Errorf("var1 lit value 1, want %v got %v", want, got)
	}
}
