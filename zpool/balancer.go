//go:build go1.18
// +build go1.18

package zpool

import (
	"errors"
	"math"
	"sync/atomic"
	"time"

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
	// StrategyRoundRobin represents round-robin selection strategy
	StrategyRoundRobin
	// StrategyLeastConn represents least connections selection strategy
	StrategyLeastConn
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
	// Cooldown period after node failure, default 1000ms
	Cooldown int64
}

// NewBalancer creates a new Balancer
func NewBalancer[T any]() *Balancer[T] {
	b := &Balancer[T]{
		nodes: make(map[string]*balancerNode[T]),
		mu:    zsync.NewRBMutex(),
	}

	return b
}

// Get returns the node with the given key
func (b *Balancer[T]) Get(key string) *balancerNode[T] {
	r := b.mu.RLock()
	defer b.mu.RUnlock(r)

	return b.nodes[key]
}

// Add adds a new node to the balancer
func (b *Balancer[T]) Add(key string, node T, opt ...func(opts *BalancerNodeOptions)) error {
	if key == "" {
		return errors.New("key is required")
	}
	if _, ok := b.nodes[key]; ok {
		return errors.New("node already exists")
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
	b.nodes[key] = n
	b.nodeKeys = append(b.nodeKeys, key)
	b.mu.Unlock()

	return nil
}

// Remove removes a node from the balancer
func (b *Balancer[T]) Remove(key string) {
	b.mu.Lock()
	if _, exists := b.nodes[key]; !exists {
		b.mu.Unlock()
		return
	}

	delete(b.nodes, key)
	n := 0
	for _, k := range b.nodeKeys {
		if k != key {
			b.nodeKeys[n] = k
			n++
		}
	}
	b.nodeKeys = b.nodeKeys[:n]
	b.mu.Unlock()
}

// getAvailableNodes returns all nodes that are available to use
func (b *Balancer[T]) getAvailableNodes() []*balancerNode[T] {
	r := b.mu.RLock()
	nodes := make([]*balancerNode[T], 0, len(b.nodes))
	now := time.Now().UnixMilli()
	for _, key := range b.nodeKeys {
		node := b.nodes[key]
		if node.failedAt.Load() > 0 && (now-node.failedAt.Load()) <= node.cooldown.Load() {
			continue
		}

		if node.max <= 0 || node.total.Load() < node.max {
			nodes = append(nodes, node)
		}
	}
	b.mu.RUnlock(r)

	return nodes
}

// selectNode selects a node based on the given strategy
func (b *Balancer[T]) selectNode(nodes []*balancerNode[T], strategy BalancerStrategy) (*balancerNode[T], error) {
	if len(nodes) == 0 {
		return nil, errors.New("no available nodes")
	}

	var selectedNode *balancerNode[T]

	switch strategy {
	case StrategyRandom:
		weighted := map[interface{}]uint32{}
		for _, node := range nodes {
			weighted[node] = uint32(node.weight)
		}
		randNode, err := zstring.WeightedRand(weighted)
		if err != nil {
			return nil, err
		}
		selectedNode = randNode.(*balancerNode[T])

	case StrategyRoundRobin:
		idx := atomic.AddUint64(&b.lastNodeIdx, 1)
		selectedNode = nodes[idx%uint64(len(nodes))]

	case StrategyLeastConn:
		minScore := int64(math.MaxInt64)
		minIdx := 0
		for i, node := range nodes {
			curr := node.total.Load()
			score := curr * 100 / int64(node.weight)
			if score < minScore {
				minScore = score
				minIdx = i
			}
		}
		selectedNode = nodes[minIdx]
	}

	if selectedNode.max > 0 {
		if !selectedNode.total.CAS(selectedNode.total.Load(), selectedNode.total.Load()+1) {
			return nil, errors.New("node connection limit reached")
		}
	} else {
		selectedNode.total.Add(1)
	}

	return selectedNode, nil
}

// Run runs the given function on the selected node
func (b *Balancer[T]) Run(fn func(node T) (normal bool, err error), strategy ...BalancerStrategy) error {
	if fn == nil {
		return errors.New("callback function cannot be empty")
	}

	r := b.mu.RLock()
	if len(b.nodes) == 0 {
		b.mu.RUnlock(r)
		return errors.New("please add nodes first")
	}

	b.mu.RUnlock(r)

	s := StrategyRoundRobin
	if len(strategy) > 0 {
		s = strategy[0]
	}

	nodes := b.getAvailableNodes()

	if len(nodes) == 0 {
		return errors.New("no available nodes")
	}

	err := zutil.DoRetry(len(nodes)-1, func() error {
		node, err := b.selectNode(nodes, s)
		if err != nil {
			return err
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
			nodes = b.getAvailableNodes()
			return err
		}

		return nil
	}, func(rc *zutil.RetryConf) {
		rc.BackOffDelay = true
	})

	return err
}
