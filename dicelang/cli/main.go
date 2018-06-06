package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/aasmall/dicemagic/dicelang"
)

func main() {
	var path, cmd string
	var verbose bool
	flag.StringVar(&path, "path", "", "Path to a file with one roll command per line.")
	flag.StringVar(&cmd, "cmd", "roll 1d20", "Roll command")
	flag.BoolVar(&verbose, "v", false, "Display ast and probability map for each statement")
	flag.Parse()
	if path == "" {
		fmt.Println(cmd)
		printDiceInfo(cmd, verbose)
	} else {
		c := make(chan string)
		go readRollsFromFile(c, path)
		for cmd := range c {
			fmt.Println(cmd)
			printDiceInfo(cmd, verbose)
		}
	}
}

func sortProbMap(m map[int64]float64) []int64 {
	var keys []int64
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func printDiceInfo(cmd string, verbose bool) {
	var p *dicelang.Parser
	p = dicelang.NewParser(cmd)
	stmts, err := p.Statements()
	if err != nil {
		fmt.Println(err.Error(), err.(dicelang.LexError).Col, err.(dicelang.LexError).Line)
		return
	}
	for i, stmt := range stmts {
		fmt.Printf("Statement %d\n", i+1)
		total, diceSet, err := stmt.GetDiceSet()
		if err != nil {
			fmt.Printf("Could not parse input: %v\n", err)
			return
		}
		if verbose {
			fmt.Print("AST:\n----------")
			dicelang.PrintAST(stmt, 0)
			fmt.Print("\n----------")
			for _, v := range diceSet.Dice {
				probMap := dicelang.DiceProbability(v.Count, v.Sides, v.DropHighest, v.DropLowest)
				keys := sortProbMap(probMap)
				fmt.Print("\nProbability Map:\n")
				for _, k := range keys {
					fmt.Printf("%2d:  %2.5F%%\n", k, probMap[k])
				}
				fmt.Print("----------\n")
			}
		}
		fmt.Printf("Total: %+v\n", total)
		fmt.Printf("Color Map: %+v\n", diceSet.TotalsByColor)
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
