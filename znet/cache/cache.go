package cache

import (
	"sort"
	"time"

	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zstring"
)

type (
	// Config configuration
	Config struct {
		Custom func(c *znet.Context) (key string, expiration time.Duration)
		zcache.Option
	}
	cacheContext struct {
		Type    string
		Content []byte
		Code    int32
	}
)

func New(opt ...func(conf *Config)) znet.HandlerFunc {
	conf := Config{
		Custom: func(c *znet.Context) (key string, expiration time.Duration) {
			return QueryKey(c), 0
		},
	}

	cache := zcache.NewFast(func(o *zcache.Option) {
		conf.Option = *o
		conf.Option.Expiration = time.Minute * 10
		for _, f := range opt {
			f(&conf)
		}
		*o = conf.Option
	})

	return func(c *znet.Context) {
		key, expiration := conf.Custom(c)
		if key == "" {
			c.Next()
			return
		}

		v, ok := cache.ProvideGet(key, func() (interface{}, bool) {
			c.Next()

			p := c.PrevContent()
			if p.Code.Load() != 0 {
				return &cacheContext{
					Code:    p.Code.Load(),
					Type:    p.Type,
					Content: p.Content,
				}, true
			}
			return nil, false
		}, expiration)

		if !ok {
			return
		}

		if data, ok := v.(*cacheContext); ok {
			p := c.PrevContent()
			p.Code.Store(data.Code)
			p.Content = data.Content
			p.Type = data.Type
			c.Abort()
		}
	}
}

func QueryKey(c *znet.Context) (key string) {
	m := c.GetAllQueryMaps()
	mLen := len(m)
	if mLen == 0 {
		return c.Request.URL.Path
	}

	keys := make([]string, 0, mLen)
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	b := zstring.Buffer((len(m) * 4) + 2)
	b.WriteString(c.Request.URL.Path)
	b.WriteString("?")
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(m[k])
		b.WriteString("&")
	}
	return b.String()
}
