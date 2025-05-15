package znet

import (
	"strings"
)

type (
	// Tree represents a radix tree for HTTP routing.
	// It stores routes and their handlers in an efficient tree structure
	// for fast URL matching and parameter extraction.
	Tree struct {
		root       *Node            // Root node of the routing tree
		routes     map[string]*Node // Named routes for reverse routing
		parameters Parameters       // Route parameters and configuration
	}

	// Node represents a single node in the routing tree.
	// Each node can have a handler function and child nodes.
	// Nodes can represent static path segments or pattern parameters.
	Node struct {
		value      interface{}      // Custom data associated with this node
		handle     handlerFn        // Handler function for this route
		children   map[string]*Node // Child nodes indexed by path segment
		engine     *Engine          // Reference to the engine instance
		key        string           // Path segment this node represents
		path       string           // Full path from root to this node
		middleware []handlerFn      // Middleware functions for this route
		depth      int              // Depth in the tree (distance from root)
		isPattern  bool             // Whether this node represents a complete route pattern
	}
)

// NewNode creates a new tree node with the given key and depth.
// The key represents the path segment this node matches.
func NewNode(key string, depth int) *Node {
	return &Node{
		key:      key,
		depth:    depth,
		children: make(map[string]*Node),
	}
}

// WithValue associates a value with this node and returns the node.
// This allows for method chaining when building the tree.
func (t *Node) WithValue(v interface{}) *Node {
	t.value = v
	return t
}

// Value returns the value associated with this node.
func (t *Node) Value() interface{} {
	return t.value
}

// Path returns the full path that this node represents.
func (t *Node) Path() string {
	return t.path
}

// Handle returns the handler function associated with this node.
func (t *Node) Handle() handlerFn {
	return t.handle
}

// NewTree creates a new routing tree with a root node.
// The root node represents the '/' path and serves as the starting point
// for all route matching operations.
func NewTree() *Tree {
	return &Tree{
		root:   NewNode("/", 1),
		routes: make(map[string]*Node),
	}
}

// Add registers a new path with its handler and middleware in the routing tree.
// It splits the path into segments and creates nodes as needed for each segment.
// Middleware functions are attached to the leaf node or root node as appropriate.
// Returns the leaf node that was created or updated, which can be used for further configuration.
func (t *Tree) Add(e *Engine, path string, handle handlerFn, middleware ...handlerFn) (currentNode *Node) {
	currentNode = t.root
	wareLen := len(middleware)
	if path != currentNode.key {
		res := strings.Split(path, "/")
		end := len(res) - 1
		for i, key := range res {
			if key == "" {
				if i != end {
					continue
				}
				key = "/"
			}
			node, ok := currentNode.children[key]
			if !ok {
				node = NewNode(key, currentNode.depth+1)
				if wareLen > 0 && i == end {
					node.middleware = append(node.middleware, middleware...)
				}
				currentNode.children[key] = node
			} else if node.handle == nil {
				if wareLen > 0 && i == end {
					node.middleware = append(node.middleware, middleware...)
				}
			}
			currentNode = node
		}
	}

	if wareLen > 0 && currentNode.depth == 1 {
		currentNode.middleware = append(currentNode.middleware, middleware...)
	}

	currentNode.handle = handle
	currentNode.isPattern = true
	currentNode.path = path
	currentNode.engine = e
	if routeName := t.parameters.routeName; routeName != "" {
		t.routes[routeName] = currentNode
	}
	return
}

// Find searches for nodes that match the given pattern in the routing tree.
// If isRegex is true, it will return all nodes that match the pattern as a prefix,
// which is useful for finding all routes under a specific path.
// Otherwise, it will only return nodes that exactly match the pattern.
// Returns a slice of matching nodes, which may be empty if no matches are found.
func (t *Tree) Find(pattern string, isRegex bool) (nodes []*Node) {
	var (
		node  = t.root
		queue []*Node
	)
	if pattern == node.path {
		nodes = append(nodes, node)
		return
	}

	res := strings.Split(pattern, "/")
	for i := range res {
		key := res[i]
		if key == "" {
			continue
		}
		child, ok := node.children[key]
		if !ok && isRegex {
			break
		}
		if !ok && !isRegex {
			return
		}
		if pattern == child.path && !isRegex {
			nodes = append(nodes, child)
			return
		}
		node = child
	}

	queue = append(queue, node)
	for len(queue) > 0 {
		var queueTemp []*Node
		for _, n := range queue {
			if n.isPattern {
				nodes = append(nodes, n)
			}
			for _, childNode := range n.children {
				queueTemp = append(queueTemp, childNode)
			}
		}
		queue = queueTemp
	}
	return
}
