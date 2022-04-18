package znet

import (
	"strings"
)

type (
	// Tree records node
	Tree struct {
		root       *Node
		parameters Parameters
		routes     map[string]*Node
	}

	// Node records any URL params, and executes an end handler.
	Node struct {
		key string
		// path records a request path
		path   string
		handle HandlerFunc
		// depth records Node's depth
		depth int
		// children records Node's children node
		children map[string]*Node
		// isPattern flag
		isPattern bool
		// middleware records middleware stack
		middleware []HandlerFunc
	}
)

func NewNode(key string, depth int) *Node {
	return &Node{
		key:      key,
		depth:    depth,
		children: make(map[string]*Node),
	}
}

func NewTree() *Tree {
	return &Tree{
		root:   NewNode("/", 1),
		routes: make(map[string]*Node),
	}
}

func (t *Tree) Add(pattern string, handle HandlerFunc, middleware ...HandlerFunc) {
	var currentNode = t.root
	if pattern != currentNode.key {
		res := strings.Split(pattern, "/")
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
				if len(middleware) > 0 {
					node.middleware = append(node.middleware, middleware...)
				}
				currentNode.children[key] = node
			}
			currentNode = node
		}
	}

	if len(middleware) > 0 && currentNode.depth == 1 {
		currentNode.middleware = append(currentNode.middleware, middleware...)
	}
	currentNode.handle = handle
	currentNode.isPattern = true
	currentNode.path = pattern
	if routeName := t.parameters.routeName; routeName != "" {
		t.routes[routeName] = currentNode
	}

}

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
