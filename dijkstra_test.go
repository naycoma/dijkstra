package dijkstra_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/naycoma/dijkstra"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

type Cost uint

type Node struct {
	X int
	Y int
}

func (p Node) String() string {
	return fmt.Sprintf("(%d, %d)", p.Y, p.X)
}

func FlatGraph(cols, rows uint, cost Cost) map[Node]Cost {
	costs := make(map[Node]Cost)
	for col := range lo.Range(int(cols)) {
		for row := range lo.Range(int(rows)) {
			pos := Node{X: row, Y: col}
			costs[pos] = cost
		}
	}
	return costs
}

func RandomWallGraph(cols, rows uint) map[Node]Cost {
	costs := make(map[Node]Cost)
	for pos, cost := range FlatGraph(cols, rows, 1) {
		if rand.Intn(7) == 0 {
			continue
		}
		costs[pos] = cost
	}
	return costs
}

func MockOptions(graph map[Node]Cost) dijkstra.Options[Node, Cost] {
	return dijkstra.Options[Node, Cost]{
		Accumulator: func(agg Cost, from, to Node) (next Cost, ok bool) {
			cost, ok := graph[to]
			return agg + cost, ok
		},
		Less: func(i, j Cost) bool {
			return i < j
		},
		Edges: func(p Node) (edges []Node) {
			for _, to := range []Node{
				{X: p.X, Y: p.Y + 1},
				{X: p.X, Y: p.Y - 1},
				{X: p.X + 1, Y: p.Y},
				{X: p.X - 1, Y: p.Y},
			} {
				if _, ok := graph[to]; ok {
					edges = append(edges, to)
				}
			}
			return edges
		},
	}
}

func TestReachable(t *testing.T) {
	graph := Text2Graph(`
	1  ■  1  1  1  1  1  1  ■  1 
	1  1  1  1  1  1  1  1  ■  1 
	1  1  1  1  1  1  1  1  1  1 
	■  1  1  1  1  1  1  1  1  1 
	■  1  1  1  ■  ■  ■  1  1  1 
	1  ■  1  1  ■  1  1  1  1  ■ 
	1  1  1  1  ■  1  ■  1  1  1 
	1  1  1  1  1  ■  1  1  1  1
	`)
	options := MockOptions(graph)
	costs := options.Dijkstra(Node{X: 0, Y: 0}, Cost(0))
	t.Log("\n" + Graph2Text(graph) + "\n" + Graph2Text(costs))
	path := lo.Must(options.PathResolve(costs, Node{X: 5, Y: 5}))
	t.Log(path)
}

func TestReachablePathFinder(t *testing.T) {
	graph := Text2Graph(`
	1  ■  1  1  1  1  1  1  ■  1 
	1  1  1  1  1  1  1  1  ■  1 
	1  1  1  1  1  1  1  1  1  1 
	■  1  1  1  1  1  1  1  1  1 
	■  1  1  1  ■  ■  ■  1  1  1 
	1  ■  1  1  ■  1  1  1  1  ■ 
	1  1  1  1  ■  1  ■  1  1  1 
	1  1  1  1  1  ■  1  1  1  1
	`)
	options := MockOptions(graph)
	finder := options.CreatePathFinder(Node{X: 0, Y: 0}, Cost(0))
	path := lo.Must(finder(Node{X: 5, Y: 5}))
	t.Log(path)
}

func TestUnreachable(t *testing.T) {
	a := assert.New(t)
	graph := Text2Graph(`
	1  1  1  ■  1  1  1  ■  1  ■ 
	1  1  1  ■  1  1  ■  1  1  1 
	1  1  1  ■  1  1  1  1  1  1 
	1  1  1  ■  1  1  1  ■  1  1 
	1  1  1  ■  1  ■  1  1  1  1 
	1  1  1  ■  1  1  1  1  1  1 
	1  1  1  ■  1  1  ■  1  1  1 
	1  1  1  ■  1  1  1  1  1  1
	`)
	options := MockOptions(graph)
	costs := options.Dijkstra(Node{X: 0, Y: 0}, Cost(0))
	t.Log("\n" + Graph2Text(graph) + "\n" + Graph2Text(costs))
	_, err := options.PathResolve(costs, Node{X: 5, Y: 5})
	var notReachableErr *dijkstra.NotReachableError[Node, Cost]
	a.ErrorAs(err, &notReachableErr)
}

func TestOverGraphEdges(t *testing.T) {
	a := assert.New(t)
	graph := FlatGraph(10, 8, 1)
	options := MockOptions(graph)
	options.Edges = func(p Node) []Node {
		return []Node{
			{X: p.X, Y: p.Y + 1},
			{X: p.X, Y: p.Y - 1},
			{X: p.X + 1, Y: p.Y},
			{X: p.X - 1, Y: p.Y},
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	go func() {
		defer cancel()
		costs := options.Dijkstra(Node{X: 0, Y: 0}, Cost(0))
		t.Log("\n" + Graph2Text(graph) + "\n" + Graph2Text(costs))
		path := lo.Must(options.PathResolve(costs, Node{X: 5, Y: 5}))
		t.Log(path)
	}()
	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		a.Fail("Dijkstra took too long")
	}
}

func Graph2Text(graph map[Node]Cost) string {
	var builder strings.Builder
	maxRow := 0
	maxCol := 0
	// Determine the size of the graph
	for pos := range graph {
		if pos.X > maxRow {
			maxRow = pos.X
		}
		if pos.Y > maxCol {
			maxCol = pos.Y
		}
	}
	// Generate the graph text representation
	for row := 0; row <= maxRow; row++ {
		for col := 0; col <= maxCol; col++ {
			cost, exists := graph[Node{X: row, Y: col}]
			if exists {
				builder.WriteString(fmt.Sprintf("%2d ", cost))
			} else {
				builder.WriteString(fmt.Sprintf("%2s ", "■"))
			}
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func Text2Graph(text string) map[Node]Cost {
	graph := make(map[Node]Cost)
	for row, line := range strings.Split(strings.TrimSpace(text), "\n") {
		for col, cell := range strings.Fields(line) {
			if cost, err := strconv.Atoi(cell); err == nil {
				graph[Node{X: row, Y: col}] = Cost(cost)
			}
		}
	}
	return graph
}
