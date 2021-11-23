package parser

import (
	"fmt"
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

func TestIdentifierExpression(t *testing.T) {
	program := parse("foobar;", 1, t)

	st, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
	}

	id, ok := st.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("Expected Identifier, got %T", st.Expression)
	}
	if id.Value != "foobar" {
		t.Fatalf("Expected \"foobar\", got %v", id.Value)
	}
}

func TestIntegerLiteral(t *testing.T) {
	program := parse("5;", 1, t)

	st, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
	}

	id, ok := st.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("Expected IntegerLiteral, got %T", st.Expression)
	}
	if id.Value != 5 {
		t.Fatalf("Expected 5, got %d", id.Value)
	}
}

func TestPrefixExpression(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		value    int64
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
	}

	for _, test := range tests {
		program := parse(test.input, 1, t)
		st, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected ExpressionStatement, got %T", program.Statements[0])
		}

		prefix, ok := st.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("Expected PrefixExpression, got %T", st.Expression)
		}

		if prefix.Operator != test.operator {
			t.Fatalf("Expected Operator %s, got %s", test.operator, prefix.Operator)
		}

		if !testIntegerLiteral(t, prefix.Right, test.value) {
			return
		}
	}
}

func TestInfixExpression(t *testing.T) {
	tests := []struct {
		input    string
		left     int
		operator string
		right    int
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
	}

	for _, test := range tests {
		program := parse(test.input, 1, t)
		st, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Not ExpressionStatement, but %T", program.Statements[0])
		}

		exp, ok := st.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("Not InfixExpression, but %T", st.Expression)
		}

		if exp.Operator != test.operator {
			t.Fatalf("Operator expected to be %s, got %s", test.operator, exp.Operator)
		}
	}

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

func parse(input string, expectedSt int, t *testing.T) *ast.Program {
	l := lexer.New(input)
	parser := New(l)
	program := parser.ParseProgram()
	checkParserErrors(t, parser)

	if program == nil {
		t.Fatalf("Failed to parse")
	}

	if len(program.Statements) != expectedSt {
		t.Fatalf("Expected %d statements, got %d", expectedSt, len(program.Statements))
	}

	return program
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.errors
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
