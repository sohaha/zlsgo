package znet

import (
	"github.com/sohaha/zlsgo/zvalid"
)

// ValidRule creates and returns a new validation engine.
// This is a convenience method to start building validation rules for request data.
func (c *Context) ValidRule() zvalid.Engine {
	return zvalid.New()
}

// ValidParam gets and validates route parameters.
// It retrieves the parameter value by key and applies the provided validation rules.
// Optional name parameter can be provided for error messages.
func (c *Context) ValidParam(defRule zvalid.Engine, key string, name ...string) zvalid.Engine {
	return valid(defRule, c.GetParam(key), key, name...)
}

// ValidQuery gets and validates URL query parameters.
// It retrieves the query parameter value by key and applies the provided validation rules.
// Optional name parameter can be provided for error messages.
func (c *Context) ValidQuery(defRule zvalid.Engine, key string, name ...string) zvalid.Engine {
	return valid(defRule, c.DefaultQuery(key, ""), key, name...)
}

// ValidForm gets and validates form data.
// It retrieves the form field value by key and applies the provided validation rules.
// Optional name parameter can be provided for error messages.
func (c *Context) ValidForm(defRule zvalid.Engine, key string, name ...string) zvalid.Engine {
	return valid(defRule, c.DefaultPostForm(key, ""), key, name...)
}

// ValidJSON gets and validates JSON data.
// It retrieves the JSON value by key path and applies the provided validation rules.
// Optional name parameter can be provided for error messages.
func (c *Context) ValidJSON(defRule zvalid.Engine, key string, name ...string) zvalid.Engine {
	return valid(defRule, c.GetJSON(key).String(), key, name...)
}

// Valid validates data from multiple sources in order: JSON/form data -> query parameters -> route parameters.
// It tries to find the value in each source and applies the provided validation rules to the first non-empty value found.
// Optional name parameter can be provided for error messages.
func (c *Context) Valid(defRule zvalid.Engine, key string, name ...string) zvalid.Engine {
	value, contentType := "", c.ContentType()
	if contentType == c.ContentType(ContentTypeJSON) || contentType == c.ContentType(ContentTypePlain) {
		value = c.GetJSON(key).String()
	}
	if value == "" {
		value = c.DefaultFormOrQuery(key, "")
	}
	if value == "" {
		value = c.GetParam(key)
	}
	return valid(defRule, value, key, name...)
}

// valid is an internal helper function that applies validation rules to a value.
// It returns the validation engine with the result of the validation.
func valid(defRule zvalid.Engine, value, _ string, name ...string) (valid zvalid.Engine) {
	return defRule.Verifi(value, name...)
}
