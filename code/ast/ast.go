package ast

import (
	"bytes"
	"fmt"
	"strings"
	"weblang/code/token"
)

// The base Node interface
type Node interface {
	TokenLiteral() string
	String() string
}

// All statement nodes implement this
type Statement interface {
	Node
	statementNode()
}

// All expression nodes implement this
type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

///////////////////////////////
// Statements

type TypeDeclStatement struct {
	Token   token.Token // the type token
	Name    *Identifier
	TypeDef Expression // StructExpression or EnumExpression
}

func (s *TypeDeclStatement) statementNode()       {}
func (s *TypeDeclStatement) TokenLiteral() string { return s.Token.Literal }
func (s *TypeDeclStatement) String() string {
	return "type " + s.Name.String() + " " + s.TypeDef.String()
}

type VarStatement struct {
	Token token.Token // the token.VAR token
	Name  *Identifier
	Type  Expression
	Value Expression
}

func (ls *VarStatement) statementNode()       {}
func (ls *VarStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *VarStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	if ls.Type != nil {
		out.WriteString(" ")
		out.WriteString(ls.Type.String())
	}

	if ls.Value != nil {
		out.WriteString(" = ")
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

type ReturnStatement struct {
	Token       token.Token // the 'return' token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type IfStatement struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative Statement
}

func (s *IfStatement) statementNode()       {}
func (s *IfStatement) TokenLiteral() string { return s.Token.Literal }
func (s *IfStatement) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(s.Condition.String())
	out.WriteString(" ")
	out.WriteString(s.Consequence.String())

	if s.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(s.Alternative.String())
	}

	return out.String()
}

type AssignStatement struct {
	Token    token.Token // The assignment token
	Lhs      Expression
	Operator string
	Rhs      Expression
}

func (s *AssignStatement) statementNode()       {}
func (s *AssignStatement) TokenLiteral() string { return s.Token.Literal }
func (s *AssignStatement) String() string {
	return s.Lhs.String() + s.Operator + s.Rhs.String()
}

///////////////////////////////
// Expressions

type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (oe *InfixExpression) expressionNode()      {}
func (oe *InfixExpression) TokenLiteral() string { return oe.Token.Literal }
func (oe *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(oe.Left.String())
	out.WriteString(" ")
	out.WriteString(oe.Operator)
	out.WriteString(" ")
	out.WriteString(oe.Right.String())
	out.WriteString(")")

	return out.String()
}

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) String() string       { return b.Token.Literal }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (il *FloatLiteral) expressionNode()      {}
func (il *FloatLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *FloatLiteral) String() string       { return il.Token.Literal }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Token.Literal + "\"" }

type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Parameters []*Identifier
	Body       *BlockStatement
	Name       string
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	if fl.Name != "" {
		out.WriteString(fmt.Sprintf("<%s>", fl.Name))
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

type SelectorExpression struct {
	Lhs   Expression  //expression
	Token token.Token // the '.' token
	Sel   *Identifier // field selector
}

func (e *SelectorExpression) expressionNode()      {}
func (e *SelectorExpression) TokenLiteral() string { return e.Token.Literal }
func (e *SelectorExpression) String() string {
	return e.Lhs.String() + "." + e.Sel.String()
}

type StructExpression struct {
	Token  token.Token // the 'struct' token
	Fields *FieldList
}

func (e *StructExpression) expressionNode()      {}
func (e *StructExpression) TokenLiteral() string { return e.Token.Literal }
func (e *StructExpression) String() string {
	return "struct " + e.Fields.String()
}

type EnumExpression struct {
	Token  token.Token // the 'enum' token
	Fields *FieldList
}

func (e *EnumExpression) expressionNode()      {}
func (e *EnumExpression) TokenLiteral() string { return e.Token.Literal }
func (e *EnumExpression) String() string {
	return "enum " + e.Fields.String()
}

type FieldList struct {
	Opening token.Token // open brace
	Fields  []*Field
	Closing token.Token
}

func (fl *FieldList) String() string {
	var out bytes.Buffer

	fields := []string{}
	for _, f := range fl.Fields {
		fields = append(fields, f.String())
	}

	out.WriteString(fl.Opening.Literal)
	out.WriteString(strings.Join(fields, ", "))
	out.WriteString(fl.Closing.Literal)

	return out.String()
}

type Field struct {
	Name  *Identifier // fieldname
	Type  Expression  // its type, or nil if needs to be guessed
	Value Expression  // if it's set to something, otherwise nil
}

func (f *Field) String() string {
	out := f.Name.String() + " " + f.Type.String()
	if f.Value != nil {
		out += " = " + f.Value.String()
	}
	return out
}
