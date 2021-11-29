package parser

import (
	"fmt"
	"inter/ast"
	"inter/lexer"
	"inter/token"
	"strconv"
)

const (
	_ = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
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
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer

	curToken       token.Token
	peekToken      token.Token
	errors         []string
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:              l,
		errors:         []string{},
		prefixParseFns: make(map[token.TokenType]prefixParseFn),
		infixParseFns:  make(map[token.TokenType]infixParseFn),
	}

	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerIdentifier)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(token.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) registerPrefix(tType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tType] = fn
}

func (p *Parser) registerInfix(tType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tType] = fn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	st := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	// TODO: handle the value exp later
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	// To skip the semicolon
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return st
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	st := &ast.LetStatement{Token: p.curToken}

	if !p.peekTokenIsThenAdvance(token.IDENT) {
		return nil
	}

	st.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.peekTokenIsThenAdvance(token.ASSIGN) {
		return nil
	}

	// TODO: Skip for now
	for !p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	// To skip the semicolon
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return st
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	st := &ast.ExpressionStatement{Token: p.curToken}

	st.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return st
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) peekTokenIsThenAdvance(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}

	p.peekError(t)
	return false
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		st := p.parseStatement()
		program.Statements = append(program.Statements, st)

		p.nextToken()
	}

	return program
}

func (p *Parser) peekError(t token.TokenType) {
	err := fmt.Sprintf("Expected=%q, got=%q", t, p.peekToken.Type)
	p.errors = append(p.errors, err)
}

func (p *Parser) parseIfExpression() ast.Expression {
	ifExp := &ast.IfExpression{Token: p.curToken}

	if !p.peekTokenIsThenAdvance(token.LPAREN) {
		return nil
	}

	// Now I'm at ( token
	p.nextToken()
	ifExp.Condition = p.parseExpression(LOWEST)

	if !p.peekTokenIsThenAdvance(token.RPAREN) {
		return nil
	}

	if !p.peekTokenIsThenAdvance(token.LBRACE) {
		return nil
	}

	ifExp.Body = p.parseBlockStatements()

	if p.peekTokenIsThenAdvance(token.ELSE) {
		if !p.peekTokenIsThenAdvance(token.LBRACE) {
			return nil
		}

		ifExp.ElseBody = p.parseBlockStatements()
	}

	return ifExp
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	ids := []*ast.Identifier{}
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return ids
	}

	p.nextToken()
	id := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	ids = append(ids, id)
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		id := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		ids = append(ids, id)
	}

	if !p.peekTokenIsThenAdvance(token.RPAREN) {
		return nil
	}
	return ids
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	res := &ast.FunctionLiteral{Token: p.curToken}
	res.FunctionParameters = []*ast.Identifier{}

	if !p.peekTokenIsThenAdvance(token.LPAREN) {
		return nil
	}

	res.FunctionParameters = p.parseFunctionParameters()

	if !p.peekTokenIsThenAdvance(token.LBRACE) {
		return nil
	}

	res.FunctionBody = p.parseBlockStatements()

	return res
}

func (p *Parser) parseBlockStatements() *ast.BlockStatement {
	st := &ast.BlockStatement{Token: p.curToken}
	st.Statements = []ast.Statement{}
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		curSt := p.parseStatement()
		st.Statements = append(st.Statements, curSt)
		p.nextToken()
	}

	return st
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	// defer untrace(trace("parseExpression"))

	prefixFn := p.prefixParseFns[p.curToken.Type]

	if prefixFn == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefixFn()

	for !p.peekTokenIs(token.SEMICOLON) && p.peekPrecedence() > precedence {
		infixFn := p.infixParseFns[p.peekToken.Type]
		if infixFn == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infixFn(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerIdentifier() ast.Expression {
	exp := &ast.IntegerLiteral{Token: p.curToken}

	val, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("Cannot parse %s as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	exp.Value = val
	return exp
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	exp := &ast.PrefixExpression{Token: p.curToken}
	exp.Operator = p.curToken.Literal

	p.nextToken()
	exp.Right = p.parseExpression(PREFIX)
	return exp
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	exp := &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
	return exp
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)
	if !p.peekTokenIs(token.RPAREN) {
		p.errors = append(p.errors, "Unmatched left ( found")
		return nil
	}
	p.nextToken()
	return exp
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseInfixExpression"))

	exp := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedence)
	return exp
}

func (p *Parser) peekPrecedence() int {
	pre, ok := precedences[p.peekToken.Type]
	if !ok {
		return LOWEST
	}

	return pre
}

func (p *Parser) curPrecedence() int {
	pre, ok := precedences[p.curToken.Type]
	if !ok {
		return LOWEST
	}

	return pre
}
