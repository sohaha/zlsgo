package znet

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/ztype"
)

// handlerFuncs separates regular handlers from first handlers.
// First handlers are executed before regular middleware in the request pipeline.
// This function returns two slices: regular middleware handlers and first middleware handlers.
func handlerFuncs(h []Handler) (middleware []handlerFn, firstMiddleware []handlerFn) {
	middleware = make([]handlerFn, 0, len(h))
	firstMiddleware = make([]handlerFn, 0, len(h))
	for i := range h {
		fn := h[i]
		if v, ok := fn.(firstHandler); ok {
			firstMiddleware = append(firstMiddleware, Utils.ParseHandlerFunc(v[0]))
		} else {
			middleware = append(middleware, Utils.ParseHandlerFunc(fn))
		}
	}
	return
}

// invokeHandler processes the return values from handler functions and updates the context accordingly.
// It handles various return types including status codes, strings, errors, renderers, and custom types.
// This function is used internally by the dependency injection system.
func invokeHandler(c *Context, v []reflect.Value) (err error) {
	for i := range v {
		value := v[i]
		v := value.Interface()
		switch vv := v.(type) {
		case int:
			c.prevData.Code.Store(int32(vv))
		case int32:
			c.prevData.Code.Store(vv)
		case uint:
			c.prevData.Code.Store(int32(vv))
		case string:
			c.render = &renderString{Format: vv}
		case error:
			err = vv
		case Renderer:
			c.render = vv
		case []byte:
			c.render = &renderByte{Data: vv}
		case ApiData:
			c.render = &renderJSON{Data: vv}
		default:
			if vv != nil {
				c.render = &renderJSON{Data: ApiData{Data: v}}
			}
		}
	}
	return
}

// ParseHandlerFunc converts various handler function signatures to the internal handlerFn type.
// It supports multiple function signatures including standard handlers, dependency-injected handlers,
// and functions returning various combinations of values and errors.
func (utils) ParseHandlerFunc(h Handler) (fn handlerFn) {
	if h == nil {
		return func(c *Context) error {
			return errors.New("handler is nil")
		}
	}

	switch v := h.(type) {
	case HandlerFunc:
		return func(c *Context) error {
			v(c)
			return nil
		}
	case func(*Context):
		return func(c *Context) error {
			v(c)
			return nil
		}
	case func(c *Context) error:
		return v
	case func(*Context) (interface{}, error):
		return func(c *Context) error {
			res, err := v(c)

			if c.stopHandle.Load() {
				return nil
			}

			if err != nil {
				return err
			}

			if res == nil {
				res = struct{}{}
			}
			data := ApiData{
				Data: res,
			}
			c.JSON(200, data)
			return nil
		}
	case func(*Context) (ztype.Map, error):
		return func(c *Context) error {
			res, err := v(c)

			if c.stopHandle.Load() {
				return nil
			}

			if err != nil {
				return err
			}
			data := ApiData{
				Data: res,
			}
			c.JSON(200, data)
			return nil
		}
	case handlerFn:
		return v
	case zdi.PreInvoker:
		return func(c *Context) error {
			v, err := c.injector.Invoke(v)

			if c.stopHandle.Load() {
				return nil
			}

			if err != nil {
				return err
			}

			if len(v) == 0 {
				return nil
			}
			return invokeHandler(c, v)
		}
	case func() (int, string):
		return func(c *Context) error {
			v, err := c.injector.Invoke(invokerCodeText(v))

			if c.stopHandle.Load() {
				return nil
			}

			c.String(int32(v[0].Int()), v[1].String())
			return err
		}
	default:
		val := zreflect.ValueOf(v)

		var fn interface{}
		for i := range preInvokers {
			if val.Type().ConvertibleTo(preInvokers[i]) {
				fn = val.Convert(preInvokers[i]).Interface()
				break
			}
		}

		if fn == nil {
			if val.Kind() != reflect.Func {
				b := ztype.ToBytes(v)
				isJSON := zjson.ValidBytes(b)
				return func(c *Context) error {
					c.Byte(http.StatusOK, b)
					if isJSON {
						c.SetContentType(ContentTypeJSON)
					}
					return nil
				}
				// panic("znet Handler is not a function: " + val.Kind().String())
			}
			fn = v
		}

		return func(c *Context) error {
			v, err := c.injector.Invoke(fn)

			if c.stopHandle.Load() {
				return nil
			}

			if err != nil {
				return err
			}

			return invokeHandler(c, v)
		}
	}
}
