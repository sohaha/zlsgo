//go:build go1.18
// +build go1.18

package zpool

import (
	"errors"
	"math"
	"sync/atomic"
	"time"

	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zsync"
	"github.com/sohaha/zlsgo/zutil"
)

// BalancerStrategy represents the load balancing strategy
type BalancerStrategy int

const (
	// StrategyRandom represents random selection strategy
	StrategyRandom BalancerStrategy = iota
	// StrategyLeastConn represents least connections selection strategy
	StrategyLeastConn
	// StrategyRoundRobin represents round-robin selection strategy
	StrategyRoundRobin
)

var (
	ErrKeyRequired      = errors.New("key is required")
	ErrNodeExists       = errors.New("node already exists")
	ErrNoAvailableNodes = errors.New("no available nodes")
	ErrEmptyCallback    = errors.New("callback function cannot be empty")
	ErrNoNodesAdded     = errors.New("please add nodes first")
)

type Balancer[T any] struct {
	nodes       map[string]*balancerNode[T]
	nodeKeys    []string
	mu          *zsync.RBMutex
	lastNodeIdx uint64
}

type balancerNode[T any] struct {
	node     T
	total    *zutil.Int64
	max      int64
	weight   uint64
	failedAt *zutil.Int64
	cooldown *zutil.Int64
}

type BalancerNodeOptions struct {
	// Maximum concurrent connections per node, fairness not guaranteed
	MaxConns int64
	// Node weight, default is 1
	Weight uint64
	// Node cooldown period after failure, default is 1000ms
	Cooldown int64
}

// NewBalancer creates a new load balancer
func NewBalancer[T any]() *Balancer[T] {
	return &Balancer[T]{
		nodes: make(map[string]*balancerNode[T], 8),
		mu:    zsync.NewRBMutex(),
	}
}

// Get returns the node with the given key
func (b *Balancer[T]) Get(key string) (node T, available bool, exists bool) {
	r := b.mu.RLock()
	defer b.mu.RUnlock(r)

	n, ok := b.nodes[key]
	if !ok {
		var d T
		return d, false, false
	}

	failedAt := n.failedAt.Load()
	isAvailable := !(failedAt > 0 && (time.Now().UnixMilli()-failedAt) <= n.cooldown.Load())
	return n.node, isAvailable, true
}

// Add adds a new node to the load balancer
func (b *Balancer[T]) Add(key string, node T, opt ...func(opts *BalancerNodeOptions)) error {
	if key == "" {
		return ErrKeyRequired
	}

	r := b.mu.RLock()
	_, exists := b.nodes[key]
	b.mu.RUnlock(r)

	if exists {
		return ErrNodeExists
	}

	o := zutil.Optional(BalancerNodeOptions{
		Weight:   1,
		Cooldown: 1000,
	}, opt...)

	n := &balancerNode[T]{
		node:     node,
		max:      o.MaxConns,
		weight:   o.Weight,
		total:    zutil.NewInt64(0),
		failedAt: zutil.NewInt64(0),
		cooldown: zutil.NewInt64(o.Cooldown),
	}
	b.mu.Lock()
	if _, exists := b.nodes[key]; exists {
		b.mu.Unlock()
		return ErrNodeExists
	}

	b.nodes[key] = n
	b.nodeKeys = append(b.nodeKeys, key)
	b.mu.Unlock()

	return nil
}

// Remove removes a node from the load balancer
func (b *Balancer[T]) Remove(key string) {
	b.mu.Lock()
	if _, exists := b.nodes[key]; !exists {
		b.mu.Unlock()
		return
	}

	delete(b.nodes, key)

	for i, k := range b.nodeKeys {
		if k == key {
			lastIdx := len(b.nodeKeys) - 1
			b.nodeKeys[i] = b.nodeKeys[lastIdx]
			b.nodeKeys = b.nodeKeys[:lastIdx]
			break
		}
	}

	b.mu.Unlock()
}

// getAvailableNodes returns all available nodes
func (b *Balancer[T]) getAvailableNodes(keys ...string) []*balancerNode[T] {
	r := b.mu.RLock()
	nodes := make([]*balancerNode[T], 0, len(b.nodes))
	now := time.Now().UnixMilli()

	for _, key := range b.nodeKeys {
		if len(keys) > 0 && !zarray.Contains(keys, key) {
			continue
		}
		node := b.nodes[key]
		failedAt := node.failedAt.Load()

		if failedAt > 0 && (now-failedAt) <= node.cooldown.Load() {
			continue
		}

		if node.max <= 0 || node.total.Load() <= node.max {
			nodes = append(nodes, node)
		}
	}
	b.mu.RUnlock(r)

	return nodes
}

// selectNode selects a node based on the given strategy
func (b *Balancer[T]) selectNode(nodes []*balancerNode[T], strategy BalancerStrategy) (*balancerNode[T], error) {
	nodeCount := len(nodes)
	if nodeCount == 0 {
		return nil, ErrNoAvailableNodes
	}

	availableNodes := make([]*balancerNode[T], nodeCount)
	copy(availableNodes, nodes)
	currentCount := nodeCount
	for currentCount > 0 {
		var selectedNode *balancerNode[T]
		var selectedIdx int

		if nodeCount == 1 {
			node := nodes[0]
			if node.max > 0 && node.total.Load() > node.max {
				return nil, ErrNoAvailableNodes
			}
			node.total.Add(1)
			return node, nil
		}
		switch strategy {
		case StrategyRandom:
			if allSameWeight := allNodesHaveSameWeight(availableNodes[:currentCount]); allSameWeight {
				selectedIdx = fastRand(currentCount)
				selectedNode = availableNodes[selectedIdx]
			} else {
				weighted := make(map[interface{}]uint32, currentCount)
				for _, node := range availableNodes[:currentCount] {
					weighted[node] = uint32(node.weight)
				}
				randNode, err := zstring.WeightedRand(weighted)
				if err != nil {
					return nil, err
				}
				selectedNode = randNode.(*balancerNode[T])
				for i, node := range availableNodes[:currentCount] {
					if node == selectedNode {
						selectedIdx = i
						break
					}
				}
			}

		case StrategyRoundRobin:
			idx := atomic.AddUint64(&b.lastNodeIdx, 1)
			selectedIdx = int(idx % uint64(currentCount))
			selectedNode = availableNodes[selectedIdx]

		case StrategyLeastConn:
			minScore := int64(math.MaxInt64)
			for i, node := range availableNodes[:currentCount] {
				curr := node.total.Load()
				score := curr * 100 / int64(node.weight)
				if score < minScore {
					minScore = score
					selectedIdx = i
				}
			}
			selectedNode = availableNodes[selectedIdx]
		}

		if selectedNode.max > 0 && selectedNode.total.Load() > selectedNode.max {
			availableNodes[selectedIdx] = availableNodes[currentCount-1]
			currentCount--
			continue
		}
		selectedNode.total.Add(1)
		return selectedNode, nil
	}

	return nil, ErrNoAvailableNodes
}

// allNodesHaveSameWeight checks if all nodes have the same weight
func allNodesHaveSameWeight[T any](nodes []*balancerNode[T]) bool {
	if len(nodes) <= 1 {
		return true
	}

	weight := nodes[0].weight
	for i := 1; i < len(nodes); i++ {
		if nodes[i].weight != weight {
			return false
		}
	}
	return true
}

var rngSeed uint32

func fastRand(n int) int {
	return int(atomic.AddUint32(&rngSeed, 1) % uint32(n))
}

// Run runs the given function on the selected node
func (b *Balancer[T]) Run(fn func(node T) (normal bool, err error), strategy ...BalancerStrategy) error {
	return b.RunByKeys(nil, fn, strategy...)
}

func (b *Balancer[T]) RunByKeys(keys []string, fn func(node T) (normal bool, err error), strategy ...BalancerStrategy) error {
	if fn == nil {
		return ErrEmptyCallback
	}

	r := b.mu.RLock()
	if len(b.nodes) == 0 {
		b.mu.RUnlock(r)
		return ErrNoNodesAdded
	}
	b.mu.RUnlock(r)

	var s BalancerStrategy
	if len(strategy) > 0 {
		s = strategy[0]
	}

	nodes := b.getAvailableNodes(keys...)
	if len(nodes) == 0 {
		return ErrNoAvailableNodes
	}

	err := zutil.DoRetry(len(nodes), func() error {
		if len(nodes) == 0 {
			return ErrNoAvailableNodes
		}

		node, err := b.selectNode(nodes, s)
		if err != nil {
			return err
		}

		if node.max > 0 && node.total.Load() > node.max {
			nodes = b.getAvailableNodes(keys...)
			return ErrNoAvailableNodes
		}

		err = zerror.TryCatch(func() error {
			normal, callErr := fn(node.node)

			if !normal {
				node.failedAt.Store(time.Now().UnixMilli())
			}

			return callErr
		})

		node.total.Add(-1)
		if err != nil {
			nodes = b.getAvailableNodes(keys...)
			return err
		}

		return nil
	}, func(rc *zutil.RetryConf) {
		// rc.BackOffDelay = true
		rc.Interval = time.Nanosecond
	})

	return err
}

// WalkNodes walks all nodes
func (b *Balancer[T]) WalkNodes(fn func(node T, available bool) (normal bool)) {
	keys := b.Keys()
	now := time.Now().UnixMilli()

	for _, key := range keys {
		r := b.mu.RLock()
		n, ok := b.nodes[key]
		if !ok {
			b.mu.RUnlock(r)
			continue
		}

		failedAt := n.failedAt.Load()
		available := !(failedAt > 0 && (now-failedAt) <= n.cooldown.Load())
		node := n.node
		b.mu.RUnlock(r)

		if !fn(node, available) {
			if available {
				if n, ok := b.nodes[key]; ok {
					n.failedAt.Store(now)
				}
			}
		} else if failedAt > 0 {
			if n, ok := b.nodes[key]; ok {
				n.failedAt.Store(0)
			}
		}
	}
}

// Keys returns all node keys
func (b *Balancer[T]) Keys() []string {
	r := b.mu.RLock()
	keys := make([]string, len(b.nodeKeys))
	copy(keys, b.nodeKeys)
	b.mu.RUnlock(r)
	return keys
}

// Len returns the number of nodes
func (b *Balancer[T]) Len() int {
	r := b.mu.RLock()
	defer b.mu.RUnlock(r)
	return len(b.nodes)
}
