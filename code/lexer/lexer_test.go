package lexer

import (
	"testing"

	"weblang/code/token"
)

func TestNextToken(t *testing.T) {
	input := `var t = "test"
	+(){},
const x = 123.45
if a == b {
	return true
} else {
	return false
}
var result = add(five, ten)
10 == 10
10 != 9
type testing struct { 
    a int
    b string = "default?"
    c struct { inner string }
}
d = e.f.g`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
		line, col       int
	}{
		{token.VAR, "var", 1, 1}, //0
		{token.IDENT, "t", 1, 5},
		{token.ASSIGN, "=", 1, 7},
		{token.STRING, "test", 1, 10},
		{token.SEMICOLON, "\n", 1, 15},
		{token.PLUS, "+", 2, 2},
		{token.LPAREN, "(", 2, 3},
		{token.RPAREN, ")", 2, 4},
		{token.LBRACE, "{", 2, 5},
		{token.RBRACE, "}", 2, 6},
		{token.COMMA, ",", 2, 7}, //10
		{token.CONST, "const", 3, 1},
		{token.IDENT, "x", 3, 7},
		{token.ASSIGN, "=", 3, 9},
		{token.FLOAT, "123.45", 3, 11},
		{token.SEMICOLON, "\n", 3, 17},
		{token.IF, "if", 4, 1},
		{token.IDENT, "a", 4, 4},
		{token.EQ, "==", 4, 6},
		{token.IDENT, "b", 4, 9},
		{token.LBRACE, "{", 4, 11}, //20
		{token.RETURN, "return", 5, 2},
		{token.TRUE, "true", 5, 9},
		{token.SEMICOLON, "\n", 5, 13},
		{token.RBRACE, "}", 6, 1},
		{token.ELSE, "else", 6, 3},
		{token.LBRACE, "{", 6, 8},
		{token.RETURN, "return", 7, 2},
		{token.FALSE, "false", 7, 9},
		{token.SEMICOLON, "\n", 7, 14},
		{token.RBRACE, "}", 8, 1}, //30
		{token.SEMICOLON, "\n", 8, 2},
		{token.VAR, "var", 9, 1},
		{token.IDENT, "result", 9, 5},
		{token.ASSIGN, "=", 9, 12},
		{token.IDENT, "add", 9, 14},
		{token.LPAREN, "(", 9, 17},
		{token.IDENT, "five", 9, 18},
		{token.COMMA, ",", 9, 22},
		{token.IDENT, "ten", 9, 24},
		{token.RPAREN, ")", 9, 27},
		{token.SEMICOLON, "\n", 9, 28}, //40
		{token.INT, "10", 10, 1},
		{token.EQ, "==", 10, 4},
		{token.INT, "10", 10, 7},
		{token.SEMICOLON, "\n", 10, 9},
		{token.INT, "10", 11, 1},
		{token.NOT_EQ, "!=", 11, 4},
		{token.INT, "9", 11, 7},
		{token.SEMICOLON, "\n", 11, 8},
		{token.TYPE, "type", 12, 1},
		{token.IDENT, "testing", 12, 6}, //50
		{token.STRUCT, "struct", 12, 14},
		{token.LBRACE, "{", 12, 21},
		{token.IDENT, "a", 13, 5},
		{token.IDENT, "int", 13, 7},
		{token.SEMICOLON, "\n", 13, 10},
		{token.IDENT, "b", 14, 5},
		{token.IDENT, "string", 14, 7},
		{token.ASSIGN, "=", 14, 14},
		{token.STRING, "default?", 14, 17},
		{token.SEMICOLON, "\n", 14, 26}, //60
		{token.IDENT, "c", 15, 5},
		{token.STRUCT, "struct", 15, 7},
		{token.LBRACE, "{", 15, 14},
		{token.IDENT, "inner", 15, 16},
		{token.IDENT, "string", 15, 22},
		{token.RBRACE, "}", 15, 29},
		{token.SEMICOLON, "\n", 15, 30},
		{token.RBRACE, "}", 16, 1},
		{token.SEMICOLON, "\n", 16, 2},
		{token.IDENT, "d", 17, 1}, //70
		{token.ASSIGN, "=", 17, 3},
		{token.IDENT, "e", 17, 5},
		{token.DOT, ".", 17, 6},
		{token.IDENT, "f", 17, 7},
		{token.DOT, ".", 17, 8},
		{token.IDENT, "g", 17, 9},
		{token.SEMICOLON, "", 17, 10}, //77
	}

	l := New(input, "testFile")

	for i, tt := range tests {
		tok := l.NextToken()

		if want, got := tt.expectedType, tok.Type; want != got {
			t.Fatalf("test[%d: %v] - Type expected %v, got %v", i, tests[i], want, got)
		}
		if want, got := tt.expectedLiteral, tok.Literal; want != got {
			t.Fatalf("test[%d: %v] - Literal expected %v, got %v", i, tests[i], want, got)
		}
		if want, got := tt.line, tok.LineNum; want != got {
			t.Fatalf("test[%d: %v] - Line expected %v, got %v", i, tests[i], want, got)
		}
		if want, got := tt.col, tok.ColNum; want != got {
			t.Fatalf("test[%d: %+v] - Col expected %v, got %v", i, tests[i], want, got)
		}
		if want, got := "testFile", tok.FileName; want != got {
			t.Fatalf("test[%d: %v] - File expected %s, got %s", i, tests[i], want, got)
		}
	}
}
