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

type Key struct {
	X int
	Y int
}

func (p Key) String() string {
	return fmt.Sprintf("(%d, %d)", p.Y, p.X)
}

func FlatGraph(cols, rows uint, cost Cost) map[Key]Cost {
	costs := make(map[Key]Cost)
	for col := range lo.Range(int(cols)) {
		for row := range lo.Range(int(rows)) {
			pos := Key{X: row, Y: col}
			costs[pos] = cost
		}
	}
	return costs
}

func RandomWallGraph(cols, rows uint) map[Key]Cost {
	costs := make(map[Key]Cost)
	for pos, cost := range FlatGraph(cols, rows, 1) {
		if rand.Intn(7) == 0 {
			continue
		}
		costs[pos] = cost
	}
	return costs
}

func MockOptions(graph map[Key]Cost) dijkstra.Options[Key, Cost] {
	return dijkstra.Options[Key, Cost]{
		Accumulator: func(agg Cost, from, to Key) (next Cost, ok bool) {
			cost, ok := graph[to]
			return agg + cost, ok
		},
		Less: func(i, j Cost) bool {
			return i < j
		},
		Edges: func(p Key) (edges []Key) {
			for _, to := range []Key{
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
	costs := options.Dijkstra(Key{X: 0, Y: 0}, Cost(0))
	t.Log("\n" + Graph2Text(graph) + "\n" + Graph2Text(Costs2Graph(costs)))
	path := lo.Must(options.ShortestPath(costs, Key{X: 5, Y: 5}))
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
	finder := options.CreatePathFinder(Key{X: 0, Y: 0}, Cost(0))
	path := lo.Must(finder(Key{X: 5, Y: 5}))
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
	costs := options.Dijkstra(Key{X: 0, Y: 0}, Cost(0))
	t.Log("\n" + Graph2Text(graph) + "\n" + Graph2Text(Costs2Graph(costs)))
	_, err := options.ShortestPath(costs, Key{X: 5, Y: 5})
	var notReachableErr *dijkstra.NotReachableError[Key, Cost]
	a.ErrorAs(err, &notReachableErr)
}

func TestOverGraphEdges(t *testing.T) {
	a := assert.New(t)
	graph := FlatGraph(10, 8, 1)
	options := MockOptions(graph)
	options.Edges = func(p Key) []Key {
		return []Key{
			{X: p.X, Y: p.Y + 1},
			{X: p.X, Y: p.Y - 1},
			{X: p.X + 1, Y: p.Y},
			{X: p.X - 1, Y: p.Y},
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	go func() {
		defer cancel()
		costs := options.Dijkstra(Key{X: 0, Y: 0}, Cost(0))
		t.Log("\n" + Graph2Text(graph) + "\n" + Graph2Text(Costs2Graph(costs)))
		path := lo.Must(options.ShortestPath(costs, Key{X: 5, Y: 5}))
		t.Log(path)
	}()
	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		a.Fail("Dijkstra took too long")
	}
}

func Costs2Graph(costs map[Key]dijkstra.Node[Key, Cost]) map[Key]Cost {
	graph := make(map[Key]Cost)
	for node, cost := range costs {
		graph[node] = cost.Cost
	}
	return graph
}

func Graph2Text(graph map[Key]Cost) string {
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
			cost, exists := graph[Key{X: row, Y: col}]
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

func Text2Graph(text string) map[Key]Cost {
	graph := make(map[Key]Cost)
	for row, line := range strings.Split(strings.TrimSpace(text), "\n") {
		for col, cell := range strings.Fields(line) {
			if cost, err := strconv.Atoi(cell); err == nil {
				graph[Key{X: row, Y: col}] = Cost(cost)
			}
		}
	}
	return graph
}
