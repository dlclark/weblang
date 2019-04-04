package parser

import (
	"fmt"
	"strconv"
	"weblang/code/ast"
	"weblang/code/lexer"
	"weblang/code/token"
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	//expression parse functions
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
	SELECTOR    // a.b
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
	token.DOT:      SELECTOR,
}

type prefixParseFn func() ast.Expression

type infixParseFn func(ast.Expression) ast.Expression

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:              l,
		errors:         []string{},
		prefixParseFns: make(map[token.TokenType]prefixParseFn),
		infixParseFns:  make(map[token.TokenType]infixParseFn),
	}

	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)

	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.DOT, p.parseSelectorExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	if p.curToken.Type == token.EOF {
		//lets stop at the end of the file
		return
	}
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expect(t token.TokenType) bool {
	if p.curToken.Type != t {
		p.errExpected("'" + string(t) + "'")
		return false
	}

	return true
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	p.err(p.peekToken, fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type))
}

func (p *Parser) noPrefixParseFnError(t token.Token) {
	p.err(t, fmt.Sprintf("no prefix parse function for %s found", t.Type))
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.expect(token.SEMICOLON)
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.VAR:
		return p.parseVarStatement()
	case token.TYPE:
		return p.parseTypeDeclStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.LBRACE:
		return p.parseBlockStatement()
	case token.IDENT, token.INT, token.FLOAT, token.STRING, token.RAWSTRING, token.FUNCTION, token.LPAREN, // operands
		token.LBRACKET, token.STRUCT, //composite types
		token.PLUS, token.MINUS, token.ASTERISK, token.BANG: //unary operators
		return p.parseSimpleStatement()
	case token.SEMICOLON:
		return nil
	default:
		//error
		p.errExpected("statement")
		return nil
	}
}

func (p *Parser) parseTypeDeclStatement() *ast.TypeDeclStatement {
	p.expect(token.TYPE)
	s := &ast.TypeDeclStatement{Token: p.curToken}

	// capture the name
	p.nextToken()
	p.expect(token.IDENT)
	s.Name = p.parseIdentifier().(*ast.Identifier)

	p.nextToken()
	//could be enum or struct
	switch p.curToken.Type {
	case token.STRUCT:
		s.TypeDef = p.parseStructExpression()
	case token.ENUM:
		s.TypeDef = p.parseEnumExpression()
	default:
		p.errExpected("struct or enum")
		return nil
	}

	return s
}

func (p *Parser) parseStructExpression() *ast.StructExpression {
	p.expect(token.STRUCT)
	e := &ast.StructExpression{Token: p.curToken, Fields: &ast.FieldList{}}

	// should be on the LBRACE
	p.nextToken()
	p.expect(token.LBRACE)
	e.Fields.Opening = p.curToken

	// move past LBRACE
	p.nextToken()
	// now parse the fields in a loop
	for p.curTokenIs(token.IDENT) {
		e.Fields.Fields = append(e.Fields.Fields, p.parseFieldDecl())
	}

	p.expect(token.RBRACE)
	e.Fields.Closing = p.curToken

	p.nextToken()

	return e
}

func (p *Parser) parseEnumExpression() *ast.EnumExpression {
	p.expect(token.ENUM)
	e := &ast.EnumExpression{Token: p.curToken, Fields: &ast.FieldList{}}

	// should be on the LBRACE
	p.nextToken()
	p.expect(token.LBRACE)
	e.Fields.Opening = p.curToken

	p.nextToken()
	// now parse the fields in a loop
	for p.curTokenIs(token.IDENT) {
		e.Fields.Fields = append(e.Fields.Fields, p.parseFieldDecl())
	}

	p.expect(token.RBRACE)
	e.Fields.Closing = p.curToken

	p.nextToken()

	return e
}

func (p *Parser) parseFieldDecl() *ast.Field {
	// name type [= exp]
	p.expect(token.IDENT)
	f := &ast.Field{}
	f.Name = p.parseIdentifier().(*ast.Identifier)

	p.nextToken()

	//could be enum or struct
	switch p.curToken.Type {
	case token.SEMICOLON:
		// type matches the name, done
		f.Type = f.Name
	case token.ASSIGN:
		// like a const/enum: someVal = 0
		p.nextToken()
		f.Value = p.parseExpression(LOWEST)
		//TODO: validate expression is only literals

	case token.IDENT:
		// named type ident
		f.Type = p.parseIdentifier()
		p.nextToken()
		if p.curTokenIs(token.ASSIGN) {
			p.nextToken()
			f.Value = p.parseExpression(LOWEST)
			//TODO: validate expression is only literals
		}
	case token.STRUCT:
		// struct literal
		f.Type = p.parseStructExpression()
	case token.ENUM:
		// enum literal
		f.Type = p.parseEnumExpression()
	default:
		p.errExpected("end of statement, an identifier, a struct, or an enum")
		return nil
	}

	// eat the semicolon at the end, can't require it
	// because the final item in a block does not require it
	if p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return f
}

func (p *Parser) parseVarStatement() *ast.VarStatement {
	stmt := &ast.VarStatement{Token: p.curToken}

	p.nextToken()
	if !p.expect(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken()

	// either be an type OR a assignment
	if p.curTokenIs(token.ASSIGN) {
		p.nextToken()
		stmt.Value = p.parseExpression(LOWEST)
	} else {
		stmt.Type = p.parseExpression(LOWEST)

		if p.curTokenIs(token.ASSIGN) {
			stmt.Value = p.parseExpression(LOWEST)
		}
	}

	if fl, ok := stmt.Value.(*ast.FunctionLiteral); ok {
		fl.Name = stmt.Name.Value
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	return &ast.ReturnStatement{
		Token:       p.curToken,
		ReturnValue: exp,
	}
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken)
		return nil
	}
	leftExp := prefix()

	p.nextToken()

	for !p.curTokenIs(token.SEMICOLON) && precedence < p.curPrecedence() {
		infix := p.infixParseFns[p.curToken.Type]
		if infix == nil {
			return leftExp
		}

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseSelectorExpression(left ast.Expression) ast.Expression {
	exp := &ast.SelectorExpression{
		Token: p.curToken,
		Lhs:   left,
	}

	p.expectPeek(token.IDENT)
	exp.Sel = p.parseIdentifier().(*ast.Identifier)
	p.nextToken()

	return exp
}

func (p *Parser) parseSimpleStatement() ast.Statement {
	// parse LHS - as an expression
	left := p.parseExpression(LOWEST)

	// if our next token is
	switch p.curToken.Type {
	case token.ASSIGN:
		op := p.curToken
		//assignment statement
		p.nextToken()
		right := p.parseExpression(LOWEST)

		return &ast.AssignStatement{
			Token:    op,
			Lhs:      left,
			Operator: op.Literal,
			Rhs:      right,
		}
		//case token.INC, token.DEC:
	}

	return &ast.ExpressionStatement{
		Token:      p.curToken,
		Expression: left,
	}
}

func (p *Parser) parseIfStatement() ast.Statement {
	stmt := &ast.IfStatement{Token: p.curToken}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expect(token.LBRACE) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	if p.curToken.Type == token.ELSE {
		// eat the ELSE
		p.nextToken()

		switch p.curToken.Type {
		case token.IF:
			stmt.Alternative = p.parseIfStatement()
		case token.LBRACE:
			stmt.Alternative = p.parseBlockStatement()
		default:
			p.errExpected("if statement or block")
		}
	}

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		} else {
			p.nextToken()
		}

		//parse the semicolon if we're not at the rbrace already
		if !p.curTokenIs(token.RBRACE) {
			p.expect(token.SEMICOLON)
			p.nextToken()
		}
	}

	//eat our R-brace
	p.expect(token.RBRACE)
	p.nextToken()

	return block
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 53)
	if err != nil {
		p.err(p.curToken, fmt.Sprintf("could not parse %q as integer", p.curToken.Literal))
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		p.err(p.curToken, fmt.Sprintf("could not parse %q as float", p.curToken.Literal))
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) err(tok token.Token, msg string) {
	p.errors = append(p.errors, fmt.Sprintf("%s:%d:%d: %s", tok.FileName, tok.LineNum, tok.ColNum, msg))
	if len(p.errors) > 10 {
		panic("too many errors")
	}
}

func (p *Parser) errExpected(msg string) {
	p.err(p.curToken, fmt.Sprintf("expected %s, found %s", msg, p.curToken.Literal))
}
