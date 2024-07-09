package dijkstra

import (
	"container/heap"
	"fmt"
)

type Node[K comparable, C any] struct {
	Key  K
	Cost C
	Prev *K
}

type heapNodes[K comparable, C any] struct {
	nodes []*Node[K, C]
	less  func(i, j C) bool
}

func (pq *heapNodes[K, C]) Len() int {
	return len(pq.nodes)
}

func (pq *heapNodes[K, C]) Less(i, j int) bool {
	return pq.less(pq.nodes[i].Cost, pq.nodes[j].Cost)
}

func (pq *heapNodes[K, C]) Swap(i, j int) {
	pq.nodes[i], pq.nodes[j] = pq.nodes[j], pq.nodes[i]
}

func (pq *heapNodes[K, C]) Push(x any) {
	pq.nodes = append(pq.nodes, x.(*Node[K, C]))
}

func (pq *heapNodes[K, C]) Pop() any {
	last := pq.nodes[len(pq.nodes)-1]
	pq.nodes = pq.nodes[:len(pq.nodes)-1]
	return last
}

var _ heap.Interface = (*heapNodes[int, int])(nil)

// priorityNodes is a priority queue for nodes and their costs.
type priorityNodes[K comparable, C any] struct {
	*heapNodes[K, C]
}

func newPriorityNodes[K comparable, C any](less func(i, j C) bool) *priorityNodes[K, C] {
	h := &heapNodes[K, C]{
		nodes: []*Node[K, C]{},
		less:  less,
	}
	heap.Init(h)
	return &priorityNodes[K, C]{h}
}

// Extracts the minimum cost node from the priority queue
func (pq *priorityNodes[K, C]) Pop() (current K, prev *K, cost C) {
	nc := heap.Pop(pq.heapNodes).(*Node[K, C])
	return nc.Key, nc.Prev, nc.Cost
}

func (pq *priorityNodes[K, C]) Push(current K, prev *K, cost C) {
	heap.Push(pq.heapNodes, &Node[K, C]{Key: current, Prev: prev, Cost: cost})
}

func (pq *priorityNodes[K, C]) Empty() bool {
	return pq.heapNodes.Len() == 0
}

// Dijkstra runs Dijkstra's algorithm with the given options.
// accumulator : Function to accumulate costs from one node to another.
// initial : The initial cost to reach the start node.
// less : Comparison function to determine the order of costs.
// edges : Function to retrieve adjacent nodes.
// returns : The costs to reach each node from the start node.
func Dijkstra[K comparable, C any](
	start K,
	accumulator func(agg C, from, to K) (next C, ok bool),
	initial C,
	less func(i C, j C) bool,
	edges func(from K) (dest []K),
) (costs map[K]Node[K, C]) {
	open := newPriorityNodes[K](less)
	costs = make(map[K]Node[K, C])

	open.Push(start, nil, initial)
	for !open.Empty() {
		current, prev, cost := open.Pop()
		if _, ok := costs[current]; ok {
			continue
		}
		costs[current] = Node[K, C]{Key: current, Cost: cost, Prev: prev}
		for _, dest := range edges(current) {
			if destCost, ok := accumulator(cost, current, dest); ok {
				open.Push(dest, &current, destCost)
			}
		}
	}
	return costs
}

// Options defines the options for running Dijkstra's algorithm.
// It includes the accumulator function to aggregate costs, a comparison function to determine order,
// and a function to retrieve adjacent nodes (edges).
type Options[K comparable, C any] struct {
	// Function to accumulate costs from one node to another.
	Accumulator func(agg C, from, to K) (next C, ok bool)
	// Comparison function to determine the order of costs.
	Less func(i C, j C) bool
	// Function to retrieve adjacent nodes.
	Edges func(from K) (dest []K)
}

// Dijkstra runs Dijkstra's algorithm with the given options.
func (c Options[K, C]) Dijkstra(start K, initial C) (costs map[K]Node[K, C]) {
	if c.Edges == nil {
		var k K
		if _, ok := any(k).(interface{ Adjacent() []K }); ok {
			c.Edges = func(key K) []K {
				return any(key).(interface{ Adjacent() []K }).Adjacent()
			}
		}
	}
	return Dijkstra(
		start,
		c.Accumulator,
		initial,
		c.Less,
		c.Edges,
	)
}

// ShortestPath resolves the path from the start node to the goal node.
func (c Options[K, C]) ShortestPath(costs map[K]Node[K, C], goal K) ([]K, error) {
	if _, ok := costs[goal]; !ok {
		return nil, newNotReachableError(costs, c.Less, goal)
	}
	path := []K{goal}
	for {
		current := path[0]
		node, ok := costs[current]
		if !ok {
			return nil, newNotReachableError(costs, c.Less, goal)
		}
		if node.Prev == nil {
			return path, nil
		}
		prev := *node.Prev
		path = append([]K{prev}, path...)
	}
}

// CreatePathFinder creates a function to find the path from the start node to any other node.
func (c Options[K, C]) CreatePathFinder(start K, initial C) (resolvePath func(goal K) ([]K, error)) {
	costs := c.Dijkstra(start, initial)
	return func(goal K) ([]K, error) {
		path, err := c.ShortestPath(costs, goal)
		if err != nil {
			return nil, err
		}
		return path, nil
	}
}

var _ error = &NotReachableError[int, int]{}

// NotReachableError indicates that the specified goal cannot be reached from the start node.
type NotReachableError[K comparable, C any] struct {
	Costs           map[K]Node[K, C]
	Start           K
	Goal            K
	StartingUnknown bool
}

func (e *NotReachableError[K, C]) Error() string {
	if e.StartingUnknown {
		return fmt.Sprintf("the specified goal is not reachable from the start node: %v", e.Goal)
	}
	return fmt.Sprintf("the specified goal is not reachable from the start node: %v -> %v", e.Start, e.Goal)
}

func newNotReachableError[K comparable, C any](costs map[K]Node[K, C], less func(i C, j C) bool, goal K) error {
	start, ok := minCostNode(getKeys(costs), costs, less)
	return &NotReachableError[K, C]{Costs: costs, Start: start, Goal: goal, StartingUnknown: !ok}
}

func getKeys[K comparable, V any](collection map[K]V) []K {
	keys := make([]K, len(collection))
	i := 0
	for key := range collection {
		keys[i] = key
		i++
	}
	return keys
}

func minCostNode[K comparable, C any](nodes []K, costs map[K]Node[K, C], less func(i C, j C) bool) (min K, ok bool) {
	return filterMinBy(nodes, func(node K, _ int) bool {
		_, ok := costs[node]
		return ok
	}, func(i K, j K) bool {
		return less(costs[i].Cost, costs[j].Cost)
	})
}

func filterMinBy[V any](collection []V, predicate func(item V, index int) bool, less func(i V, j V) bool) (min V, ok bool) {
	for index, item := range collection {
		if !predicate(item, index) {
			continue
		}
		if !ok {
			ok = true
			min = item
			continue
		}
		if less(item, min) {
			min = item
		}
	}
	return
}
