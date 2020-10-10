package zhttp

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/znet"
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
	urlValues := url.Values{"ok": []string{"666"}}
	res, err = newMethod("Post", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		id, ok := r.PostForm["id"]
		if !ok {
			t.T.Fatal("err")
		}
		_, _ = w.Write([]byte(expectedText + id[0]))
	}, urlValues, Param{
		"id":  "123",
		"id2": "123",
	}, QueryParam{
		"id3": 333,
		"id6": 666,
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
	values := [...]string{"Get", "Put", "Head", "Options", "Delete", "Patch", "Trace", "Connect"}
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
		res, err = Get(curl, param...)
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
	case "Connect":
		res, err = Connect(curl, param...)
	case "Trace":
		res, err = Trace(curl, param...)
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
	u := "https://cdn.jsdelivr.net/"
	// res, err := Get("https://www.npmjs.com/package/zls-vue-spa/")
	res, err := Get(u)
	t.Log(u, err)
	tt.EqualExit(true, err == nil)
	t.Log(res.Body())
	t.Log(res.String())
	t.Log(res.Body())
	respBody, _ := ioutil.ReadAll(res.Body())
	t.Log(string(respBody))
	t.Log(res.Dump())
}

func TestHttpProxy(t *testing.T) {
	tt := zls.NewTest(t)
	err := SetProxy(func(r *http.Request) (*url.URL, error) {
		if strings.Contains(r.URL.String(), "qq.com") {
			tt.Log(r.URL.String(), "SetProxy get", "http://127.0.0.1:6666")
			return url.Parse("http://127.0.0.1:6666")
		} else {
			tt.Log(r.URL.String(), "Not SetProxy")
		}
		return nil, nil
	})
	var res *Res
	if err != nil {
		tt.T.Fatal(err)
	}

	SetTimeout(10 * time.Second)

	res, err = Get("http://www.qq.com")
	if err == nil {
		tt.Log(res.Response().Status)
	} else {
		tt.Log(err)
	}
	tt.Equal(true, err != nil)

	res, err = Get("https://cdn.jsdelivr.net/npm/zls-vue-spa@1.1.29/package.json")
	if err == nil {
		tt.Log(res.Response().Status)
	} else {
		tt.Log(err)
	}
	tt.Equal(false, err != nil)
}

func TestHttpProxyUrl(t *testing.T) {
	err := SetProxyUrl("http://127.0.0.1:6666", "http://127.0.0.1:7777")
	tt := zls.NewTest(t)
	if err != nil {
		tt.T.Fatal(err)
	}

	SetTimeout(1 * time.Second)
	_, err = newMethod("GET", func(w http.ResponseWriter, _ *http.Request) {
	})
	tt.Equal(true, err != nil)
}

func TestFile(t *testing.T) {
	tt := zls.NewTest(t)
	_ = RemoveProxy()
	SetTimeout(20 * time.Second)
	downloadProgress := func(current, total int64) {
		t.Log("downloadProgress", current, total)
	}
	res, err := Get("https://cdn.jsdelivr.net/gh/sohaha/uniapp-template/src/static/my.jpg", downloadProgress)
	tt.EqualNil(err)
	if err == nil {
		err = res.ToFile("./my.jpg")
		tt.EqualNil(err)
	}
	defer zfile.Rmdir("./my.jpg")
	r := znet.New()
	r.POST("/upload", func(c *znet.Context) {
		file, err := c.FormFile("file")
		t.Log(err, c.Host(true))
		t.Log(c.GetPostFormAll())
		tt.EqualExit("upload", c.GetHeader("type"))
		if err == nil {
			err = c.SaveUploadedFile(file, "./my2.jpg")
			tt.EqualNil(err)
			c.String(200, "上传成功")
		}
	})
	r.SetAddr("7878")
	go func() {
		znet.Run()
	}()

	std.CheckRedirect()
	time.Sleep(time.Second)

	v := url.Values{
		"name": []string{"isTest"},
	}
	q := Param{"q": "yes"}

	h := Header{
		"type": "upload",
	}
	res, err = Post("http://127.0.0.1:7878/upload", h, UploadProgress(func(current, total int64) {
		t.Log(current, total)
	}), Host("http://127.0.0.1:7878"), v, q, File("my.jpg", "file"))
	if err != nil {
		tt.EqualNil(err)
		return
	}
	tt.Equal("上传成功", res.String())
	zfile.Rmdir("./my2.jpg")

	DisableChunke()
	res, err = Post("http://127.0.0.1:7878/upload", h, UploadProgress(func(current, total int64) {
		t.Log(current, total)
	}), v, q, context.Background(), File("my.jpg", "file"))
	tt.EqualNil(err)
	tt.Equal("上传成功", res.String())
	zfile.Rmdir("./my2.jpg")
}

func TestRandomUserAgent(T *testing.T) {
	tt := zls.NewTest(T)
	for i := 0; i < 10; i++ {
		tt.Log(RandomUserAgent())
	}
	SetUserAgent(func() string {
		return ""
	})
}

func TestGetCode(t *testing.T) {
	tt := zls.NewTest(t)
	EnableInsecureTLS(true)
	r, _ := Get("https://xxxaaa--xxx.jsdelivr.net/")
	tt.EqualExit(0, r.StatusCode())

	c := newClient()
	SetClient(c)
	r, err := Get("https://cdn.jsdelivr.net/gh/sohaha/uniapp-template@master/README.md")
	if err != nil {
		t.Fatal(err)
	}
	tt.EqualExit(200, r.StatusCode())
	t.Log(r.String())
	t.Log(r.StatusCode())
	r.Dump()
}
