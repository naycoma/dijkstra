package dijkstra

import (
	"container/heap"
	"fmt"
)

type nodeAndCost[K comparable, R any] struct {
	Key  K
	Cost R
}

type heapNodes[K comparable, R any] struct {
	nodes []*nodeAndCost[K, R]
	less  func(i, j R) bool
}

func (pq *heapNodes[K, R]) Len() int {
	return len(pq.nodes)
}

func (pq *heapNodes[K, R]) Less(i, j int) bool {
	return pq.less(pq.nodes[i].Cost, pq.nodes[j].Cost)
}

func (pq *heapNodes[K, R]) Swap(i, j int) {
	pq.nodes[i], pq.nodes[j] = pq.nodes[j], pq.nodes[i]
}

func (pq *heapNodes[K, R]) Push(x any) {
	pq.nodes = append(pq.nodes, x.(*nodeAndCost[K, R]))
}

func (pq *heapNodes[K, R]) Pop() any {
	last := pq.nodes[len(pq.nodes)-1]
	pq.nodes = pq.nodes[:len(pq.nodes)-1]
	return last
}

var _ heap.Interface = (*heapNodes[int, int])(nil)

// priorityNodes is a priority queue for nodes and their costs.
type priorityNodes[K comparable, R any] struct {
	h *heapNodes[K, R]
}

func newPriorityNodes[K comparable, R any](less func(i, j R) bool) *priorityNodes[K, R] {
	h := &heapNodes[K, R]{
		nodes: []*nodeAndCost[K, R]{},
		less:  less,
	}
	heap.Init(h)
	return &priorityNodes[K, R]{h}
}

// Extracts the minimum cost node from the priority queue
func (pq *priorityNodes[K, R]) Pop() (node K, cost R) {
	nc := heap.Pop(pq.h).(*nodeAndCost[K, R])
	return nc.Key, nc.Cost
}

func (pq *priorityNodes[K, R]) Push(node K, cost R) {
	heap.Push(pq.h, &nodeAndCost[K, R]{Key: node, Cost: cost})
}

func (pq *priorityNodes[K, R]) Empty() bool {
	return pq.h.Len() == 0
}

// Dijkstra runs Dijkstra's algorithm with the given options.
// accumulator : Function to accumulate costs from one node to another.
// initial : The initial cost to reach the start node.
// less : Comparison function to determine the order of costs.
// edges : Function to retrieve adjacent nodes.
// returns : The costs to reach each node from the start node.
func Dijkstra[K comparable, R any](
	start K,
	accumulator func(agg R, key K) (next R, ok bool),
	initial R,
	less func(i R, j R) bool,
	edges func(K) []K,
) (costs map[K]R) {
	open := newPriorityNodes[K](less)
	costs = make(map[K]R)

	if startCost, ok := accumulator(initial, start); ok {
		open.Push(start, startCost)
	}
	for !open.Empty() {
		current, cost := open.Pop()
		if _, ok := costs[current]; ok {
			continue
		}
		costs[current] = cost
		for _, dest := range edges(current) {
			if destCost, ok := accumulator(cost, dest); ok {
				open.Push(dest, destCost)
			}
		}
	}
	return costs
}

// Options defines the options for running Dijkstra's algorithm.
// It includes the accumulator function to aggregate costs, a comparison function to determine order,
// and a function to retrieve adjacent nodes (edges).
type Options[K comparable, R any] struct {
	// Function to accumulate costs from one node to another.
	Accumulator func(agg R, key K) (next R, ok bool)
	// Comparison function to determine the order of costs.
	Less func(i R, j R) bool
	// Function to retrieve adjacent nodes.
	Edges func(K) []K
}

// Dijkstra runs Dijkstra's algorithm with the given options.
func (c Options[K, R]) Dijkstra(start K, initial R) (costs map[K]R) {
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

// PathResolve resolves the path from the start node to the goal node.
func (c Options[K, R]) PathResolve(costs map[K]R, goal K) ([]K, error) {
	if _, ok := costs[goal]; !ok {
		return nil, newNotReachableError(costs, c.Less, goal)
	}
	path := []K{goal}
	current := goal
	for {
		shortest, ok := minCostNode(c.Edges(current), costs, c.Less)
		if !ok {
			return nil, newNotReachableError(costs, c.Less, goal)
		}
		// If the cost is not decreasing, it means we have reached the start node.
		if !c.Less(costs[shortest], costs[current]) {
			return path, nil
		}
		path = append([]K{shortest}, path...)
		current = shortest
	}
}

// CreatePathFinder creates a function to find the path from the start node to any other node.
func (c Options[K, R]) CreatePathFinder(start K, initial R) func(goal K) ([]K, error) {
	costs := c.Dijkstra(start, initial)
	return func(goal K) ([]K, error) {
		path, err := c.PathResolve(costs, goal)
		if err != nil {
			return nil, err
		}
		return path, nil
	}
}

var _ error = &NotReachableError[int, int]{}

// NotReachableError indicates that the specified goal cannot be reached from the start node.
type NotReachableError[K comparable, R any] struct {
	Costs           map[K]R
	Start           K
	Goal            K
	StartingUnknown bool
}

func (e *NotReachableError[K, R]) Error() string {
	if e.StartingUnknown {
		return fmt.Sprintf("the specified goal is not reachable from the start node: %v", e.Goal)
	}
	return fmt.Sprintf("the specified goal is not reachable from the start node: %v -> %v", e.Start, e.Goal)
}

func newNotReachableError[K comparable, R any](costs map[K]R, less func(i R, j R) bool, goal K) error {
	start, ok := minCostNode(getKeys(costs), costs, less)
	return &NotReachableError[K, R]{Costs: costs, Start: start, Goal: goal, StartingUnknown: !ok}
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

func minCostNode[K comparable, R any](nodes []K, costs map[K]R, less func(i R, j R) bool) (min K, ok bool) {
	return filterMinBy(nodes, func(node K, _ int) bool {
		_, ok := costs[node]
		return ok
	}, func(i K, j K) bool {
		return less(costs[i], costs[j])
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
