package znet

import (
	"strings"
)

type (
	// Tree records node
	Tree struct {
		root       *Node
		routes     map[string]*Node
		parameters Parameters
	}

	// Node records any URL params, and executes an end handlerFn.
	Node struct {
		handle     handlerFn
		children   map[string]*Node
		key        string
		path       string
		middleware []handlerFn
		depth      int
		isPattern  bool
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

func (t *Tree) Add(path string, handle handlerFn, middleware ...handlerFn) {
	var currentNode = t.root
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
