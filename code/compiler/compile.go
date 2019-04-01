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
		c.emit(node.Name.Value)

		if node.Value != nil {
			c.emit(" = ")
			if err := c.Compile(node.Value); err != nil {
				return err
			}
		}
		c.emit(";\n")

	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		//TODO: map to JS name?
		c.emit(symbol.Name)

	case *ast.IfStatement:
		c.emit("if (")
		if err := c.Compile(node.Condition); err != nil {
			return err
		}
		c.emit(") ")
		c.Compile(node.Consequence)
		if node.Alternative != nil {
			c.emit("else ")
			c.Compile(node.Alternative)
		}

	case *ast.InfixExpression:
		c.Compile(node.Left)
		c.emit(" ")
		switch node.Token.Type {
		case token.EQ:
			c.emit("===")
		case token.NOT_EQ:
			c.emit("!==")
		default:
			c.emit(node.Token.Literal)
		}
		c.emit(" ")
		c.Compile(node.Right)

	case *ast.AssignStatement:
		c.Compile(node.Lhs)
		c.emit(" ")
		c.emit(node.Operator)
		c.emit(" ")
		c.Compile(node.Rhs)
		c.emit(";\n")

	case *ast.BlockStatement:
		c.emit("{\n")
		for _, s := range node.Statements {
			c.Compile(s)
		}

		c.emit("};\n")
	}

	return nil
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
