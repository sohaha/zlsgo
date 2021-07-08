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

func TestRetry(tt *testing.T) {
	t := zls.NewTest(tt)
	h := New()
	i := 0
	_, err := DoRetry(30, time.Second/10, func() (*Res, error) {
		t := time.Duration(100*2*i+1) * time.Millisecond
		h.SetTimeout(t)
		tt.Log("Retry", i, t)
		res, err := h.Get("https://cdn.jsdelivr.net/")
		i++

		return res, err
	})

	t.Log(i, err)
}

func TestToHTML(tt *testing.T) {
	t := zls.NewTest(tt)
	res, err := Get("https://cdn.jsdelivr.net/")
	if err != nil {
		t.Fatal(err)
		return
	}
	t.EqualTrue(res.HTML().Select("title").Text() != "")
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
		t.EqualExit(nil, err)
		tt.Log(res.String())
		if err = res.ToJSON(&jsonData); err == nil {
			t.Equal(201, jsonData.Code)
		}

		j := res.JSONs()
		tt.Log(j.String())
		if j.IsObject() {
			t.EqualExit(201, j.Get("code").Int())
			t.EqualExit(201, res.JSON("code").Int())
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

func TestHttpProxyUrl(tt *testing.T) {
	t := zls.NewTest(tt)
	_ = SetTransport(func(transport *http.Transport) {})
	err := SetProxyUrl()
	t.EqualTrue(err != nil)
	err = SetProxyUrl("http://127.0.0.1:66661", "http://127.0.0.1:77771")
	t.EqualNil(err)

	SetTimeout(1 * time.Second)
	_, err = newMethod("GET", func(w http.ResponseWriter, _ *http.Request) {
	})
	tt.Log(err)
	t.Equal(true, err != nil)
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
		err = res.ToFile(`../zhttp\test\my.jpg`)
		tt.EqualNil(err)
		t.Log(len(res.String()))
		err = res.ToFile(`../zhttp\test\my2.jpg`)
		tt.Log(err)
		tt.EqualNil(err)
	}
	defer zfile.Rmdir("./test/")
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
	}), Host("http://127.0.0.1:7878"), v, q, File("test\\my.jpg", "file"))
	if err != nil {
		tt.EqualNil(err)
		return
	}
	tt.Equal("上传成功", res.String())
	zfile.Rmdir("./my2.jpg")

	SetTransport(func(transport *http.Transport) {
		transport.MaxIdleConnsPerHost = 100
	})

	DisableChunke()
	res, err = Post("http://127.0.0.1:7878/upload", h, CustomReq(func(req *http.Request) {

	}), UploadProgress(func(current, total int64) {
		t.Log(current, total)
	}), v, q, context.Background(), File("./test//my.jpg", "file"))
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

func TestConvertCookie(tt *testing.T) {
	t := zls.NewTest(tt)
	cookie := ConvertCookie(
		" langx=zh-cn; lang code= zh-cn; sid=3c14598d6f2bce696a73a7649ab3df0df23c13c1; ")
	for _, c := range cookie {
		switch c.Name {
		case "langx":
			t.Equal("zh-cn", c.Value)
		case "lang code":
			t.Equal(" zh-cn", c.Value)
		case "sid":
			t.Equal("3c14598d6f2bce696a73a7649ab3df0df23c13c1", c.Value)
		default:
			tt.Fatal("no match", c.Name, c.Value)
		}
	}
}

func TestTlsCertificate(tt *testing.T) {
	t := zls.NewTest(tt)

	key := "localhost.key"
	crt := "localhost.crt"
	defer func() {
		zfile.Rmdir(key)
		zfile.Rmdir(crt)
	}()
	_ = zfile.WriteFile(crt, []byte("-----BEGIN CERTIFICATE-----\nMIIFODCCAyCgAwIBAgIUSwmVV6hatwktLUBtLCdTw0rZ5+UwDQYJKoZIhvcNAQEL\nBQAwFDESMBAGA1UEAwwJbG9jYWxob3N0MB4XDTIxMDMyOTA2NTc0M1oXDTMxMDMy\nNzA2NTc0M1owFDESMBAGA1UEAwwJbG9jYWxob3N0MIICIjANBgkqhkiG9w0BAQEF\nAAOCAg8AMIICCgKCAgEAv9FDHTRyLaNBwr7IzdvVHEBHEWsGvYoGiBDZBwnYlyDo\nl1O+zs9HdvUoMx9yJinojxY8nok0yN+2uMVTke+Si1h/Qh1dELmI9qKenOrCVtoc\nxz7KE3TVUua+Bnezx199PmIf35ZYp5jXCU4ceHA0hNXL63qedqVlDVVCl/cHgFMK\n2+dyRF0SjwGsLIguWQPHAB5+N1HbSU0QsJztL4swFr87Fm2k96Q1od3pJHiwBYX0\ntP2mzbIzRpRRyWZ8r57T+ECAGX01A2xU5IVC6gXlWHZTOe//1qf89Xf3RIk2RZiv\noSZg8UG+3q3J+npV+nvlzcS7LPhbCmne5uGTKI96tPnmke2cv00T5q7+T3EDtdhd\nOIvb3s8nM+ggih/W2PN7p8V+hvIH0BlGagzitKa59ZAiwR6zpq1IrgAcIFflE01j\nrGxftIpIMmPtB4uD++vaHxZ96BZvVucTTo3pRxZuQ7ylMyh7ZHHAVNJWrVtSk02s\nvDIju43SC1UT9p2vKtuZf9rEnHy34luzIJGKmVXBKF8FMZMd7u5S7HenQqmzQHae\nDvg9uASU0lPt6tFfs4eDOhQVmX5CcUepjPCjnWzJ5u81UHGoHx7XZYb+aTMWSO40\n/DhvjIgEkttFrQ6jr77OS14rvfIUiMn7j8cS/4R4UrYZ8bBhbCNwjWJwii1KHkUC\nAwEAAaOBgTB/MB0GA1UdDgQWBBSPfiF4gdSbwM5sxgT1eMOUY9ETwDAfBgNVHSME\nGDAWgBSPfiF4gdSbwM5sxgT1eMOUY9ETwDAPBgNVHRMBAf8EBTADAQH/MCwGA1Ud\nEQQlMCOCCWxvY2FsaG9zdIcEfwAAAYcQAAAAAAAAAAAAAAAAAAAAATANBgkqhkiG\n9w0BAQsFAAOCAgEAjE87FRgMX/2wggYZxCt3+mkaCEhquBmoif9+2yDrbd6YecNX\nW+h8QLvwl003tio1SwEcTKwDiUsFvRtyQcy2o4wEjJ3lxHEt0N2ZznnO/DJdhPBD\nL2BW/L8siLCVHmpb3jcsUydXCDUoQKZQOGFZCYf43yZPQG8KLwCW3bJkdzWJ7Oo+\nNOWS1Mz+bFL8FLL4r8ReuSWD4m2C9erj19Xu3ZZ6gVHGHhqnT528VtGKVyL7dO4P\ng6tCeGMBfe2Cc6w8iYtEmW1/7scvXe+xKYrkWiJseIiJv/JjPaZ42pHfIUryDdLZ\nbgpxJANJ1gjJ2+F+598rPNkxkM8ourN74udJfxNqLiLBAVDO7Jxih4aRQiInBVD9\nynjzfgLSitFOvl1k0lVWfBFCSCG+Fb3h7MbAodTxMei5q5OwSNPx+fLz6tTVdz1V\nh3ISgoDmFvCVdobW+r54crX0HIgyX5qNA/16VeRaI17kSXjG3rt4tqEMmFw0hz53\ntVYIr23QvlhaPxVoGJZpD3Ihkh+8yv0KYrG3Zii6Q7t1KuwSNkATcRLEE+sNNI0C\nrrwRcyTWHyfezpHWDARFpaFbN+8yA3KiSrwVu/AQJtGdaFYoRTZfATWozXPQkTGl\nMbnIaC/twCZhtfhJFDA91z+B27JMDSSSvUJu3C7B34U6OOOaFTrfpaKcbV0=\n-----END CERTIFICATE-----\n"))
	_ = zfile.WriteFile(key, []byte("-----BEGIN PRIVATE KEY-----\n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\ncZkVyZKPdYIMs055zvaLFALMaf3gWiA=\n-----END PRIVATE KEY-----\n"))
	err := TlsCertificate(Certificate{
		CertFile: crt,
		KeyFile:  key,
	})
	t.EqualNil(err)
	r, err := Get("https://wx.qq.com")
	t.EqualNil(err)
	tt.Log(r.HTML().Find("title").Text(true))
}
