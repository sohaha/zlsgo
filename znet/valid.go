package znet

import (
	"github.com/sohaha/zlsgo/zvalid"
)

func (c *Context) ValidRule() *zvalid.Engine {
	return zvalid.New()
}

// ValidParam get and verify routing parameters
func (c *Context) ValidParam(rule *zvalid.Engine, key string, name ...string) *zvalid.Engine {
	return valid(rule, c.GetParam(key), key, name...)
}

// ValidQuery get and verify query
func (c *Context) ValidQuery(rule *zvalid.Engine, key string, name ...string) *zvalid.Engine {
	return valid(rule, c.DefaultQuery(key, ""), key, name...)
}

// ValidForm get and verify postform
func (c *Context) ValidForm(rule *zvalid.Engine, key string, name ...string) *zvalid.Engine {
	return valid(rule, c.DefaultPostForm(key, ""), key, name...)
}

// ValidJSON get and verify json
func (c *Context) ValidJSON(rule *zvalid.Engine, key string, name ...string) *zvalid.Engine {
	return valid(rule, c.GetJSON(key).String(), key, name...)
}

// Valid Valid from -> query -> parame
func (c *Context) Valid(rule *zvalid.Engine, key string, name ...string) *zvalid.Engine {
	value := c.DefaultFormOrQuery(key, "")
	if value == "" {
		value = c.GetParam(key)
	}
	return valid(rule, value, key, name...)
}

func valid(rule *zvalid.Engine, value, key string, name ...string) (valid *zvalid.Engine) {
	if rule == nil {
		rule = zvalid.New()
	}
	valid = rule.Verifi(value, name...)
	return
}
