package main

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
