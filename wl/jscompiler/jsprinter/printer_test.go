package jsprinter

/*
import (
	"testing"
	"weblang/wl/ast"
	"weblang/wl/importer"
	"weblang/wl/parser"
	"weblang/wl/token"
	"weblang/wl/types"
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

func compileProgram(t *testing.T, src string) string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)

	if err != nil {
		t.Fatalf("Error during parse: %v", err)
	}

	// typecheck
	conf := types.Config{Importer: importer.Default()}
	_, err = conf.Check(f.Name.Name, fset, []*ast.File{f}, nil)
	if err != nil {
		t.Fatalf("Error During Type Check: %v", err)
	}

	c := New()
	err = c.Compile(f)

	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	return c.Output()
}
*/
