package lexer

import (
	"testing"

	"weblang/code/token"
)

func TestNextToken(t *testing.T) {
	input := `var t = "test"
	+(){},`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
		line, col       int
	}{
		{token.VAR, "var", 1, 1},
		{token.IDENT, "t", 1, 5},
		{token.ASSIGN, "=", 1, 7},
		{token.STRING, "test", 1, 10},
		{token.SEMICOLON, "\n", 1, 16},
		{token.PLUS, "+", 2, 2},
		{token.LPAREN, "(", 2, 3},
		{token.RPAREN, ")", 2, 4},
		{token.LBRACE, "{", 2, 5},
		{token.RBRACE, "}", 2, 6},
		{token.COMMA, ",", 2, 7},
	}

	l := New(input, "testFile")

	for i, tt := range tests {
		tok := l.NextToken()

		if want, got := tt.expectedType, tok.Type; want != got {
			t.Fatalf("test[%d] - Type expected %v, got %v", i, want, got)
		}
		if want, got := tt.expectedLiteral, tok.Literal; want != got {
			t.Fatalf("test[%d] - Literal expected %v, got %v", i, want, got)
		}
		if want, got := tt.line, tok.LineNum; want != got {
			t.Fatalf("test[%d] - Line expected %v, got %v", i, want, got)
		}
		if want, got := tt.col, tok.ColNum; want != got {
			t.Fatalf("test[%d] - Col expected %v, got %v", i, want, got)
		}
		if want, got := "testFile", tok.FileName; want != got {
			t.Fatalf("test[%d] - File expected %s, got %s", i, want, got)
		}
	}
}
