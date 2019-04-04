package parser

import (
	"fmt"
	"testing"
	"weblang/code/ast"
	"weblang/code/lexer"
	"weblang/code/token"
)

func TestVarStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"var x = 5", "x", 5},
		{"var y = true", "y", true},
		{"var foobar = y", "foobar", "y"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input, "junk")
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testVarStatement(t, stmt, tt.expectedIdentifier) {
			return
		}

		val := stmt.(*ast.VarStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if x < y { a = b }`

	l := lexer.New(input, "test")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IfStatement. got=%T", program.Statements[0])
	}

	if !testInfixExpression(t, stmt.Condition, "x", "<", "y") {
		return
	}

	if len(stmt.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got=%d\n", len(stmt.Consequence.Statements))
	}

	consequence, ok := stmt.Consequence.Statements[0].(*ast.AssignStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.AssignStatement. got=%T", stmt.Consequence.Statements[0])
	}

	if !testAssignStatement(t, consequence, "a", "=", "b") {
		return
	}

	if stmt.Alternative != nil {
		t.Errorf("exp.Alternative.Statements was not nil. got=%+v", stmt.Alternative)
	}
}

func TestIfElseStatement(t *testing.T) {
	input := `if x < y { a = b } else { c = d }`

	l := lexer.New(input, "test")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	stmt := program.Statements[0].(*ast.IfStatement)
	if !testInfixExpression(t, stmt.Condition, "x", "<", "y") {
		return
	}

	consequence := stmt.Consequence.Statements[0].(*ast.AssignStatement)
	if !testAssignStatement(t, consequence, "a", "=", "b") {
		return
	}

	alternative := stmt.Alternative.(*ast.BlockStatement).Statements[0].(*ast.AssignStatement)
	if !testAssignStatement(t, alternative, "c", "=", "d") {
		return
	}
}

func TestIfElseIfStatement(t *testing.T) {
	input := `if x < y { a = b } else if x == y { c = d }`

	l := lexer.New(input, "test")
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 1, len(program.Statements))
	}

	stmt := program.Statements[0].(*ast.IfStatement)
	if !testInfixExpression(t, stmt.Condition, "x", "<", "y") {
		return
	}

	consequence := stmt.Consequence.Statements[0].(*ast.AssignStatement)
	if !testAssignStatement(t, consequence, "a", "=", "b") {
		return
	}

	alternative := stmt.Alternative.(*ast.IfStatement)
	if !testInfixExpression(t, alternative.Condition, "x", "==", "y") {
		return
	}
	if !testAssignStatement(t, alternative.Consequence.Statements[0].(*ast.AssignStatement), "c", "=", "d") {
		return
	}
}

func TestStructDecl(t *testing.T) {
	input := `
	type testing struct { 
		a int
		b string = "default?"
		c struct { inner string }
	}`

	l := lexer.New(input, "test")
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if want, got := 1, len(program.Statements); want != got {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", want, got)
	}

	td := program.Statements[0].(*ast.TypeDeclStatement)
	if want, got := "testing", td.Name.Value; want != got {
		t.Fatalf("name mismatch, want=%v got=%v", want, got)
	}

	if want, got := token.TYPE, string(td.Token.Type); want != got {
		t.Fatalf("td token mismatch, want=%v got=%v", want, got)
	}

	st := td.TypeDef.(*ast.StructExpression)
	if want, got := token.STRUCT, string(st.Token.Type); want != got {
		t.Fatalf("st token mismatch, want=%v got=%v", want, got)
	}
	if want, got := 3, len(st.Fields.Fields); want != got {
		t.Fatalf("field count mismatch, want=%v got=%v", want, got)
	}
	testField(t, st.Fields.Fields[0], "a", "int", "")
	testField(t, st.Fields.Fields[1], "b", "string", "\"default?\"")
	testField(t, st.Fields.Fields[2], "c", "struct {inner string}", "")
}

func TestEnumDecl(t *testing.T) {
	input := `
type test enum {
    None = 0
    Blah = 1
    Yu = 2
}`
	l := lexer.New(input, "test")
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if want, got := 1, len(program.Statements); want != got {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", want, got)
	}

	td := program.Statements[0].(*ast.TypeDeclStatement)
	if want, got := "test", td.Name.Value; want != got {
		t.Fatalf("name mismatch, want=%v got=%v", want, got)
	}
	if want, got := token.TYPE, string(td.Token.Type); want != got {
		t.Fatalf("td token mismatch, want=%v got=%v", want, got)
	}

	st := td.TypeDef.(*ast.EnumExpression)
	if want, got := token.ENUM, string(st.Token.Type); want != got {
		t.Fatalf("st token mismatch, want=%v got=%v", want, got)
	}
	if want, got := 3, len(st.Fields.Fields); want != got {
		t.Fatalf("field count mismatch, want=%v got=%v", want, got)
	}
	testField(t, st.Fields.Fields[0], "None", "", "0")
	testField(t, st.Fields.Fields[1], "Blah", "", "1")
	testField(t, st.Fields.Fields[2], "Yu", "", "2")
}

func TestAssignmentStatement(t *testing.T) {
	input := `
	a.b = c
	d = e.f
	g = h.i.j`

	l := lexer.New(input, "test")
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if want, got := 3, len(program.Statements); want != got {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", want, got)
	}

	asn := program.Statements[0].(*ast.AssignStatement)
	if want, got := "a", asn.Lhs.(*ast.SelectorExpression).Lhs.(*ast.Identifier).Value; want != got {
		t.Fatalf("line 1 lhs.lhs  want=%v got=%v", want, got)
	}
	if want, got := "b", asn.Lhs.(*ast.SelectorExpression).Sel.Value; want != got {
		t.Fatalf("line 1 lhs.sel  want=%v got=%v", want, got)
	}
	if want, got := "c", asn.Rhs.(*ast.Identifier).String(); want != got {
		t.Fatalf("line 1 rhs. want=%v got=%v", want, got)
	}

	asn = program.Statements[1].(*ast.AssignStatement)
	if want, got := "d", asn.Lhs.(*ast.Identifier).String(); want != got {
		t.Fatalf("line 2 lhs. want=%v got=%v", want, got)
	}
	if want, got := "e", asn.Rhs.(*ast.SelectorExpression).Lhs.(*ast.Identifier).Value; want != got {
		t.Fatalf("line 2 rhs.lhs  want=%v got=%v", want, got)
	}
	if want, got := "f", asn.Rhs.(*ast.SelectorExpression).Sel.Value; want != got {
		t.Fatalf("line 2 rhs.sel  want=%v got=%v", want, got)
	}

	asn = program.Statements[2].(*ast.AssignStatement)
	if want, got := "g", asn.Lhs.(*ast.Identifier).String(); want != got {
		t.Fatalf("line 3 lhs. want=%v got=%v", want, got)
	}
	if want, got := "h", asn.Rhs.(*ast.SelectorExpression).Lhs.(*ast.SelectorExpression).Lhs.(*ast.Identifier).Value; want != got {
		t.Fatalf("line 3 rhs.left.left want=%v got=%v", want, got)
	}
	if want, got := "i", asn.Rhs.(*ast.SelectorExpression).Lhs.(*ast.SelectorExpression).Sel.Value; want != got {
		t.Fatalf("line 3 rhs.left.sel want=%v got=%v", want, got)
	}
	if want, got := "j", asn.Rhs.(*ast.SelectorExpression).Sel.Value; want != got {
		t.Fatalf("line 3 rhs.sel want=%v got=%v", want, got)
	}

}

func TestReturnStatement(t *testing.T) {
	input := `return true`

	l := lexer.New(input, "test")
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if want, got := 1, len(program.Statements); want != got {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n", want, got)
	}

	td := program.Statements[0].(*ast.ReturnStatement)
	if want, got := true, td.ReturnValue.(*ast.BooleanLiteral).Value; want != got {
		t.Fatalf("return val mismatch, want=%v got=%v", want, got)
	}
}

func testField(t *testing.T, f *ast.Field, name string, typ string, val string) {
	if want, got := name, f.Name.Value; want != got {
		t.Fatalf("field name mismatch, want=%v got=%v", want, got)
	}

	if typ != "" || f.Type != nil {
		if want, got := typ, f.Type.String(); want != got {
			t.Fatalf("field type mismatch, want=%v got=%v", want, got)
		}
	}

	if val != "" || f.Value != nil {
		if want, got := val, f.Value.String(); want != got {
			t.Fatalf("field value mismatch, want=%v got=%v", want, got)
		}
	}
}

func testAssignStatement(t *testing.T, stmt ast.Statement, left interface{}, operator string, right interface{}) bool {

	opStmt, ok := stmt.(*ast.AssignStatement)
	if !ok {
		t.Errorf("exp is not ast.AssignStatement. got=%T(%s)", stmt, stmt)
		return false
	}

	if !testLiteralExpression(t, opStmt.Lhs, left) {
		return false
	}

	if opStmt.Operator != operator {
		t.Errorf("opStmt.Operator is not '%s'. got=%q", operator, opStmt.Operator)
		return false
	}

	if !testLiteralExpression(t, opStmt.Rhs, right) {
		return false
	}

	return true
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) bool {

	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.InfixExpression. got=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func testVarStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "var" {
		t.Errorf("s.TokenLiteral not 'var'. got=%q", s.TokenLiteral())
		return false
	}

	varStmt, ok := s.(*ast.VarStatement)
	if !ok {
		t.Errorf("s not *ast.VarStatement. got=%T", s)
		return false
	}

	if varStmt.Name.Value != name {
		t.Errorf("varStmt.Name.Value not '%s'. got=%s", name, varStmt.Name.Value)
		return false
	}

	if varStmt.Name.TokenLiteral() != name {
		t.Errorf("varStmt.Name.TokenLiteral() not '%s'. got=%s",
			name, varStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)

	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value,
			integ.TokenLiteral())
		return false
	}

	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value,
			ident.TokenLiteral())
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.BooleanLiteral)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}

	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}

	if bo.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("bo.TokenLiteral not %t. got=%s",
			value, bo.TokenLiteral())
		return false
	}

	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
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
