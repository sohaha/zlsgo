package znet

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/ztype"
)

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
		default:
			if vv != nil {
				c.render = &renderJSON{Data: ApiData{Data: v}}
			}
		}
	}
	return
}

func (utils) ParseHandlerFunc(h Handler) (fn handlerFn) {
	if h == nil {
		return func(c *Context) error {
			return errors.New("Handler is nil")
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
	case func(*Context) (ztype.Map, error):
		return func(c *Context) error {
			res, err := v(c)
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
		val := reflect.ValueOf(v)
		if val.Kind() != reflect.Func {
			return func(c *Context) error {
				c.Byte(http.StatusOK, ztype.ToBytes(v))
				return nil
			}
			// panic("znet Handler is not a function: " + val.Kind().String())
		}

		return func(c *Context) error {
			v, err := c.injector.Invoke(v)
			if err != nil {
				return err
			}

			return invokeHandler(c, v)
		}
	}
}
