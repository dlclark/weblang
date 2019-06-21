package jscompiler

import (
	"io"
	"weblang/wl/ast"
	"weblang/wl/jscompiler/jsprinter"
	"weblang/wl/types"
)

// Outputer provides writers for the packages we're compiling
type Outputer interface {
	WriterFor(pkg *types.Package) io.Writer
	Done(pkg *types.Package, writer io.Writer)
}

// Compile takes a package as a set of ast files and type information
// and uses the given outputer to write it to files
func Compile(pkg *types.Package, ast []*ast.File, out Outputer) error {

	c := &jsCompiler{
		symbols: &symbolMap{
			store: make(map[string]string),
		},
	}

	jsmodule, err := c.Compile(pkg, ast)
	if err != nil {
		return err
	}

	writer := out.WriterFor(pkg)
	if err := jsprinter.Fprint(writer, jsmodule); err != nil {
		return err
	}
	out.Done(pkg, writer)
	return nil
}
