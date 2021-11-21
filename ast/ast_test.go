package ast

import (
	"inter/token"
	"testing"
)

func TestString(t *testing.T) {
	input := "let myVar = anotherVar;"

	Program := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
			},
		},
	}

	if got := Program.String(); got != input {
		t.Errorf("Expect: %s, got: %s", input, got)
	}
}
