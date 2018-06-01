package main

import (
	"fmt"
)

type nudFn func(*ast, *parser) *ast

type ledFn func(*ast, *parser, *ast) *ast

type stdFn func(*ast, *parser) *ast

type ast struct {
	sym          string
	value        string
	line         int
	col          int
	bindingPower int
	nud          nudFn
	led          ledFn
	std          stdFn
	children     []*ast
}

type parser struct {
	lexer *lexer
}

func (self *parser) expression(rbp int) *ast {
	var left *ast
	t := self.lexer.next()

	if t.nud != nil {
		left = t.nud(t, self)
	} else {
		panic(fmt.Sprint("NOT PREFIX", t))
	}
	for rbp < self.lexer.peek().bindingPower {
		t := self.lexer.next()
		if t.led != nil {
			left = t.led(t, self, left)
		} else {
			panic(fmt.Sprint("NOT INFIX", t))
		}
	}

	return left
}

func (self *parser) statements() []*ast {
	stmts := []*ast{}
	next := self.lexer.peek()
	for next.sym != "(EOF)" && next.sym != "}" {
		stmt := self.statement()
		if stmt.sym != "(EOF)" {
			stmts = append(stmts, stmt)
		}
		next = self.lexer.peek()
	}
	return stmts
}

func (self *parser) block() *ast {
	tok := self.lexer.next()
	if tok.sym != "{" {
		panic(fmt.Sprint("WAS LOOKING FOR BLOCK START", tok))
	}
	block := tok.std(tok, self)
	return block
}

func (self *parser) statement() *ast {
	tok := self.lexer.peek()
	if tok.std != nil {
		tok = self.lexer.next()
		return tok.std(tok, self)
	}
	res := self.expression(0)

	return res
}

func (self *parser) advance(sym string) *ast {
	line := self.lexer.line
	col := self.lexer.col
	token := self.lexer.next()
	if token.sym != sym {
		panic(fmt.Sprint("EXPECTED ", sym, " AT ", line, ":", col))
	}
	return token
}
