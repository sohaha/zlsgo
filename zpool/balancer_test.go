//go:build go1.18
// +build go1.18

package zpool

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zsync"
	"github.com/sohaha/zlsgo/zutil"
)

func TestNewBalancer(t *testing.T) {
	tt := zlsgo.NewTest(t)
	b := NewBalancer[string]()

	err := b.Run(func(node string) (bool, error) {
		return true, nil
	}, 0)
	tt.NotNil(err)

	err = b.Run(nil, 0)
	tt.NotNil(err)

	b.Add("n1", "n1", func(opts *BalancerNodeOptions) {
		opts.MaxConns = 1
		opts.Weight = 10
	})
	b.Add("n2", "n2", func(opts *BalancerNodeOptions) {
		opts.MaxConns = 2
		opts.Cooldown = 5000
		opts.Weight = 5
	})
	b.Add("n5", "n5", func(opts *BalancerNodeOptions) {
		opts.MaxConns = 5
		opts.Cooldown = 5000
		opts.Weight = 50
	})

	n2, available, ok := b.Get("n2")
	tt.EqualTrue(ok)
	tt.EqualTrue(available)
	tt.Equal("n2", n2)

	tt.Run("limit", func(tt *zlsgo.TestUtil) {
		var wg zsync.WaitGroup
		success := zutil.NewInt32(0)
		for i := 0; i < 10; i++ {
			wg.Go(func() {
				err := b.Run(func(node string) (bool, error) {
					tt.Log(node)
					time.Sleep(time.Second / 10)
					return true, nil
				}, StrategyRoundRobin)
				if err == nil {
					success.Add(1)
				}
			})
		}
		wg.Wait()
		tt.Log(success.Load())
		tt.EqualTrue(success.Load() < 10)
	})

	tt.Run("weight", func(tt *zlsgo.TestUtil) {
		var wg zsync.WaitGroup
		success := zutil.NewInt32(0)
		for i := 0; i < 10; i++ {
			wg.Go(func() {
				err := b.Run(func(node string) (bool, error) {
					tt.Log(node)
					time.Sleep(time.Second / 10)
					return true, nil
				}, StrategyWeighted)
				if err == nil {
					success.Add(1)
				}
			})
		}
		wg.Wait()
		tt.Log(success.Load())
		tt.EqualTrue(success.Load() < 10)
	})

	tt.Run("least", func(tt *zlsgo.TestUtil) {
		var wg zsync.WaitGroup
		success := zutil.NewInt32(0)
		for i := 0; i < 10; i++ {
			wg.Go(func() {
				err := b.Run(func(node string) (bool, error) {
					tt.Log(node)
					time.Sleep(time.Second / 10)
					return true, nil
				}, StrategyLeastConn)
				if err == nil {
					success.Add(1)
				}
			})
		}
		wg.Wait()
		tt.Log(success.Load())
	})

	tt.Run("round", func(tt *zlsgo.TestUtil) {
		var wg zsync.WaitGroup
		success := zutil.NewInt32(0)
		nodes := []string{}
		for i := 0; i < 10; i++ {
			err = b.Run(func(node string) (bool, error) {
				tt.Log(node)
				nodes = append(nodes, node)
				return true, nil
			})
			if err == nil {
				success.Add(1)
			}
		}
		wg.Wait()
		tt.Log(success.Load())
	})

	tt.Run("byKeys", func(tt *zlsgo.TestUtil) {
		var wg zsync.WaitGroup
		success := zutil.NewInt32(0)
		for i := 0; i < 10; i++ {
			wg.Go(func() {
				err := b.RunByKeys([]string{"n1", "n2"}, func(node string) (bool, error) {
					tt.Log(node)
					return true, nil
				}, StrategyRoundRobin)
				if err == nil {
					success.Add(1)
				}
			})
		}
		wg.Wait()
		tt.Log(success.Load())
	})

	tt.Run("error", func(tt *zlsgo.TestUtil) {
		var wg zsync.WaitGroup
		success := zutil.NewInt32(0)
		for i := 0; i < 10; i++ {
			err = b.Run(func(node string) (bool, error) {
				tt.Log(i, node)
				if node == "n5" || node == "n2" {
					return false, errors.New("n error")
				}

				return true, nil
			}, StrategyRandom)
			if err == nil {
				success.Add(1)
			}
		}
		wg.Wait()
		tt.Log(success.Load())
	})

	tt.Run("keys", func(tt *zlsgo.TestUtil) {
		keys := b.Keys()
		tt.Equal(3, len(keys))
		for _, key := range keys {
			tt.EqualTrue(key == "n1" || key == "n2" || key == "n5")
			tt.Log(key)
			node, available, exists := b.Get(key)
			tt.EqualTrue(exists)
			tt.Log(node, available)
		}
	})

	tt.Run("walk", func(tt *zlsgo.TestUtil) {
		b := NewBalancer[string]()
		b.Add("n1", "n1")
		b.Add("n2", "n2")
		b.WalkNodes(func(node string, available bool) (normal bool) {
			tt.Log(node, available)
			if node == "n2" {
				return false
			}
			return true
		})

		tt.Equal(2, b.Len())

		n1, n1available, n1exists := b.Get("n1")
		tt.Equal("n1", n1)
		tt.EqualTrue(n1available)
		tt.EqualTrue(n1exists)

		n2, n2available, n2exists := b.Get("n2")
		tt.Equal("n2", n2)
		tt.EqualTrue(!n2available)
		tt.EqualTrue(n2exists)

		n3, n3available, n3exists := b.Get("n3")
		tt.Equal("", n3)
		tt.EqualTrue(!n3available)
		tt.EqualTrue(!n3exists)
	})

	tt.Run("remove", func(tt *zlsgo.TestUtil) {
		b := NewBalancer[string]()
		b.Add("n1", "n1")
		b.Add("n2", "n2")
		b.Add("n3", "n3")
		tt.Equal(3, b.Len())
		b.Remove("n1")
		b.Remove("n2")
		tt.Equal(1, b.Len())
		b.Remove("n3")
		b.Remove("n4")
		tt.Equal(0, b.Len())
	})

	tt.Run("mark", func(tt *zlsgo.TestUtil) {
		b := NewBalancer[string]()
		b.Add("n1", "n1")
		b.Add("n2", "n2")
		b.Add("n3", "n3")
		tt.Equal(3, b.Len())
		b.Mark("n1", false)
		b.WalkNodes(func(node string, available bool) (normal bool) {
			tt.Log(node, available)
			if node == "n1" {
				tt.Equal(false, available)
				return true
			}
			tt.Equal(true, available)
			return true
		})
	})

	tt.Run("weight management", func(tt *zlsgo.TestUtil) {
		b := NewBalancer[string]()

		b.Add("node1", "server1", func(opts *BalancerNodeOptions) {
			opts.Weight = 5
			opts.MaxConns = 10
			opts.Cooldown = 2000
		})
		b.Add("node2", "server2", func(opts *BalancerNodeOptions) {
			opts.Weight = 3
			opts.MaxConns = 5
			opts.Cooldown = 1000
		})

		weight1, exists1 := b.GetWeight("node1")
		tt.EqualTrue(exists1)
		tt.Equal(uint64(5), weight1)

		weight2, exists2 := b.GetWeight("node2")
		tt.EqualTrue(exists2)
		tt.Equal(uint64(3), weight2)

		weight3, exists3 := b.GetWeight("node3")
		tt.EqualTrue(!exists3)
		tt.Equal(uint64(0), weight3)

		err := b.SetWeight("node1", 8)
		tt.EqualNil(err)

		newWeight1, _ := b.GetWeight("node1")
		tt.Equal(uint64(8), newWeight1)

		err = b.SetWeight("node1", 0)
		tt.EqualTrue(err != nil)
		tt.Equal("invalid weight: must be between 1 and 1000000", err.Error())

		err = b.SetWeight("node1", 1000001)
		tt.EqualTrue(err != nil)
		tt.Equal("invalid weight: must be between 1 and 1000000", err.Error())

		err = b.SetWeight("node3", 5)
		tt.EqualTrue(err != nil)
		tt.Equal(ErrNodeNotFound, err)

		info1, exists1 := b.GetNodeInfo("node1")
		tt.EqualTrue(exists1)
		tt.Equal("server1", info1.Node)
		tt.Equal(uint64(8), info1.Weight)
		tt.Equal(int64(10), info1.MaxConns)
		tt.Equal(int64(2000), info1.Cooldown)
		tt.EqualTrue(info1.Available)
		tt.Equal(int64(0), info1.Active)

		_, exists3 = b.GetNodeInfo("node3")
		tt.EqualTrue(!exists3)

		b.Mark("node2", false)
		info2, exists2 := b.GetNodeInfo("node2")
		tt.EqualTrue(exists2)
		tt.Equal("server2", info2.Node)
		tt.Equal(uint64(3), info2.Weight)
		tt.EqualTrue(!info2.Available)
	})

	tt.Run("weight impact on load balancing", func(tt *zlsgo.TestUtil) {
		b := NewBalancer[string]()

		b.Add("node1", "server1", func(opts *BalancerNodeOptions) {
			opts.Weight = 1
		})
		b.Add("node2", "server2", func(opts *BalancerNodeOptions) {
			opts.Weight = 1
		})

		selections := make(map[string]int)

		for i := 0; i < 100; i++ {
			err := b.Run(func(node string) (bool, error) {
				selections[node]++
				return true, nil
			}, StrategyWeighted)
			tt.EqualNil(err)
		}

		tt.Log("Equal weights selections:", selections)

		selections = make(map[string]int)
		err := b.SetWeight("node1", 5)
		tt.NoError(err)
		err = b.SetWeight("node2", 1)
		tt.NoError(err)

		for i := 0; i < 120; i++ {
			err := b.Run(func(node string) (bool, error) {
				selections[node]++
				return true, nil
			}, StrategyWeighted)
			tt.EqualNil(err)
		}

		tt.Log("Different weights selections:", selections)

		totalSelections := selections["server1"] + selections["server2"]
		tt.Equal(120, totalSelections)

		node1Percentage := float64(selections["server1"]) / float64(totalSelections) * 100
		node2Percentage := float64(selections["server2"]) / float64(totalSelections) * 100
		tt.Logf("Node1 percentage: %.2f%%, Node2 percentage: %.2f%%", node1Percentage, node2Percentage)

		tt.EqualTrue(node1Percentage > 60) // At least 60%
		tt.EqualTrue(node2Percentage < 40) // At most 40%
	})

	tt.Run("concurrent weight management", func(tt *zlsgo.TestUtil) {
		b := NewBalancer[string]()

		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("node%d", i)
			b.Add(key, "server"+key, func(opts *BalancerNodeOptions) {
				opts.Weight = 1
			})
		}

		var wg zsync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Go(func() {
				for j := 0; j < 10; j++ {
					key := fmt.Sprintf("node%d", j%5)
					weight := uint64((j % 5) + 1)
					err := b.SetWeight(key, weight)
					if err != nil {
						tt.Log("SetWeight error:", err)
					}
				}
			})
		}

		for i := 0; i < 10; i++ {
			wg.Go(func() {
				for j := 0; j < 10; j++ {
					key := fmt.Sprintf("node%d", j%5)
					weight, exists := b.GetWeight(key)
					if exists {
						tt.Log("GetWeight", key, ":", weight)
					}
				}
			})
		}

		for i := 0; i < 5; i++ {
			wg.Go(func() {
				for j := 0; j < 10; j++ {
					key := fmt.Sprintf("node%d", j%5)
					info, exists := b.GetNodeInfo(key)
					if exists {
						tt.Log("GetNodeInfo", key, ":", info.Weight)
					}
				}
			})
		}

		wg.Wait()

		for i := 0; i < 5; i++ {
			key := fmt.Sprintf("node%d", i)
			weight, exists := b.GetWeight(key)
			tt.EqualTrue(exists)
			tt.EqualTrue(weight > 0)
		}
	})

	tt.Run("edge cases and error handling", func(tt *zlsgo.TestUtil) {
		b := NewBalancer[string]()

		_, exists := b.GetWeight("nonexistent")
		tt.EqualTrue(!exists)

		_, exists = b.GetNodeInfo("nonexistent")
		tt.EqualTrue(!exists)

		err := b.SetWeight("nonexistent", 5)
		tt.EqualTrue(err != nil)
		tt.Equal(ErrNodeNotFound, err)

		b.Add("test", "server", func(opts *BalancerNodeOptions) {
			opts.Weight = 1
		})

		maxWeight := uint64(1000000)
		err = b.SetWeight("test", maxWeight)
		tt.EqualNil(err)

		weight, _ := b.GetWeight("test")
		tt.Equal(maxWeight, weight)

		err = b.SetWeight("test", 1)
		tt.EqualNil(err)
	})
}
