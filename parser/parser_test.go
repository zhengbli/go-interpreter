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
		value    interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!true;", "!", true},
		{"!false;", "!", false},
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

		if !testLiteralExpression(t, prefix.Right, test.value) {
			return
		}
	}
}

func TestInfixExpression(t *testing.T) {
	tests := []struct {
		input    string
		left     interface{}
		operator string
		right    interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
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

		if !testLiteralExpression(t, exp.Right, test.right) {
			return
		}
	}

}

func TestBooleanLiteral(t *testing.T) {
	tests := []struct {
		input string
		value bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, test := range tests {
		program := parse(test.input, 1, t)
		st, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("Expected expression statement, got %T", program.Statements[0])
		}

		exp, ok := st.Expression.(*ast.BooleanLiteral)
		if !ok {
			t.Fatalf("Expected boolean literal, got %T", st.Expression)
		}

		if exp.Value != test.value {
			t.Fatalf("Expected value %v, got value %v", test.value, exp.Value)
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

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`
	program := parse(input, 1, t)
	st, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Not Expression statement")
	}

	ifs, ok := st.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("Not If Expression")
	}

	condExp := ifs.Condition.(*ast.InfixExpression)
	if condExp.Operator != "<" {
		t.Fatalf("If condition operator is incorrect")
	}

	ifBody := ifs.Body.Statements[0].TokenLiteral()
	if ifBody != "x" {
		t.Fatalf("If body is incorrect")
	}

	elseBody := ifs.ElseBody.Statements[0].TokenLiteral()
	if elseBody != "y" {
		t.Fatalf("Else body is incorrect")
	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T",
			stmt.Expression)
	}

	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. got=%d", len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

func TestFuncLiteral(t *testing.T) {
	input := "fn(x, y) { return x + y; }"
	program := parse(input, 1, t)
	st := program.Statements[0].(*ast.ExpressionStatement)
	fn := st.Expression.(*ast.FunctionLiteral)
	if len(fn.FunctionParameters) != 2 {
		t.Fatalf("Wrong number of function parameters")
	}

	returnSt, ok := fn.FunctionBody.Statements[0].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("Expect return statement, got %T", returnSt)
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
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

	t.Errorf("Not handled, got type: %T", exp)
	return false
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

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("Not InfixExpression")
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
