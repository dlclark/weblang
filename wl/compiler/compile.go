package compiler

import (
	"fmt"
	"strconv"
	"strings"
	"weblang/code/ast"
	"weblang/code/object"
	"weblang/code/token"
)

type Compiler struct {
	constants []object.Object

	symbolTable *SymbolTable

	output *strings.Builder
}

func New() *Compiler {

	symbolTable := NewSymbolTable()

	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		output:      &strings.Builder{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			if err := c.Compile(s); err != nil {
				return err
			}
		}

	case *ast.IntegerLiteral:
		c.emit(strconv.FormatInt(node.Value, 10))

	case *ast.StringLiteral:
		c.emit("\"")
		c.emit(node.Value)
		c.emit("\"")

	case *ast.BooleanLiteral:
		if node.Value {
			c.emit("true")
		} else {
			c.emit("false")
		}

	case *ast.VarStatement:
		c.symbolTable.Define(node.Name.Value)

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

func (c *Compiler) compileStruct(name *ast.Identifier, n *ast.StructExpression) {
	// structs become JS classes
	/*
		class Food {
			constructor (name, protein, carbs, fat) {
				this.name = name;
			};
		};
	*/
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

func (c *Compiler) compileEnum(name *ast.Identifier, n *ast.EnumExpression) {
	//enums become const
	/*const Colors = Object.freeze({
	  RED:   { name: "red", hex: "#f00" },
	  BLUE:  { name: "blue", hex: "#00f" },
	  GREEN: { name: "green", hex: "#0f0" }
	});*/
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

func (c *Compiler) jsSafeIdent(i *ast.Identifier) string {
	//TODO: handle escaping idents that aren't valid in JS
	return i.Value
}

func (c *Compiler) Output() string {
	return strings.TrimSpace(c.output.String())
}

func (c *Compiler) emit(s string) {
	c.output.WriteString(s)
}

func (c *Compiler) emitFmt(format string, a ...interface{}) int {
	n, _ := fmt.Fprintf(c.output, format, a...)
	return n
}
