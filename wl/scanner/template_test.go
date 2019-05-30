package scanner

import (
	"testing"
	"weblang/wl/token"
)

func TestTemplateStringScanBasic(t *testing.T) {
	var s Scanner
	src := "package p; var a = `testing ${test} blah`"
	file := fset.AddFile("interfaceTest", fset.Base(), len(src))
	s.Init(file, []byte(src), func(pos token.Position, msg string) { t.Error(Error{pos, msg}) }, 0)

	want := []struct {
		t   token.Token
		lit string
		s   bool
	}{
		{t: token.PACKAGE, lit: "package"}, //0
		{t: token.IDENT, lit: "p"},
		{t: token.SEMICOLON, lit: ";"},
		{t: token.VAR, lit: "var"},
		{t: token.IDENT, lit: "a"},
		{t: token.ASSIGN, lit: ""},
		{t: token.TEMPLATE, lit: "`"},
		{t: token.STRING, lit: `testing `, s: true}, // 7
		{t: token.TEMPLATEEXPR, lit: "${", s: true},
		{t: token.IDENT, lit: "test"},
		{t: token.RBRACE, lit: ""},
		{t: token.STRING, lit: ` blah`, s: true}, // 11
		{t: token.TEMPLATE, lit: "`", s: true},
		{t: token.SEMICOLON, lit: "\n"},
	}
	//(pos token.Pos, tok token.Token, lit string)
	for i := 0; i < len(want); i++ {
		var tok token.Token
		var lit string
		if want[i].s {
			_, tok, lit = s.ScanTemplateString()
		} else {
			_, tok, lit = s.Scan()
		}

		if tok != want[i].t {
			t.Fatalf("Wanted token %v but got %v at index %v", want[i].t, tok, i)
		}
		if lit != want[i].lit {
			t.Fatalf("Wanted literal %v but got %v at index %v", want[i].lit, lit, i)
		}
	}
	if _, tok, _ := s.Scan(); tok != token.EOF {
		t.Fatalf("had tokens after expected set: %v", tok)
	}

}
