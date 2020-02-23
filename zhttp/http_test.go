package zhttp

import (
	"fmt"
	zls "github.com/sohaha/zlsgo"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestHttp(T *testing.T) {
	t := zls.NewTest(T)
	var (
		res          *Res
		err          error
		data         string
		expectedText string
	)

	forMethod(t)

	// test post
	expectedText = "ok"
	res, err = newMethod("Post", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		id, ok := r.PostForm["id"]
		if !ok {
			t.T.Fatal("err")
		}
		_, _ = w.Write([]byte(expectedText + id[0]))
	}, Param{
		"id": "123",
	})
	if err != nil {
		t.T.Fatal(err)
	}
	data = res.String()
	t.Equal(expectedText+"123", data)

	// test post application/x-www-form-urlencoded
	expectedText = "ok"
	res, err = newMethod("Post", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		id, ok := query["id"]
		if !ok {
			t.T.Fatal("err")
		}
		_, _ = w.Write([]byte(expectedText + id[0]))
	}, QueryParam{
		"id": "123",
	})
	if err != nil {
		t.T.Fatal(err)
	}
	data = res.String()
	t.Equal(expectedText+"123", data)
	t.Log(res.GetCookie())
}

func TestJONS(tt *testing.T) {
	t := zls.NewTest(tt)
	jsonData := `{"name":"is json"}`
	v := BodyJSON(jsonData)
	_, _ = newMethod("POST", func(w http.ResponseWriter, r *http.Request) {
		tt.Log(v)
		body, err := ioutil.ReadAll(r.Body)
		t.EqualExit(nil, err)
		t.EqualExit(jsonData, string(body))
		t.EqualExit("application/json; charset=UTF-8", r.Header.Get("Content-Type"))
	}, v, Header{"name": "ok"})
}

func TestGetMethod(tt *testing.T) {
	t := zls.NewTest(tt)
	jsonData := struct {
		Code int `json:"code"`
	}{}
	data := ""
	values := [...]string{
		"text",
		"{\"code\":201}",
	}
	EnableCookie(false)
	for i, v := range values {
		cookie := &http.Cookie{
			Name:     "c",
			Value:    "ok" + fmt.Sprint(i),
			Path:     "/",
			HttpOnly: true,
			MaxAge:   0,
		}
		res, err := newMethod("GET", func(w http.ResponseWriter, _ *http.Request) {
			tt.Log(v)
			w.Header().Add("Set-Cookie", cookie.String())
			_, _ = w.Write([]byte(v))
		}, v)
		tt.Log("get ok", i, err)
		t.Equal(nil, err)
		if err = res.ToJSON(&jsonData); err == nil {
			t.Equal(201, jsonData.Code)
		}
		if data, err = res.ToString(); err == nil {
			t.Equal(v, data)
		}
		t.Equal("GET", res.Request().Method)
		tt.Log(res.GetCookie())
		tt.Log(res.String(), "\n")
	}
	EnableCookie(true)
}

func forMethod(t *zls.TestUtil) {
	values := [...]string{"Get", "Put", "Head", "Options", "Delete", "Patch"}
	for _, v := range values {
		_, err := newMethod(v, func(_ http.ResponseWriter, _ *http.Request) {
		})
		if err != nil {
			t.T.Fatal(v, err)
		}
	}
}

func newMethod(method string, handler func(_ http.ResponseWriter, _ *http.Request), param ...interface{}) (res *Res, err error) {
	ts := httptest.NewServer(http.HandlerFunc(handler))
	curl := ts.URL
	switch method {
	case "Get":
		std.debug = true
		res, err = Get(curl, param...)
		std.debug = false
	case "Post":
		res, err = Post(curl, param...)
	case "Put":
		res, err = Put(curl, param...)
	case "Head":
		res, err = Head(curl, param...)
	case "Options":
		res, err = Options(curl, param...)
	case "Delete":
		res, err = Delete(curl, param...)
	case "Patch":
		res, err = Patch(curl, param...)
	default:
		method = strings.Title(method)
		res, err = Do(method, curl, param...)
		if err == nil {
			fmt.Println(res.Dump())
		}
	}

	return
}

func TestRes(t *testing.T) {
	tt := zls.NewTest(t)
	// res, err := Get("https://www.npmjs.com/package/zls-vue-spa/")
	res, err := Get("http://baidu.com")
	tt.Equal(nil, err)
	// t.Log(res.Body())
	// respBody, err := ioutil.ReadAll(res.Body())
	// t.Log(res.Body())
	// t.Log(string(respBody))
	t.Log(res.Body())
	t.Log(res.String())
	t.Log(res.Body())
	respBody, _ := ioutil.ReadAll(res.Body())
	t.Log(string(respBody))
}
func TestHttpProxy(T *testing.T) {
	t := zls.NewTest(T)
	err := SetProxy(func(r *http.Request) (*url.URL, error) {
		if strings.Contains(r.URL.String(), "qq.com") {
			t.Log(r.URL.String(), "SetProxy get", "http://127.0.0.1:6666")
			return url.Parse("http://127.0.0.1:6666")
		} else {
			t.Log(r.URL.String(), "Not SetProxy")
		}
		return nil, nil
	})
	var res *Res
	if err != nil {
		t.T.Fatal(err)
	}

	SetTimeout(10 * time.Second)

	res, err = Get("http://www.qq.com")
	if err == nil {
		t.Log(res.Response().Status)
	} else {
		t.Log(err)
	}
	t.Equal(true, err != nil)

	res, err = Get("https://www.npmjs.com/package/zls-vue-spa/")
	if err == nil {
		t.Log(res.Response().Status)
	} else {
		t.Log(err)
	}
	t.Equal(false, err != nil)
}

func TestHttpProxyUrl(T *testing.T) {
	err := SetProxyUrl("http://127.0.0.1:6666")
	t := zls.NewTest(T)
	if err != nil {
		t.T.Fatal(err)
	}

	SetTimeout(1 * time.Second)
	_, err = newMethod("GET", func(w http.ResponseWriter, _ *http.Request) {
	})
	t.Equal(true, err != nil)
}
