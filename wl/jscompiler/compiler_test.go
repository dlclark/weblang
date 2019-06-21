package jscompiler

import (
	"io"
	"strings"
	"testing"
	"weblang/wl/ast"
	"weblang/wl/importer"
	"weblang/wl/parser"
	"weblang/wl/token"
	"weblang/wl/types"
)

func TestHelloWorldVar(t *testing.T) {
	output := compileProgram(t, `package p; var message = "Hello World!"`)

	if want, got := `let message = "Hello World!";`, output; want != got {
		t.Fatalf("output wanted:\n`%v`\ngot:\n`%v`", want, got)
	}
}

func TestIfVar(t *testing.T) {
	output := compileProgram(t, `
package p
func a() {
	var test = "Hello"
	if test == "World" { }
}`)

	if want, got := `function a() {
let test = "Hello";
if (test === "World") {
};
};`, output; want != got {
		t.Fatalf("output wanted:\n%v\ngot:\n%v", want, got)
	}
}

func TestIfVarSet(t *testing.T) {
	output := compileProgram(t, `
package p
func a() {
	var test = "Hello"
	if test == "Hello" { 
    	test = "World"
	 }
}`)

	if want, got := `function a() {
let test = "Hello";
if (test === "Hello") {
test = "World";
};
};`, output; want != got {
		t.Fatalf("output wanted:\n%v\ngot:\n%v", want, got)
	}
}

func TestEnum(t *testing.T) {
	output := compileProgram(t, `
package p

type test enum {
    None = 0
    Blah = 1
    Yu = 2
}

func a() {
	var v = test.Blah
	if v == test.Yu { 
    	return false
	}
}`)

	if want, got := `const test = Object.freeze({
None: { name: "None", value: 0 },
Blah: { name: "Blah", value: 1 },
Yu: { name: "Yu", value: 2 }
});
function a() {
let v = test.Blah;
if (v === test.Yu) {
return false;
};
};`, output; want != got {
		t.Fatalf("output wanted:\n%v\ngot:\n%v", want, got)
	}
}

func TestStruct(t *testing.T) {
	output := compileProgram(t, `
package p

type Test struct {
    val int
    val2 string
}

func a() bool {
	var v Test
	if v.val == 100 { 
		return false
	}
	return true
}`)

	if want, got := `class Test {
 val;
 val2;
};
function a() {
let v = new Test();
if (v.val === 100) {
return false;
};
return true;
};`, output; want != got {
		t.Fatalf("output wanted:\n%v\ngot:\n%v", want, got)
	}
}

func compileProgram(t *testing.T, src string) string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.wl", src, 0)

	if err != nil {
		t.Fatalf("Error during parse: %v", err)
	}

	astF := []*ast.File{f}

	// typecheck
	conf := types.Config{Importer: importer.Default()}
	pkg, err := conf.Check(f.Name.Name, fset, astF, nil)
	if err != nil {
		t.Fatalf("Error During Type Check: %v", err)
	}

	out := newTestOutputer(t, 1)
	err = Compile(pkg, astF, out)

	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	return out.Output()
}

type testOutputer struct {
	output               map[string]string
	pending              []string
	expectedPackageCount int
	t                    *testing.T
}

func newTestOutputer(t *testing.T, expectedPackageCount int) *testOutputer {
	return &testOutputer{
		output:               make(map[string]string),
		expectedPackageCount: expectedPackageCount,
		t:                    t,
	}
}
func (o *testOutputer) WriterFor(pkg *types.Package) io.Writer {
	if len(o.pending)+len(o.output) > o.expectedPackageCount {
		o.t.Fatalf("expected %v package(s)", o.expectedPackageCount)
	}
	name := pkg.Name()
	o.pending = append(o.pending, name)
	return &strings.Builder{}
}
func (o *testOutputer) Done(pkg *types.Package, writer io.Writer) {
	name := pkg.Name()
	buf := writer.(*strings.Builder)
	// trim off start/end whitespace to make testing easier
	o.output[name] = strings.TrimSpace(buf.String())
}
func (o *testOutputer) Output() string {
	if len(o.pending) != 1 {
		o.t.Fatalf("expected 1 output file, but got %v", len(o.pending))
	}
	if out, ok := o.output[o.pending[0]]; ok {
		return out
	}
	o.t.Fatalf("did not have output file for package '%v'", o.pending[0])
	return ""
}
