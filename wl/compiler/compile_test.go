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

func TestEnum(t *testing.T) {
	output := compileProgram(t, `
type test enum {
    None = 0
    Blah = 1
    Yu = 2
}

var v = test.Blah
if v == test.Yu { 
    return false
}`)

	if want, got := `const test = Object.freeze({
None: { name: "None", value: 0 },
Blah: { name: "Blah", value: 1 },
Yu: { name: "Yu", value: 2 }
});
let v = test.Blah;
if (v === test.Yu) {
return false;
};`, output; want != got {
		t.Fatalf("output wanted:\n%v\ngot:\n%v", want, got)
	}
}

func TestStruct(t *testing.T) {
	output := compileProgram(t, `
type Test struct {
    val int
    val2 string
}

var v Test
if v.val == 100 { 
    return false
}`)

	if want, got := `class Test {
constructor(val, val2) {
this.val = val;
this.val2 = val2;
};
};
let v = new Test();
if (v.val === 100) {
return false;
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
