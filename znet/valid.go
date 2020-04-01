package znet

import (
	"github.com/sohaha/zlsgo/zvalid"
)

func (c *Context) ValidRule() *zvalid.Engine {
	return zvalid.New()
}

// ValidParam get and verify routing parameters
func (c *Context) ValidParam(defRule *zvalid.Engine, key string, name ...string) *zvalid.Engine {
	return valid(defRule, c.GetParam(key), key, name...)
}

// ValidQuery get and verify query
func (c *Context) ValidQuery(defRule *zvalid.Engine, key string, name ...string) *zvalid.Engine {
	return valid(defRule, c.DefaultQuery(key, ""), key, name...)
}

// ValidForm get and verify postform
func (c *Context) ValidForm(defRule *zvalid.Engine, key string, name ...string) *zvalid.Engine {
	return valid(defRule, c.DefaultPostForm(key, ""), key, name...)
}

// ValidJSON get and verify json
func (c *Context) ValidJSON(defRule *zvalid.Engine, key string, name ...string) *zvalid.Engine {
	return valid(defRule, c.GetJSON(key).String(), key, name...)
}

// Valid Valid from -> query -> parame
func (c *Context) Valid(defRule *zvalid.Engine, key string, name ...string) *zvalid.Engine {
	value := c.DefaultFormOrQuery(key, "")
	if value == "" {
		value = c.GetParam(key)
	}
	return valid(defRule, value, key, name...)
}

func valid(defRule *zvalid.Engine, value, key string, name ...string) (valid *zvalid.Engine) {
	if defRule == nil {
		defRule = zvalid.New()
	}
	valid = defRule.Verifi(value, name...)
	return
}
