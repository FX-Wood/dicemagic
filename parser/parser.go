package main

import (
	"fmt"
)

type parser struct {
	lexer *lexer
}

func (self *parser) expression(rbp int) *token {
	var left *token
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

func (self *parser) statements() []*token {
	stmts := []*token{}
	next := self.lexer.peek()
	for next.sym != "(EOF)" && next.sym != "}" {
		stmts = append(stmts, self.statement())
		next = self.lexer.peek()
	}
	return stmts
}

func (self *parser) block() *token {
	tok := self.lexer.next()
	if tok.sym != "{" {
		panic(fmt.Sprint("WAS LOOKING FOR BLOCK START", tok))
	}
	block := tok.std(tok, self)
	return block
}

func (self *parser) statement() *token {
	tok := self.lexer.peek()
	if tok.std != nil {
		tok = self.lexer.next()
		return tok.std(tok, self)
	}
	res := self.expression(0)

	return res
}

func (self *parser) advance(sym string) *token {
	line := self.lexer.line
	col := self.lexer.col
	token := self.lexer.next()
	if token.sym != sym {
		panic(fmt.Sprint("EXPECTED ", sym, " AT ", line, ":", col))
	}
	return token
}
