package ast

import (
	"testing"
	"weblang/code/token"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&VarStatement{
				Token: token.Token{Type: token.VAR, Literal: "var"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &StringLiteral{
					Token: token.Token{Type: token.STRING, Literal: "test"},
					Value: "test",
				},
			},
		},
	}

	if program.String() != "var myVar = \"test\";" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}
