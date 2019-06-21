package jscompiler

import (
	"fmt"
	"weblang/wl/ast"
	"weblang/wl/jscompiler/jsast"
	"weblang/wl/token"
	"weblang/wl/types"
)

type jsCompiler struct {
	symbols *symbolMap
}

func (c *jsCompiler) Compile(pkg *types.Package, files []*ast.File) (*jsast.Module, error) {
	m := &jsast.Module{
		Name: pkg.Name(),
	}

	// validate all our files are the same package...
	for _, f := range files {
		if m.Name != f.Name.Name {
			return nil, fmt.Errorf("all files must be for the same package.  Expected '%v' but got '%v'", m.Name, f.Name.Name)
		}

		for _, i := range f.Imports {
			if i.Name != nil {
				// TODO: make sure all files have the same aliases
				// for the same imports...
				// we dont support having both:
				// 		file A with import a "fmt" and
				//		file B with import a "testing"
			}
		}
	}

	// setup our imports
	for _, imp := range pkg.Imports() {
		m.Imports = append(m.Imports, jsast.Import{
			Alias: imp.Name(),
			File:  imp.Path(),
		})
	}

	// iterate the files ASTs and compile them one at a time
	for _, f := range files {
		for _, d := range f.Decls {
			m.Decls = append(m.Decls, c.convertDecl(d))
		}
	}

	return m, nil
}

func (c *jsCompiler) convertNode(node ast.Node) jsast.Node {
	switch n := node.(type) {
	case ast.Decl:
		return c.convertDecl(n)
	case ast.Stmt:
		return c.convertStmt(n)
	case ast.Expr:
		return c.convertExpr(n)
	}

	panic(fmt.Sprintf("unexpected node type: %T", node))
}

func (c *jsCompiler) convertDecl(decl ast.Decl) jsast.Decl {
	if decl == nil {
		return nil
	}

	switch n := decl.(type) {
	case *ast.GenDecl:
		if len(n.Specs) == 1 {
			return c.convertSpec(n.Specs[0], n.Tok)
		}

		var sub []jsast.Node
		for _, s := range n.Specs {
			sub = append(sub, c.convertSpec(s, n.Tok))
		}
		return &jsast.Placeholder{sub}
	case *ast.FuncDecl:
		if n.Recv != nil {
			panic("method receivers not yet supported")
		}

		//TODO: isExported
		return &jsast.FuncDecl{
			Func: c.convertFunc(n.Name, n.Type, n.Body.List),
		}

	}

	panic(fmt.Sprintf("Unknown decl node type: %T", decl))
}

func (c *jsCompiler) convertFunc(name *ast.Ident, def *ast.FuncType, body []ast.Stmt) jsast.FunctionLiteral {
	fun := jsast.FunctionLiteral{}
	// convert the name
	if name != nil {
		nm := c.getJsIdent(name)
		fun.Name = &nm
	}

	// convert input params
	for _, p := range def.Params.List {
		for _, n := range p.Names {
			fun.Params = append(fun.Params, c.getJsIdent(n))
		}
	}

	// convert the body
	for _, s := range body {
		fun.Body = append(fun.Body, c.convertStmt(s))
	}

	return fun
}

func (c *jsCompiler) convertStmt(stmt ast.Stmt) jsast.Stmt {
	if stmt == nil {
		return nil
	}

	switch n := stmt.(type) {
	case *ast.DeclStmt:
		return &jsast.DeclStmt{Decl: c.convertDecl(n.Decl)}
	case *ast.IfStmt:
		return &jsast.IfStmt{
			Cond: c.convertExpr(n.Cond),
			Body: c.convertStmt(n.Body).(*jsast.BlockStmt),
			Else: c.convertStmt(n.Else),
		}
	case *ast.BlockStmt:
		var sub []jsast.Stmt
		for _, s := range n.List {
			sub = append(sub, c.convertStmt(s))
		}
		return &jsast.BlockStmt{Body: sub}
	case *ast.AssignStmt:
		//TODO: define token?
		if n.Tok == token.DEFINE {
			panic("define assignment not supported")
		}

		if len(n.Lhs) != len(n.Rhs) {
			//TODO: if RHS is a function call to a function with multiple
			// return values then we need to de-structure it: {a, b} =
			panic("function multi-return-value assignment not supported")
		}

		var sub []jsast.Node
		for i := range n.Lhs {
			sub = append(sub, &jsast.AssignStmt{
				Lhs: c.convertExpr(n.Lhs[i]),
				Op:  c.convertOp(n.Tok),
				Rhs: c.convertExpr(n.Rhs[i]),
			})
		}
		if len(sub) == 1 {
			return sub[0].(jsast.Stmt)
		}
		return &jsast.Placeholder{Children: sub}
	case *ast.ReturnStmt:
		if len(n.Results) == 0 {
			return &jsast.ReturnStmt{}
		}

		if len(n.Results) > 1 {
			panic("multi-return not supported yet")
		}
		return &jsast.ReturnStmt{
			Result: c.convertExpr(n.Results[0]),
		}
	}

	panic(fmt.Sprintf("Unknown stmt node type: %T", stmt))
}

func (c *jsCompiler) convertExpr(expr ast.Expr) jsast.Expr {
	if expr == nil {
		return nil
	}

	switch n := expr.(type) {
	case *ast.BasicLit:
		return &jsast.BasicLiteral{Value: n.Value}
	case *ast.BinaryExpr:
		return &jsast.BinaryExpression{
			Lhs: c.convertExpr(n.X),
			Op:  c.convertOp(n.Op),
			Rhs: c.convertExpr(n.Y),
		}
	case *ast.Ident:
		return &jsast.Identifier{Name: c.getJsIdent(n)}
	case *ast.StructType:
		return &jsast.DeclExpr{Decl: &jsast.ClassDecl{
			Fields: c.convertFields(n.Fields.List),
		}}
	case *ast.SelectorExpr:
		return &jsast.SelectorExpr{
			X:   c.convertExpr(n.X),
			Sel: c.getJsIdent(n.Sel),
		}

	}

	panic(fmt.Sprintf("Unknown expr node type: %T", expr))
}

func (c *jsCompiler) convertFields(fields []*ast.Field) []*jsast.VarDecl {
	var vars []*jsast.VarDecl
	for _, f := range fields {
		for _, n := range f.Names {
			//TODO: isExported?
			vars = append(vars, &jsast.VarDecl{
				Name: c.getJsIdent(n),
			})
		}
	}
	return vars
}

func (c *jsCompiler) convertOp(tok token.Token) string {
	switch tok {
	case token.EQL:
		return "==="
	case token.NEQ:
		return "!=="
	case token.DEFINE:
		return "="
	default:
		return tok.String()
	}
}

func (c *jsCompiler) convertSpec(spec ast.Spec, typ token.Token) jsast.Decl {
	switch n := spec.(type) {
	case *ast.ValueSpec:
		var sub []jsast.Node
		for idx, i := range n.Names {
			varDecl := &jsast.VarDecl{
				Name: c.getJsIdent(i),
			}
			if typ == token.CONST {
				varDecl.Kind = "const"
			} else {
				varDecl.Kind = "let"
			}

			if len(n.Values) > idx {
				varDecl.Value = c.convertExpr(n.Values[idx])
			} else if tName, ok := n.Type.(*ast.Ident); ok && tName != nil {
				// if our type is a named struct then we
				// need to instantiate it as a class
				// so our prototype has all the methods and fields

				// TODO: detect struct types
				varDecl.Value = &jsast.ClassInstantiate{ClassName: tName.Name}
			}

			//TODO: exported
			sub = append(sub, varDecl)
		}
		if len(sub) == 1 {
			return sub[0].(jsast.Decl)
		}
		return &jsast.Placeholder{sub}

	case *ast.TypeSpec:
		//TODO: other type spec types
		//- type a int
		//- type b mod.Thing
		//- type c = d ....aliases?

		nm := c.getJsIdent(n.Name)
		typ := c.convertExpr(n.Type).(*jsast.DeclExpr)
		switch t := typ.Decl.(type) {
		case *jsast.ClassDecl:
			t.Name = nm
		default:
			panic(fmt.Sprintf("unsupported decl type: %T", t))
		}
		return typ.Decl
	}

	panic(fmt.Sprintf("Unknown spec node type: %T", spec))
}

func (c *jsCompiler) getJsIdent(i *ast.Ident) string {
	//TODO: handle escaping idents that aren't valid in JS
	// look up in our map
	return i.Name
}

/*

	case *ast.BasicLit:
		c.writeLiteral(node)

	case *ast.DeclStmt:
		c.symbols.defineSymbol(node.Name.Value)

		c.emit("let ")
		c.emit(c.jsSafeIdent(node.Name))

		if node.Value != nil {
			c.emit(" = ")
			if err := c.Compile(node.Value); err != nil {
				return err
			}
		} else if node.Type != nil {
			// if we have no set value, but do have a type
			// set it equal to 'new Type()'
			c.emit(" = new ")
			if err := c.Compile(node.Type); err != nil {
				return err
			}
			c.emit("()")
		}

		c.emit(";\n")

	case *ast.Identifier:
		_, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.emit(c.jsSafeIdent(node))

	case *ast.IfStatement:
		c.emit("if (")
		if err := c.Compile(node.Condition); err != nil {
			return err
		}
		c.emit(") ")
		if err := c.Compile(node.Consequence); err != nil {
			return err
		}
		if node.Alternative != nil {
			c.emit("else ")
			if err := c.Compile(node.Alternative); err != nil {
				return err
			}
		}

	case *ast.InfixExpression:
		if err := c.Compile(node.Left); err != nil {
			return err
		}

		switch node.Token.Type {
		case token.EQ:
			c.emit(" === ")
		case token.NOT_EQ:
			c.emit(" !== ")
		default:
			c.emit(" ")
			c.emit(node.Token.Literal)
			c.emit(" ")
		}
		if err := c.Compile(node.Right); err != nil {
			return err
		}

	case *ast.SelectorExpression:
		if err := c.Compile(node.Lhs); err != nil {
			return err
		}
		c.emit(".")
		c.emit(c.jsSafeIdent(node.Sel))

	case *ast.AssignStatement:
		if err := c.Compile(node.Lhs); err != nil {
			return err
		}
		c.emit(" ")
		c.emit(node.Operator)
		c.emit(" ")
		if err := c.Compile(node.Rhs); err != nil {
			return err
		}
		c.emit(";\n")

	case *ast.BlockStatement:
		c.emit("{\n")
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}
		c.emit("};\n")

	case *ast.TypeDeclStatement:
		if n, ok := node.TypeDef.(*ast.StructExpression); ok {
			// structs become classes with constructors
			c.compileStruct(node.Name, n)
		} else if n, ok := node.TypeDef.(*ast.EnumExpression); ok {
			c.compileEnum(node.Name, n)
		} else {
			return fmt.Errorf("unknown type decl: %+v", node.TypeDef)
		}
		c.symbolTable.Define(node.Name.Value)

	case *ast.ReturnStatement:
		c.emit("return ")
		if err := c.Compile(node.ReturnValue); err != nil {
			return err
		}
		c.emit(";\n")
	}

	return nil
}

func (p *JsPrinter) writeLiteral(n *ast.BasicLit) {
	switch n.Kind {
	case token.STRING:
		p.emit("\"")
		p.emit(n.Value)
		p.emit("\"")

	case token.INT, token.FLOAT:
		p.emit(n.Value)

	default:
		panic("unknown literal type: " + n.Kind.String())
	}

}

func (c *JsPrinter) writeStruct(name *ast.Ident, n *ast.StructType) {
	// structs become JS classes

	//	class Food {
	//		constructor (name, protein, carbs, fat) {
	//			this.name = name;
	//		};
	//	};

	c.emit("class ")
	c.emit(c.jsSafeIdent(name))

	c.emit(" {\nconstructor(")
	sep := ""
	for _, f := range n.Fields.Fields {
		c.emit(sep)
		c.emit(c.jsSafeIdent(f.Name))
		sep = ", "
	}
	c.emit(") {\n")
	//constructor body
	for _, f := range n.Fields.Fields {
		c.emitFmt("this.%s = %s;\n", c.jsSafeIdent(f.Name), c.jsSafeIdent(f.Name))
	}
	c.emit("};\n") //endconstructor

	c.emit("};\n") //endclass
}

func (c *JsPrinter) writeEnum(name *ast.Ident, n *ast.EnumType) {
	//enums become const
	//const Colors = Object.freeze({
	//  RED:   { name: "red", hex: "#f00" },
	//  BLUE:  { name: "blue", hex: "#00f" },
	//  GREEN: { name: "green", hex: "#0f0" }
	//});
	c.emit("const ")
	c.emit(c.jsSafeIdent(name))
	c.emit(" = Object.freeze({")
	sep := "\n"
	for _, f := range n.Fields.Fields {
		c.emit(sep)
		c.emitFmt("%s: { name: \"%s\", value: %s }", c.jsSafeIdent(f.Name), c.jsSafeIdent(f.Name), f.Value.String())
		sep = ",\n"
	}
	c.emit("\n});\n")
}

func (c *JsPrinter) jsSafeIdent(i *ast.Ident) string {
	//TODO: handle escaping idents that aren't valid in JS
	return i.Name
}

func (c *JsPrinter) emit(s string) {
	io.WriteString(c.output, s)
}

func (c *JsPrinter) emitFmt(format string, a ...interface{}) int {
	n, _ := fmt.Fprintf(c.output, format, a...)
	return n
}

*/
