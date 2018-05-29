package main

type nudFn func(*token, *parser) *token

type ledFn func(*token, *parser, *token) *token

type stdFn func(*token, *parser) *token

type token struct {
	sym          string
	value        string
	line         int
	col          int
	bindingPower int
	nud          nudFn
	led          ledFn
	std          stdFn
	children     []*token
}
