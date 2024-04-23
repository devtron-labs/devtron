package util

// IsAncestor finds if a node1 is an ancestor of node2 in a n-ary tree
// given adjacency map of the tree as map[int][]int
func IsAncestor[T int | string](tree map[T][]T, source, dest T) bool {
	parentMappings := make(map[T]T)
	for parent, children := range tree {
		for _, child := range children {
			parentMappings[child] = parent
		}
	}
	ok := true
	for {
		if !ok {
			return false
		}

		if dest == source {
			return true
		}
		dest, ok = parentMappings[dest]
	}

}
