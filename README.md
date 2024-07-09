# Dijkstra's algorithm in Golang

Type-safe Dijkstra's algorithm implemented in Go.

## Features

- Flexible implementation using generics
- Customizable cost comparison functions to support multiple types
- Clean API for easy integration

## Installation

If you are using Go modules, you can install the package with the following
command:

```bash
go get github.com/naycoma/dijkstra
```

## Usage

### Example Usage

```go
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
```

> (0, 0) (1, 0) (1, 1) (1, 2) (1, 3) (0, 3) (0, 4) (0, 5) (1, 5) (1, 6) (2, 6) (3, 6) (3, 7) (4, 7) (5, 7) (5, 6) (5, 5)


### Options

You can customize the behavior of the algorithm using the `Options` struct.

## License

This project is licensed under the [MIT license](LICENSE).
