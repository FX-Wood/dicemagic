package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

func main() {
	args := os.Args
	if len(args) > 1 {
		buf, _ := ioutil.ReadFile(args[1])
		source := string(buf)

		l := tdopLexer(source)
		p := parser{lexer: l}

		stmts := p.statements()
		fmt.Println(source)
		for _, stmt := range stmts {
			printAST(stmt, 0)
			i, _ := Evaluate(stmt)
			fmt.Print("evaluated:", i)
			fmt.Println("\n--next statement--")
		}
	}
}

func printAST(t *token, identation int) {
	fmt.Println()
	for i := 0; i < identation; i++ {
		fmt.Print(" ")
	}
	fmt.Print("(")
	fmt.Print(t.sym, ":", t.value)
	if len(t.children) > 0 {
		for _, c := range t.children {
			fmt.Print(" ")
			printAST(c, identation+4)
		}
	}
	fmt.Print(")")
}

func Evaluate(t *token) (int64, error) {
	switch t.sym {
	case "(NUMBER)":
		i, _ := strconv.ParseInt(t.value, 10, 0)
		return i, nil
	case "d":
		//d is always binary
		l, err := Evaluate(t.children[0])
		if err != nil {
			return 0, err
		}
		r, err := Evaluate(t.children[1])
		if err != nil {
			return 0, err
		}
		//actually roll dice here
		return l * r, nil
	case "+":
		var x int64
		for _, c := range t.children {
			y, err := Evaluate(c)
			if err != nil {
				return 0, err
			}
			x += y
		}
		return x, nil
	case "(IDENT)":
		// apply type to all children
	case "if":
	}

	return 0, fmt.Errorf("bad")
}
