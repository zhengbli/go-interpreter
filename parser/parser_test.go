package parser

import (
	"inter/ast"
	"inter/lexer"
	"inter/token"
	"testing"
)

func TestLetStatements(t *testing.T) {
	input := `
let x = 5;
let y= 10;
let foobar = 838383;
`
	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	if program == nil {
		t.Fatalf("ParseProgram() returns nil")
	}

	if len(p.errors) > 0 {
		t.Errorf("Parser has %d errors", len(p.errors))
		for _, err := range p.errors {
			t.Errorf("Parser error: %q", err)
		}

		t.FailNow()
	}

	if len(program.Statements) != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d",
			len(program.Statements))
	}

	tests := []struct{ expectedIdentifier string }{
		{"x"},
		{"y"},
		{"foobar"},
	}
	for i, st := range program.Statements {
		if !testLetStatements(t, st, tests[i].expectedIdentifier) {
			return
		}
	}
}

func testLetStatements(t *testing.T, st ast.Statement, iden string) bool {
	if st.TokenLiteral() != "let" {
		t.Errorf("Statement token expected to be 'let', got=%q", st.TokenLiteral())
		return false
	}

	letSt, ok := st.(*ast.LetStatement)
	if !ok {
		t.Errorf("Statement is not let statement")
		return false
	}

	if letSt.Name.Value != iden {
		t.Errorf("Let statement name expected=%q, got=%q", iden, letSt.Name.Value)
		return false
	}

	if letSt.Name.TokenLiteral() != iden {
		t.Errorf("Let statement name token literal expected=%q, got=%q", iden, letSt.Name.TokenLiteral())
		return false
	}

	return true
}

func TestReturnStatements(t *testing.T) {
	input := `
return 123;
return 5;
return 999;
`
	l := lexer.New(input)
	parser := New(l)
	program := parser.ParseProgram()
	if program == nil {
		t.Fatalf("Could not parse program")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("Expected 3 statements, got %d", len(program.Statements))
	}

	for _, st := range program.Statements {
		rSt, ok := st.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("Statement is not *ast.ReturnStatement. got %T", st)
			continue
		}

		if rSt.Token.Type != token.RETURN {
			t.Errorf("Statement token is expected to be RETURN, got %s", rSt.Token.Type)
		}
	}
}
