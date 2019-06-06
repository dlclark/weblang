// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types_test

import (
	"testing"
	"weblang/wl/parser"

	. "weblang/wl/types"
)

var testExprs = []testEntry{
	// basic type literals
	dup("x"),
	dup("true"),
	dup("42"),
	dup("3.1415"),
	dup(`"foo"`),
	dup("`bar`"),

	// func and composite literals
	{"func(){}", "(func() literal)"},
	{"func(x int) string {}", "(func(x int) string literal)"},
	{"[]int{1, 2, 3}", "([]int literal)"},

	// non-type expressions
	dup("(x)"),
	dup("x.f"),
	dup("a[i]"),

	dup("s[:]"),
	dup("s[i:]"),
	dup("s[:j]"),
	dup("s[i:j]"),
	dup("s[:j:k]"),
	dup("s[i:j:k]"),

	dup("x.(T)"),

	dup("x.([10]int)"),
	dup("x.([...]int)"),

	dup("x.(struct{})"),
	dup("x.(struct{x int; y, z float; E})"),

	dup("x.(func())"),
	dup("x.(func(x int))"),
	dup("x.(func() int)"),
	dup("x.(func(x, y int, z float) (r int))"),
	dup("x.(func(a, b, c int))"),
	dup("x.(func(x ...T))"),

	dup("x.(interface{})"),
	dup("x.(interface{m(); n(x int); E})"),
	dup("x.(interface{m(); n(x int) T; E; F})"),

	//	dup("x.(chan E)"),
	//	dup("x.(<-chan E)"),
	//	dup("x.(chan<- chan int)"),
	//	dup("x.(chan<- <-chan int)"),
	//	dup("x.(<-chan chan int)"),
	//	dup("x.(chan (<-chan int))"),

	dup("f()"),
	dup("f(x)"),
	dup("int(x)"),
	dup("f(x, x + y)"),
	dup("f(s...)"),
	dup("f(a, s...)"),

	//	dup("*x"),
	//	dup("&x"),
	dup("x + y"),
	dup("x + y << (2 * s)"),

	// generics
	dup("x.(map<K, V>)"),
	dup("f(<int> 1)"),
	dup("x.(struct<T>{x T; y, z float; E})"),
	dup("x.(interface<T>{m(x T); n T})"),
	dup("x.(func<T>(a int, b T) T)"),
}

func TestExprString(t *testing.T) {
	for _, test := range testExprs {
		x, err := parser.ParseExpr(test.src)
		if err != nil {
			t.Errorf("%s: %s", test.src, err)
			continue
		}
		if got := ExprString(x); got != test.str {
			t.Errorf("%s: got %s, want %s", test.src, got, test.str)
		}
	}
}

type testEntry struct {
	src, str string
}

// dup returns a testEntry where both src and str are the same.
func dup(s string) testEntry {
	return testEntry{s, s}
}
