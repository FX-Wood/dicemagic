package dicelang

import (
	"fmt"
)

type nudFn func(*AST, *Parser) (*AST, error)

type ledFn func(*AST, *Parser, *AST) (*AST, error)

type stdFn func(*AST, *Parser) (*AST, error)

//AST represents a node in an abstract syntax tree
type AST struct {
	sym          string
	value        string
	line         int
	col          int
	bindingPower int
	nud          nudFn
	led          ledFn
	std          stdFn
	children     []*AST
}

// Parser holds a Lexer and implements a top down operator precedence parser (https://tdop.github.io/)
// credit to: https://github.com/cristiandima/tdop for most of this code.
type Parser struct {
	lexer *Lexer
}

//NewParser creates a new Parser from an input string
func NewParser(source string) *Parser {
	l := NewLexer(source)
	return &Parser{lexer: l}
}

func (parse *Parser) expression(rbp int) (*AST, error) {
	var left *AST
	t, err := parse.lexer.next()
	if err != nil {
		return nil, err
	}

	if t.nud != nil {
		left, _ = t.nud(t, parse)
	} else {
		return nil, LexError{err: fmt.Sprintf("token \"%s\" is not prefix", t.value), Col: parse.lexer.col, Line: parse.lexer.line}
	}
	t, err = parse.lexer.peek()
	if err != nil {
		return nil, err
	}
	for rbp < t.bindingPower {
		t, err = parse.lexer.next()
		if err != nil {
			return nil, err
		}
		if t.led != nil {
			left, err = t.led(t, parse, left)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, LexError{err: fmt.Sprintf("token \"%s\" is not infix", t.value), Col: parse.lexer.col, Line: parse.lexer.line}
		}
		t, err = parse.lexer.peek()
		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

//Statements returns all statements from the parser as []*AST
func (parse *Parser) Statements() ([]*AST, error) {
	stmts := []*AST{}
	next, err := parse.lexer.peek()
	if err != nil {
		return nil, err
	}
	for next.sym != "(EOF)" && next.sym != "}" {
		stmt, err := parse.Statement()
		if err != nil {
			return nil, err
		}
		if stmt.sym != "(EOF)" {
			stmts = append(stmts, stmt)
		}
		next, err = parse.lexer.peek()
		if err != nil {
			return nil, err
		}
	}
	return stmts, nil
}

func (parse *Parser) block() (*AST, error) {
	tok, err := parse.lexer.next()
	if err != nil {
		return nil, err
	}
	if tok.sym != "{" {
		return nil, LexError{err: fmt.Sprintf("expected block start not found: %s", tok.sym), Col: parse.lexer.col, Line: parse.lexer.line}
	}
	return tok.std(tok, parse)
}

//Statement returns the next statement from the parser as *AST
func (parse *Parser) Statement() (*AST, error) {
	tok, err := parse.lexer.peek()
	if err != nil {
		return nil, err
	}
	if tok.std != nil {
		tok, err = parse.lexer.next()
		if err != nil {
			return nil, err
		}
		return tok.std(tok, parse)
	}
	return parse.expression(0)
}

func (parse *Parser) advance(sym string) (*AST, error) {
	line := parse.lexer.line
	col := parse.lexer.col
	token, err := parse.lexer.next()

	if err != nil {
		return nil, err
	}
	if token.sym != sym {
		return nil, LexError{err: fmt.Sprintf("did not find expected character \"%s\". Found \"%s\"", sym, token.sym), Col: col, Line: line}
	}
	return token, nil
}
