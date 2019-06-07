package types_test

import (
	"fmt"
	"testing"
	//. "weblang/wl/types"
)

func TestBasicStructGenerics(t *testing.T) {
	pkg, err := check(t, `package a; type s struct<T>{a T}; var v1 s<string> = s{"test"}; var v2 s<int> = s{1};`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	scope := pkg.Scope()
	if want, got := "a.s", scope.Lookup("s").Type().String(); want != got {
		t.Errorf("type s, wanted %v got %v", want, got)
	}
	if want, got := "a.s", scope.Lookup("v1").Type().String(); want != got {
		t.Errorf("type v1, wanted %v got %v", want, got)
	}
	if want, got := "a.s", scope.Lookup("v2").Type().String(); want != got {
		t.Errorf("type v2, wanted %v got %v", want, got)
	}

	fmt.Printf("v1: %#v\n", scope.Lookup("v1"))
	fmt.Printf("v2: %#v\n", scope.Lookup("v2"))
	t.Fail()
}
