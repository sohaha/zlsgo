package znet

import (
	"reflect"

	"github.com/sohaha/zlsgo/zdi"
)

func handlerFuncs(h []Handler) (middleware []handlerFn, firstMiddleware []handlerFn) {
	middleware = make([]handlerFn, 0, len(h))
	firstMiddleware = make([]handlerFn, 0, len(h))
	for i := range h {
		fn := h[i]
		if v, ok := fn.(firstHandler); ok {
			firstMiddleware = append(firstMiddleware, handlerFunc(v[0]))
		} else {
			middleware = append(middleware, handlerFunc(fn))
		}
	}
	return
}

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
		case []byte:
			c.render = &renderByte{Data: vv}
		case ApiData:
			c.render = &renderJSON{Data: vv}
		}
	}
	return
}

func handlerFunc(h Handler) (fn handlerFn) {
	if h == nil {
		return nil
	}

	switch v := h.(type) {
	case func(*Context):
		return func(c *Context) error {
			v(c)
			return nil
		}
	case func(*Context) (interface{}, error):
		return func(c *Context) error {
			res, err := v(c)
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
	case handlerFn:
		return v
	case zdi.PreInvoker:
		return func(c *Context) error {
			v, err := c.injector.Invoke(v)
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
			c.String(int32(v[0].Int()), v[1].String())
			return err
		}
	default:
		return func(c *Context) error {
			v, err := c.injector.Invoke(v)
			if err != nil {
				return err
			}

			return invokeHandler(c, v)
		}
	}
}
