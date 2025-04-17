//go:build go1.18
// +build go1.18

package zpool

import (
	"errors"
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
}
