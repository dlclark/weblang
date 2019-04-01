package compiler

import (
	"testing"
	"weblang/code/lexer"
	"weblang/code/parser"
)

func TestHelloWorldVar(t *testing.T) {
	output := compileProgram(t, `var message = "Hello World!"`)

	if want, got := `let message = "Hello World!";`, output; want != got {
		t.Fatalf("output wanted:\n%v\ngot:\n%v", want, got)
	}
}

func TestIfVar(t *testing.T) {
	output := compileProgram(t, `
var test = "Hello"
if test == "World" { }`)

	if want, got := `let test = "Hello";
if (test === "World") {
};`, output; want != got {
		t.Fatalf("output wanted:\n%v\ngot:\n%v", want, got)
	}
}

func TestIfVarSet(t *testing.T) {
	output := compileProgram(t, `
var test = "Hello"
if test == "Hello" { 
    test = "World"
 }`)

	if want, got := `let test = "Hello";
if (test === "Hello") {
test = "World";
};`, output; want != got {
		t.Fatalf("output wanted:\n%v\ngot:\n%v", want, got)
	}
}

func compileProgram(t *testing.T, program string) string {
	l := lexer.New(program, "junk")
	p := parser.New(l)

	tree := p.ParseProgram()
	checkParserErrors(t, p)

	c := New()
	err := c.Compile(tree)

	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	return c.Output()
}

func checkParserErrors(t *testing.T, p *parser.Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
