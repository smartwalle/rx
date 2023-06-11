package rx

func (n *node) getNode(path string) *node {
walk: // Outer loop for walking the tree
	for {
		prefix := n.path
		if path == prefix {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if n.route != nil {
				return n
			}

			if path == "/" && n.wildChild && n.nType != root {
				return nil
			}

			// No handle found. Check if a handle for this path + a
			// trailing slash exists for trailing slash recommendation
			indices := n.indices
			for i, max := 0, len(indices); i < max; i++ {
				if indices[i] == '/' {
					return nil
				}
			}

			return nil
		}

		if len(path) > len(prefix) && path[:len(prefix)] == prefix {
			path = path[len(prefix):]
			// If this node does not have a wildcard (param or catchAll)
			// child,  we can just look up the next child node and continue
			// to walk down the tree
			if !n.wildChild {
				c := path[0]
				indices := n.indices
				for i, max := 0, len(indices); i < max; i++ {
					if c == indices[i] {
						n = n.children[i]
						prefix = n.path
						continue walk
					}
				}

				// Nothing found.
				// We can recommend to redirect to the same URL without a
				// trailing slash if a leaf exists for that path.
				return nil
			}

			// handle wildcard child
			n = n.children[0]
			switch n.nType {
			case param:
				// find param end (either '/' or path end)
				end := 0
				for end < len(path) && path[end] != '/' {
					end++
				}

				// we need to go deeper!
				if end < len(path) {
					if len(n.children) > 0 {
						path = path[end:]
						n = n.children[0]
						prefix = n.path
						continue walk
					}
					return nil
				}

				if n.route != nil {
					return n
				}
				if len(n.children) == 1 {
					// No handle found. Check if a handle for this path + a
					// trailing slash exists for TSR recommendation
					n = n.children[0]
				}
				return nil

			case catchAll:
				return n

			default:
				panic("invalid node type")
			}
		}
		return nil
	}
}
