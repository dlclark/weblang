package parser

import (
	"os"
	"testing"
	"weblang/wl/ast"
	"weblang/wl/token"
)

func TestTemplateParseBasic(t *testing.T) {
	const src = `package main
		
		func main() {
			a := ` + "`this ${test} a`" + `
		}`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	str := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.TemplateExprLit)
	if want, got := 3, len(str.Parts); want != got {
		t.Fatalf("parts, want %v got %v", want, got)
	}
	if want, got := "\"this \"", str.Parts[0].(*ast.BasicLit).Value; want != got {
		t.Errorf("first part text, want %v got %v", want, got)
	}
	if want, got := "test", str.Parts[1].(*ast.Ident).Name; want != got {
		t.Errorf("second part ident, want %v got %v", want, got)
	}
	if want, got := "\" a\"", str.Parts[2].(*ast.BasicLit).Value; want != got {
		t.Errorf("third part text, want %v got %v", want, got)
	}
}

func TestTemplateParseExpr(t *testing.T) {
	const src = `package main
		
		func main() {
			a := ` + "`this \" ${test+1} a`" + `
		}`

	fset := token.NewFileSet()
	f, err := ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := ast.Fprint(os.Stdout, fset, f, nil); err != nil {
		t.Fatal(err)
	}

	str := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.TemplateExprLit)
	if want, got := 3, len(str.Parts); want != got {
		t.Fatalf("parts, want %v got %v", want, got)
	}
	if want, got := "this \" ", str.Parts[0].(*ast.BasicLit).Value; want != got {
		t.Errorf("first part text, want %v got %v", want, got)
	}
	if want, got := "test", str.Parts[1].(*ast.BinaryExpr).X.(*ast.Ident).Name; want != got {
		t.Errorf("second part ident, want %v got %v", want, got)
	}
	if want, got := token.ADD, str.Parts[1].(*ast.BinaryExpr).Op; want != got {
		t.Errorf("second part op, want %v got %v", want, got)
	}
	if want, got := "1", str.Parts[1].(*ast.BinaryExpr).Y.(*ast.BasicLit).Value; want != got {
		t.Errorf("second part y, want %v got %v", want, got)
	}
	if want, got := " a", str.Parts[2].(*ast.BasicLit).Value; want != got {
		t.Errorf("third part text, want %v got %v", want, got)
	}
}
