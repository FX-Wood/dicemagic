package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/aasmall/dicemagic/dicelang"
)

func main() {
	var path, cmd string
	flag.StringVar(&path, "path", "", "Path to a file with one roll command per line.")
	flag.StringVar(&cmd, "cmd", "roll 1d20", "Roll command")
	flag.Parse()
	if path == "" {
		fmt.Println(cmd)
		printDiceInfo(cmd)
	} else {
		c := make(chan string)
		go readRollsFromFile(c, path)
		for cmd := range c {
			fmt.Println(cmd)
			printDiceInfo(cmd)
		}
	}
}

func printDiceInfo(cmd string) {
	var p *dicelang.Parser
	p = dicelang.NewParser(cmd)
	stmts, err := p.Statements()
	if err != nil {
		fmt.Println(err.Error(), err.(dicelang.LexError).Col, err.(dicelang.LexError).Line)
		return
	}
	for _, stmt := range stmts {
		dicelang.PrintAST(stmt, 0)
		total, diceSet, err := stmt.GetDiceSet()
		if err != nil {
			fmt.Printf("Could not parse input: %v\n", err)
			return
		}
		for _, v := range diceSet.Dice {
			fmt.Printf("%+v\nProbability: %.2f%%\n", v, v.Probability("==", v.Total)*100)
		}
		fmt.Printf("for a total of: %+v\n", total)
		fmt.Printf("map: %+v\n", diceSet.TotalsByColor)
		fmt.Println("----------")
	}
}

func readRollsFromFile(c chan string, path string) {
	defer close(c)
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Could not open file: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		c <- scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Could not scan file: %v\n", err)
		return
	}
}
