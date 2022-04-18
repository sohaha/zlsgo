package zhttp

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zjson"
)

type Res struct {
	r      *Engine
	req    *http.Request
	resp   *http.Response
	client *http.Client
	cost   time.Duration
	*multipartHelper
	requesterBody    []byte
	responseBody     []byte
	downloadProgress DownloadProgress
	err              error
	tmpFile          string
}

func (r *Res) Request() *http.Request {
	return r.req
}

func (r *Res) Response() *http.Response {

	return r.resp
}

func (r *Res) StatusCode() int {
	if r == nil || r.resp == nil {
		return 0
	}
	_, _ = r.ToBytes()
	return r.resp.StatusCode
}

func (r *Res) GetCookie() map[string]*http.Cookie {
	cookiesRaw := r.Response().Cookies()
	cookies := make(map[string]*http.Cookie, len(cookiesRaw))
	var cookie *http.Cookie
	for i := range cookiesRaw {
		if cookie = cookiesRaw[i]; cookie != nil {
			cookies[cookie.Name] = cookie
		}
	}
	return cookies
}

func (r *Res) Bytes() []byte {
	data, _ := r.ToBytes()
	return data
}

func (r *Res) ToBytes() ([]byte, error) {
	if r.err != nil || r.resp == nil {
		return nil, r.err
	}
	if r.responseBody != nil {
		return r.responseBody, nil
	}
	defer r.resp.Body.Close()
	respBody, err := ioutil.ReadAll(r.resp.Body)
	_, _ = io.Copy(ioutil.Discard, r.resp.Body)
	if err != nil {
		r.err = err
		return nil, err
	}
	r.responseBody = respBody
	return r.responseBody, nil
}

func (r *Res) Body() (body io.ReadCloser) {
	if r.err != nil {
		return nil
	}
	if r.responseBody != nil {
		return ioutil.NopCloser(bytes.NewReader(r.responseBody))
	}
	defer r.resp.Body.Close()
	respBody, err := ioutil.ReadAll(r.resp.Body)
	_, _ = io.Copy(ioutil.Discard, r.resp.Body)
	if err != nil {
		r.err = err
		return nil
	}
	r.responseBody = respBody
	return ioutil.NopCloser(bytes.NewReader(r.responseBody))
}

func (r *Res) HTML() (doc QueryHTML) {
	data, err := r.ToBytes()
	if err != nil {
		return QueryHTML{}
	}
	doc, _ = HTMLParse(data)
	return
}

func (r *Res) String() string {
	data, _ := r.ToBytes()
	return string(data)
}

func (r *Res) JSONs() zjson.Res {
	data, _ := r.ToBytes()
	return zjson.ParseBytes(data)
}

func (r *Res) JSON(key string) zjson.Res {
	j := r.JSONs()
	return j.Get(key)
}

func (r *Res) ToString() (string, error) {
	data, err := r.ToBytes()
	return string(data), err
}

func (r *Res) ToJSON(v interface{}) error {
	data, err := r.ToBytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (r *Res) ToXML(v interface{}) error {
	data, err := r.ToBytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

func (r *Res) ToFile(name string) error {
	nameSplit := strings.Split(zfile.RealPath(name), "/")
	nameSplitLen := len(nameSplit)
	if nameSplitLen > 1 {
		dir := strings.Join(nameSplit[0:nameSplitLen-1], "/")
		name = zfile.RealPathMkdir(dir) + "/" + nameSplit[nameSplitLen-1]
	}

	if r.tmpFile != "" {
		return zfile.CopyFile(r.tmpFile, name)
	}

	file, err := os.Create(name)
	if err != nil {
		return err
	}
	//noinspection GoUnhandledErrorResult
	defer file.Close()

	if r.responseBody != nil {
		_, err = file.Write(r.responseBody)
		return err
	}

	if r.downloadProgress != nil && r.resp.ContentLength > 0 {
		err = r.download(file)
	} else {
		//noinspection GoUnhandledErrorResult
		defer r.resp.Body.Close()
		_, err = io.Copy(file, r.resp.Body)
	}
	if err == nil {
		r.tmpFile = name
	}
	return err
}

func (r *Res) download(file *os.File) error {
	var (
		current  int64
		lastTime time.Time
	)
	p, b := make([]byte, 1024), r.resp.Body
	duration, total := 200*time.Millisecond, r.resp.ContentLength
	//noinspection GoUnhandledErrorResult
	defer b.Close()
	for {
		l, err := b.Read(p)
		if l > 0 {
			_, _err := file.Write(p[:l])
			if _err != nil {
				return _err
			}
			current += int64(l)
			if now := time.Now(); now.Sub(lastTime) > duration {
				lastTime = now
				r.downloadProgress(current, total)
			}
		}
		if err != nil {
			if err == io.EOF {
				r.downloadProgress(total, total)
				return nil
			}
			return err
		}
	}
}
