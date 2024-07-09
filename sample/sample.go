package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/naycoma/dijkstra"
)

type Cost uint

type Pos struct {
	Y int
	X int
}

func main() {
	costMap := Text2CostMap(`
	1  ■  ■  1  1  1  ■  1 
	1  1  1  1  ■  1  1  1 
	■  ■  1  ■  ■  ■  1  ■ 
	■  1  ■  1  1  1  1  1 
	■  1  ■  1  ■  ■  ■  1 
	■  1  ■  1  ■  1  1  1 
	■  1  1  1  ■  1  ■  1 
	`)
	options := dijkstra.Options[Pos, Cost]{
		Accumulator: func(agg Cost, from, to Pos) (Cost, bool) {
			cost, ok := costMap[to]
			return agg + cost, ok
		},
		Less: func(i, j Cost) bool {
			return i < j
		},
		Edges: func(p Pos) (edges []Pos) {
			for _, to := range []Pos{
				{Y: p.Y, X: p.X + 1},
				{Y: p.Y, X: p.X - 1},
				{Y: p.Y + 1, X: p.X},
				{Y: p.Y - 1, X: p.X},
			} {
				if _, ok := costMap[to]; ok {
					edges = append(edges, to)
				}
			}
			return edges
		},
	}

	start := Pos{Y: 0, X: 0}
	goal := Pos{Y: 5, X: 5}

	findPath := options.CreatePathFinder(start, Cost(0))
	path, err := findPath(goal)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, pos := range path {
		fmt.Printf("(%d, %d) ", pos.Y, pos.X)
	}
}

func Text2CostMap(text string) map[Pos]Cost {
	graph := make(map[Pos]Cost)
	for row, line := range strings.Split(strings.TrimSpace(text), "\n") {
		for col, cell := range strings.Fields(line) {
			if cost, err := strconv.Atoi(cell); err == nil {
				graph[Pos{Y: row, X: col}] = Cost(cost)
			}
		}
	}
	return graph
}
