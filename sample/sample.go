package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/naycoma/dijkstra"
)

type Cost uint

type Node struct {
	Y int
	X int
}

func main() {
	graph := Text2Graph(`
	1  ■  ■  1  1  1  ■  1 
	1  1  1  1  ■  1  1  1 
	■  ■  1  ■  ■  ■  1  ■ 
	■  1  ■  1  1  1  1  1 
	■  1  ■  1  ■  ■  ■  1 
	■  1  ■  1  ■  1  1  1 
	■  1  1  1  ■  1  ■  1 
	`)
	options := dijkstra.Options[Node, Cost]{
		Accumulator: func(agg Cost, key Node) (Cost, bool) {
			cost, ok := graph[key]
			return agg + cost, ok
		},
		Less: func(i, j Cost) bool {
			return i < j
		},
		Edges: func(p Node) []Node {
			return []Node{
				{Y: p.Y, X: p.X + 1},
				{Y: p.Y, X: p.X - 1},
				{Y: p.Y + 1, X: p.X},
				{Y: p.Y - 1, X: p.X},
			}
		},
	}

	start := Node{Y: 0, X: 0}
	goal := Node{Y: 5, X: 5}

	pathFinder := options.CreatePathFinder(start, Cost(0))
	path, err := pathFinder(goal)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, pos := range path {
		fmt.Printf("(%d, %d) ", pos.Y, pos.X)
	}
}

func Text2Graph(text string) map[Node]Cost {
	graph := make(map[Node]Cost)
	for row, line := range strings.Split(strings.TrimSpace(text), "\n") {
		for col, cell := range strings.Fields(line) {
			if cost, err := strconv.Atoi(cell); err == nil {
				graph[Node{Y: row, X: col}] = Cost(cost)
			}
		}
	}
	return graph
}
