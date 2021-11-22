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
	for !p.tokenIs(token.SEMICOLON) {
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

func (p *Parser) tokenIs(t token.TokenType) bool {
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
		if st != nil {
			program.Statements = append(program.Statements, st)
		}

		p.nextToken()
	}

	return program
}

func (p *Parser) peekError(t token.TokenType) {
	err := fmt.Sprintf("Expected=%q, got=%q", t, p.peekToken.Type)
	p.errors = append(p.errors, err)
}

func (p *Parser) parseExpression(priority int) ast.Expression {
	prefixFn := p.prefixParseFns[p.curToken.Type]

	if prefixFn == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefixFn()
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
