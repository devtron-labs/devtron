/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package util

func TopoSort(graph map[int][]int) []int {
	var sorted []int
	inDegree := map[int]int{}

	// 01. Calculate this.indegree of all vertices by going through every edge of the graph;
	// Each child gets indegree++ during breadth-first run.
	for element, children := range graph {
		if inDegree[element] == 0 {
			inDegree[element] = 0
		}
		for _, child := range children {
			inDegree[child]++
		}
	}

	// 02. Collect all vertices with indegree==0 onto a stack;
	var stack []int
	for rule, value := range inDegree {
		if value == 0 {
			stack = append(stack, rule)
			inDegree[rule] = -1
		}
	}

	// 03. While zero-degree-stack is not empty:
	for len(stack) > 0 {
		// 03.01. Pop element from zero-degree-stack and append it to topological order;
		var node int = stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// 03.02. Find all children of element and decrease indegree. If indegree becomes 0, add to zero-degree-stack;
		for _, child := range graph[node] {
			inDegree[child]--
			if inDegree[child] == 0 {
				stack = append(stack, child)
				inDegree[child] = -1
			}
		}

		// 03.03. Append to the sorted list.
		sorted = append(sorted, node)
	}
	return sorted
}
