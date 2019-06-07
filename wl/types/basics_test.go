package types_test

import (
	"testing"

	"weblang/wl/ast"
	"weblang/wl/importer"
	. "weblang/wl/types"
)

func TestBasicConstTypeInference(t *testing.T) {
	pkg, err := check(t, `package a; var i = 1; var f = 1.2; var s = "s"`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	scope := pkg.Scope()
	if want, got := "int", scope.Lookup("i").Type().String(); want != got {
		t.Errorf("type i, wanted %v got %v", want, got)
	}
	if want, got := "float", scope.Lookup("f").Type().String(); want != got {
		t.Errorf("type f, wanted %v got %v", want, got)
	}
	if want, got := "string", scope.Lookup("s").Type().String(); want != got {
		t.Errorf("type s, wanted %v got %v", want, got)
	}
}

func TestBasicStructLitCheck(t *testing.T) {
	pkg, err := check(t, `package a; type s struct{a string}; var v = s{"test"};`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	scope := pkg.Scope()
	if want, got := "a.s", scope.Lookup("s").Type().String(); want != got {
		t.Errorf("type s, wanted %v got %v", want, got)
	}
	if want, got := "a.s", scope.Lookup("v").Type().String(); want != got {
		t.Errorf("type v, wanted %v got %v", want, got)
	}
}

func check(t *testing.T, src string) (*Package, error) {
	f := mustParse(t, src)
	conf := Config{Importer: importer.Default()}
	return conf.Check(f.Name.Name, fset, []*ast.File{f}, nil)
}
