package jsprinter

import (
	"fmt"
	"io"
	"strings"
	"weblang/wl/jscompiler/jsast"
)

type jsPrinter struct {
	output io.Writer

	justPrintedEndStmt bool
}

// Fprint writes the javascript into the out writer.
// The node param must be one of:
// 		- *jsast.Module
//		- assignable to jsast.Expr, jsast.Decl, jsast.Stmt
//		- or []jsast.Stmt, []jsast.Decl
func Fprint(out io.Writer, node interface{}) error {
	p := &jsPrinter{output: out}
	return p.Fprint(node)
}

func (p *jsPrinter) Fprint(node interface{}) error {
	switch n := node.(type) {
	case jsast.Expr:
		p.expr(n)
	case jsast.Stmt:
		p.stmt(n)
	case jsast.Decl:
		p.decl(n)
	case []jsast.Stmt:
		p.stmtList(n)
	case []jsast.Decl:
		p.declList(n)
	case *jsast.Module:
		p.module(n)
	default:
		return fmt.Errorf("jsprinter: unsupported node type %T", node)
	}
	return nil
}

func (p *jsPrinter) expr(expr jsast.Expr) {
	switch x := expr.(type) {
	case *jsast.Placeholder:
		p.placeholder(x)
	case *jsast.RawJs:
		p.print(x.RawJs)
	case *jsast.Identifier:
		p.print(x.Name)
	case *jsast.BinaryExpression:
		p.expr(x.Lhs)
		p.print(" ", x.Op, " ")
		p.expr(x.Rhs)
	case *jsast.UnaryExpression:
		p.print(x.Op)
		p.expr(x.Exp)
	case *jsast.FunctionLiteral:
		p.print("function ")
		if x.Name != nil {
			p.print(*x.Name)
		}
		p.print("(")
		// print params
		for i, param := range x.Params {
			if i > 0 {
				p.print(", ")
			}
			p.print(param)
		}
		p.print(") {\n")
		p.stmtList(x.Body)
		p.print("}")
	case *jsast.BasicLiteral:
		p.print(x.Value)
	case *jsast.SelectorExpr:
		p.expr(x.X)
		p.print(".", x.Sel)
	case *jsast.ClassInstantiate:
		p.print("new ", x.ClassName, "(")
		for i, param := range x.CtorParams {
			if i > 0 {
				p.print(", ")
			}
			p.expr(param)
		}
		p.print(")")
	default:
		panic(fmt.Sprintf("jsprinter: unsupported node type: %T", expr))
	}
}

func (p *jsPrinter) stmtList(list []jsast.Stmt) {
	for _, stmt := range list {
		p.stmt(stmt)
	}
}

func (p *jsPrinter) stmt(stmt jsast.Stmt) {
	switch x := stmt.(type) {
	case *jsast.Placeholder:
		p.placeholder(x)
	case *jsast.RawJs:
		p.print(x.RawJs)
	case *jsast.ExprStmt:
		p.expr(x.Exp)
	case *jsast.ReturnStmt:
		p.print("return ")
		p.expr(x.Result)
	case *jsast.BlockStmt:
		p.print("{\n")
		p.stmtList(x.Body)
		p.print("}")
	case *jsast.IfStmt:
		p.print("if (")
		p.expr(x.Cond)
		p.print(") ")
		p.stmt(x.Body)
		if x.Else != nil {
			p.print("else ")
			p.stmt(x.Else)
		}
	case *jsast.DeclStmt:
		p.decl(x.Decl)
	case *jsast.AssignStmt:
		p.expr(x.Lhs)
		p.print(" ", x.Op, " ")
		p.expr(x.Rhs)
	default:
		panic(fmt.Sprintf("jsprinter: unsupported node type: %T", stmt))
	}
	//statements end in semicolons and newlines
	p.printEndStatement()
}

func (p *jsPrinter) declList(list []jsast.Decl) {
	for _, decl := range list {
		p.decl(decl)
	}
}

func (p *jsPrinter) decl(decl jsast.Decl) {
	switch x := decl.(type) {
	case *jsast.Placeholder:
		p.placeholder(x)
	case *jsast.ClassDecl:
		if x.IsExported {
			p.print("export ")
		}
		p.print("class ", x.Name, " {\n")
		//print fields
		for _, f := range x.Fields {
			f.Kind = ""
			p.decl(f)
		}
		//print functions
		for _, f := range x.Methods {
			p.decl(f)
		}
		p.print("};\n")
	case *jsast.FuncDecl:
		if x.IsExported {
			p.print("export ")
		}
		p.expr(&x.Func)
		p.printEndStatement()
	case *jsast.VarDecl:
		if x.IsExported {
			p.print("export ")
		}
		p.print(x.Kind, " ", x.Name)
		if x.Value != nil {
			p.print(" = ")
			p.expr(x.Value)
		}
		p.printEndStatement()
	default:
		panic(fmt.Sprintf("jsprinter: unsupported node type: %T", decl))
	}
}

func (p *jsPrinter) module(mod *jsast.Module) {
	for _, i := range mod.Imports {
		p.print("import * as ", i.Alias, " from \"", i.File, "\";\n")
	}

	// add spaces after imports
	if len(mod.Imports) > 0 {
		p.print("\n\n")
	}

	p.declList(mod.Decls)
}

func (p *jsPrinter) placeholder(n *jsast.Placeholder) {
	for _, c := range n.Children {
		p.Fprint(c)
	}
}

func (p *jsPrinter) printEndStatement() {
	if p.justPrintedEndStmt {
		return
	}
	p.print(";\n")
}

func (p *jsPrinter) print(s ...string) {
	for i := range s {
		io.WriteString(p.output, s[i])
	}
	p.justPrintedEndStmt = false

	if len(s) > 0 && strings.HasSuffix(s[len(s)-1], ";\n") {
		p.justPrintedEndStmt = true
	}
}
