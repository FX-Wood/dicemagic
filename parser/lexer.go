package main

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type lexer struct {
	tokReg *tokenRegistry
	source string
	index  int
	line   int
	col    int
	tok    *ast
	cached bool
	last   *ast
}

func (self *lexer) nextString() *ast {
	var text bytes.Buffer
	r, size := utf8.DecodeRuneInString(self.source[self.index:])
	for size > 0 {
		if r == '"' {
			self.col++
			self.index += size
			break
		}
		if r == '\n' {
			panic(fmt.Sprint("UNTERMINATED STRING AT ", self.line, ":", self.col))
		}
		if r == '\\' {
			self.col++
			self.index += size
			r, size = utf8.DecodeRuneInString(self.source[self.index:])
			if size > 0 && !unicode.IsSpace(r) {
				if r == 'r' {
					self.consumeRune(&text, '\r', size)
				} else if r == 'n' {
					self.consumeRune(&text, '\n', size)
				} else if r == 't' {
					self.consumeRune(&text, '\t', size)
				} else {
					self.consumeRune(&text, r, size)
				}
				r, size = utf8.DecodeRuneInString(self.source[self.index:])
				continue
			}
		}
		self.consumeRune(&text, r, size)
		r, size = utf8.DecodeRuneInString(self.source[self.index:])
	}
	return self.tokReg.token("(STRING)", text.String(), self.line, self.col)
}

func (self *lexer) nextOperator() *ast {
	var text bytes.Buffer
	r, size := utf8.DecodeRuneInString(self.source[self.index:])
	col := self.col
	self.consumeRune(&text, r, size)

	// try to parse operators made of two characters
	var twoChar bytes.Buffer
	twoChar.WriteRune(r)
	r, size = utf8.DecodeRuneInString(self.source[self.index:])
	if size > 0 && isOperatorChar(r) {
		twoChar.WriteRune(r)
		if self.tokReg.defined(twoChar.String()) {
			self.consumeRune(&text, r, size)
			textStr := text.String()
			return self.tokReg.token(textStr, textStr, self.line, col)
		}
	}

	// single character operator
	textStr := strings.ToLower(text.String())
	if !self.tokReg.defined(textStr) {
		panic("OPERATOR NOT DEFINED")
	}
	return self.tokReg.token(textStr, textStr, self.line, col)
}

func (self *lexer) nextIdent() *ast {
	var text bytes.Buffer
	col := self.col
	r, size := utf8.DecodeRuneInString(self.source[self.index:])
	if r == 'd' || r == 'D' {
		r1, _ := utf8.DecodeRuneInString(self.source[self.index+1:])
		if unicode.IsDigit(r1) {
			return self.nextOperator()
		}
	}
	self.consumeRune(&text, r, size)
	for {
		r, size = utf8.DecodeRuneInString(self.source[self.index:])
		if size > 0 && isIdentChar(r) {
			self.consumeRune(&text, r, size)
		} else {
			break
		}
	}
	symbol := text.String()
	if self.tokReg.defined(symbol) {
		return self.tokReg.token(symbol, symbol, self.line, col)
	}
	return self.tokReg.token("(IDENT)", symbol, self.line, col)
}
func (self *lexer) next() *ast {
	// invalidate peekable cache
	self.cached = false

	tmpIndex := -1
	for self.index != tmpIndex {
		tmpIndex = self.index
		self.consumeWhitespace()
		self.consumeComments()
	}

	// end of file
	if len(self.source[self.index:]) == 0 {
		return self.tokReg.token("(EOF)", "EOF", self.line, self.col)
	}

	var text bytes.Buffer
	r, size := utf8.DecodeRuneInString(self.source[self.index:])
	for size > 0 {
		if r == '"' { // parse string
			self.col++
			self.index += size
			return self.nextString()
		} else if isFirstIdentChar(r) { // parse identifiers/keywords
			return self.nextIdent()
		} else if unicode.IsDigit(r) { // parse numbers
			col := self.col
			self.consumeRune(&text, r, size)
			for {
				r, size = utf8.DecodeRuneInString(self.source[self.index:])
				if size > 0 && unicode.IsDigit(r) {
					self.consumeRune(&text, r, size)
				} else {
					break
				}
			}
			if size > 0 && r == '.' {
				self.consumeRune(&text, r, size)
				for {
					r, size = utf8.DecodeRuneInString(self.source[self.index:])
					if size > 0 && unicode.IsDigit(r) {
						self.consumeRune(&text, r, size)
					} else {
						break
					}
				}
			}
			return self.tokReg.token("(NUMBER)", text.String(), self.line, col)
		} else if r == '\n' {
			self.line++
			self.consumeRune(&text, r, size)
			return self.tokReg.token("(NEWLINE)", "\n", self.line-1, self.col)

		} else if isOperatorChar(r) { // parse operators
			return self.nextOperator()
		} else {
			break
		}
	}
	panic(fmt.Sprint("INVALID CHARACTER ", self.line, self.col))
}

func isSpaceNotNewline(r rune) bool {
	switch r {
	case '\t', '\v', '\f', '\r', ' ', 0x85, 0xA0:
		return true
	}
	return false
}
func (self *lexer) consumeWhitespace() {

	r, size := utf8.DecodeRuneInString(self.source[self.index:])
	for size > 0 && isSpaceNotNewline(r) {
		self.col++
		self.index += size
		r, size = utf8.DecodeRuneInString(self.source[self.index:])
	}
}

func (self *lexer) consumeComments() {
	r, size := utf8.DecodeRuneInString(self.source[self.index:])
	if r == '#' {
		for size > 0 && r != '\n' {
			self.col++
			self.index += size
			r, size = utf8.DecodeRuneInString(self.source[self.index:])
		}
	}
}

func (self *lexer) consumeRune(text *bytes.Buffer, r rune, size int) {
	text.WriteRune(r)
	self.col++
	self.index += size
}

func (self *lexer) peek() *ast {
	if self.cached {
		return self.tok
	}
	// save current state
	index := self.index
	line := self.line
	col := self.col

	// get token and cache it
	nextToken := self.next()
	self.tok = nextToken
	self.cached = true

	// restore state
	self.index = index
	self.line = line
	self.col = col

	return nextToken
}

func tdopLexer(source string) *lexer {
	return &lexer{tokReg: getTokenRegistry(), source: source, index: 0, line: 1, col: 1}
}

func getTokenRegistry() *tokenRegistry {
	t := &tokenRegistry{symTable: make(map[string]*ast)}

	t.symbol("(NUMBER)")
	t.symbol("(STRING)")

	t.symbol("true")
	t.symbol("false")
	t.symbol("none")

	t.consumable(";")
	t.consumable(")")
	t.consumable("]")
	t.consumable(",")
	t.consumable("else")

	t.consumable("(EOF)")
	t.consumable("{")
	t.consumable("}")
	t.consumable("roll")
	t.consumable("(NEWLINE)")

	t.infix("+", 50)
	t.infix("-", 50)

	t.infix("*", 60)
	t.infix("/", 60)
	t.infix("^", 70)
	t.infix("d", 80)
	t.infixLed("-L", 55, func(t *ast, p *parser, left *ast) *ast {
		next := p.lexer.peek()
		if next.sym == "(NUMBER)" {
			t.children = append(t.children, p.expression(t.bindingPower))
		}
		left.children = append(left.children, t)
		return left
	})
	t.infixLed("-H", 55, func(t *ast, p *parser, left *ast) *ast {
		next := p.lexer.peek()
		if next.sym == "(NUMBER)" {
			t.children = append(t.children, p.expression(t.bindingPower))
		}
		left.children = append(left.children, t)
		return left
	})

	t.infix("mod", 95)

	t.infix("<", 30)
	t.infix(">", 30)
	t.infix("<=", 30)
	t.infix(">=", 30)
	t.infix("==", 30)

	t.symbol("(IDENT)")

	t.infixLed("if", 20, func(token *ast, p *parser, left *ast) *ast {
		cond := p.expression(0)
		token.children = append(token.children, cond)
		p.advance("else")
		token.children = append(token.children, left)
		token.children = append(token.children, p.expression(0))
		return token
	})

	t.infixLed("(", 90, func(token *ast, p *parser, left *ast) *ast {
		if left.sym != "(IDENT)" && left.sym != "[" && left.sym != "(" && left.sym != "->" {
			panic(fmt.Sprint("BAD FUNC CALL LEFT OPERAND", left))
		}
		token.children = append(token.children, left)
		t := p.lexer.peek()
		if t.sym != ")" {
			for {
				exp := p.expression(0)
				token.children = append(token.children, exp)
				token := p.lexer.peek()
				if token.sym != "," {
					break
				}
				p.advance(",")
			}
			p.advance(")")
		} else {
			p.advance(")")
		}
		return token
	})

	t.infixLed("[", 80, func(token *ast, p *parser, left *ast) *ast {
		if left.sym != "(IDENT)" && left.sym != "[" && left.sym != "(" {
			panic(fmt.Sprint("BAD ARRAY LEFT OPERAND", left))
		}
		token.children = append(token.children, left)
		t := p.lexer.peek()
		if t.sym != "]" {
			for {
				exp := p.expression(0)
				token.children = append(token.children, exp)
				token := p.lexer.peek()
				if token.sym != "," {
					break
				}
				p.advance(",")
			}
			p.advance("]")
		} else {
			p.advance("]")
		}
		return token
	})

	t.infixRight("and", 25)
	t.infixRight("or", 25)

	t.infixRight("=", 10)
	t.infixRight("+=", 10)
	t.infixRight("-=", 10)

	t.infixRightLed("->", 10, func(token *ast, p *parser, left *ast) *ast {
		if left.sym != "()" && left.sym != "(IDENT)" {
			panic(fmt.Sprint("INVALID FUNC DECLARATION TUPLE", left))
		}
		if left.sym == "()" && len(left.children) != 0 {
			named := true
			for _, child := range left.children {
				if child.sym != "(IDENT)" {
					named = false
					break
				}
			}
			if !named {
				panic(fmt.Sprint("INVALID FUNC DECLARATION TUPLE", left))
			}
		}
		token.children = append(token.children, left)
		if p.lexer.peek().sym == "{" {
			token.children = append(token.children, p.block())
		} else {
			token.children = append(token.children, p.expression(0))
		}
		return token
	})

	t.prefix("-")
	t.prefix("not")

	t.prefixNud("(", func(t *ast, p *parser) *ast {
		comma := false
		if p.lexer.peek().sym != ")" {
			for {
				if p.lexer.peek().sym == ")" {
					break
				}
				t.children = append(t.children, p.expression(0))
				if p.lexer.peek().sym != "," {
					break
				}
				comma = true
				p.advance(",")
			}
		}
		p.advance(")")
		if len(t.children) == 0 || comma {
			t.sym = "()"
			t.value = "TUPLE"
			return t
		} else {
			return t.children[0]
		}
	})

	t.prefixNud("[", func(t *ast, p *parser) *ast {
		if p.lexer.peek().sym != "]" {
			for {
				if p.lexer.peek().sym == "]" {
					break
				}
				t.children = append(t.children, p.expression(0))
				if p.lexer.peek().sym != "," {
					break
				}
				p.advance(",")
			}
		}
		p.advance("]")
		t.sym = "[]"
		t.value = "ARRAY"
		return t
	})

	t.stmt("if", func(t *ast, p *parser) *ast {
		t.children = append(t.children, p.expression(0))
		t.children = append(t.children, p.block())
		next := p.lexer.peek()
		if next.value == "else" {
			p.lexer.next()
			next = p.lexer.peek()
			if next.value == "if" {
				t.children = append(t.children, p.statement())
			} else {
				t.children = append(t.children, p.block())
			}
		}
		return t
	})

	// t.stmt("-L", func(t *ast, p *parser) *ast {
	// 	t.children = append(t.children, p.statements()...)
	// 	return t
	// })
	// t.stmt("-H", func(t *ast, p *parser) *ast {
	// 	t.children = append(t.children, p.statements()...)
	// 	return t
	// })

	t.stmt("(IDENT)", func(t *ast, p *parser) *ast {
		t.children = append(t.children, p.statements()...)
		return t
	})
	t.stmt("(NEWLINE)", func(t *ast, p *parser) *ast {
		next := p.lexer.peek()
		for next.sym == "(NEWLINE)" {
			p.advance("(NEWLINE)")
			next = p.lexer.peek()
		}

		if next.sym == "(EOF)" {
			p.advance("(EOF)")
			return p.lexer.tokReg.token("(EOF)", "EOF", p.lexer.line, p.lexer.col)
		}
		return p.statement()
	})
	// t.prefixNud("roll", func(t *token, p *parser) *token {
	// 	t.children = append(t.children, p.statement())
	// 	t.children = append(t.children, p.expression(0))
	// 	return t
	// })
	t.stmt("while", func(t *ast, p *parser) *ast {
		t.children = append(t.children, p.expression(0))
		t.children = append(t.children, p.block())
		return t
	})

	t.stmt("roll", func(t *ast, p *parser) *ast {
		t.children = append(t.children, p.statement())
		return t
	})

	t.stmt("{", func(t *ast, p *parser) *ast {
		t.children = append(t.children, p.statements()...)
		p.advance("}")
		return t
	})
	t.stmt("return", func(t *ast, p *parser) *ast {
		if p.lexer.peek().sym != ";" {
			t.children = append(t.children, p.expression(0))
		}
		return t
	})

	return t
}

func isFirstIdentChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r == '_')
}

func isIdentChar(r rune) bool {
	return isFirstIdentChar(r) || unicode.IsDigit(r)
}

func isOperatorChar(r rune) bool {
	operators := "!@#$%^*()-+=/?.,:;\"|/{}[]><dDLH"
	for _, c := range operators {
		if c == r {
			return true
		}
	}
	return false
}
