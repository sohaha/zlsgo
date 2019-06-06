/*
 * @Author: seekwe
 * @Date:   2019-05-30 13:19:45
 * @Last Modified by:   seekwe
 * @Last Modified time: 2019-05-30 13:41:41
 */

package zhttp

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/sohaha/zlsgo/zlog"
)

func newClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Jar:       jar,
		Transport: transport,
		Timeout:   2 * time.Minute,
	}
}

func (r *R) Client() *http.Client {
	if r.client == nil {
		r.client = newClient()
	}
	return r.client
}

func (r *R) SetClient(client *http.Client) {
	r.client = client
}

func (r *R) Get(url string, v ...interface{}) (*Res, error) {
	return r.Do("GET", url, v...)
}

func (r *R) Post(url string, v ...interface{}) (*Res, error) {
	return r.Do("POST", url, v...)
}

func (r *R) Put(url string, v ...interface{}) (*Res, error) {
	return r.Do("PUT", url, v...)
}

func (r *R) Patch(url string, v ...interface{}) (*Res, error) {
	return r.Do("PATCH", url, v...)
}

func (r *R) Delete(url string, v ...interface{}) (*Res, error) {
	return r.Do("DELETE", url, v...)
}

func (r *R) Head(url string, v ...interface{}) (*Res, error) {
	return r.Do("HEAD", url, v...)
}

func (r *R) Options(url string, v ...interface{}) (*Res, error) {
	return r.Do("OPTIONS", url, v...)
}

func (r *R) getTransport() *http.Transport {
	trans, _ := r.Client().Transport.(*http.Transport)
	return trans
}

func (r *R) EnableInsecureTLS(enable bool) {
	trans := r.getTransport()
	if trans == nil {
		return
	}
	if trans.TLSClientConfig == nil {
		trans.TLSClientConfig = &tls.Config{}
	}
	trans.TLSClientConfig.InsecureSkipVerify = enable
}

func (r *R) EnableCookie(enable bool) {
	if enable {
		jar, _ := cookiejar.New(nil)
		r.Client().Jar = jar
	} else {
		r.Client().Jar = nil
	}
}

func (r *R) SetTimeout(d time.Duration) {
	r.Client().Timeout = d
}

func (r *R) SetProxyUrl(rawurl string) error {
	trans := r.getTransport()
	if trans == nil {
		return ErrNoTransport
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return err
	}
	trans.Proxy = http.ProxyURL(u)
	return nil
}

func (r *R) SetProxy(proxy func(*http.Request) (*url.URL, error)) error {
	trans := r.getTransport()
	if trans == nil {
		return ErrNoTransport
	}
	trans.Proxy = proxy
	return nil
}

func (r *R) getJSONEncOpts() *jsonEncOpts {
	if r.jsonEncOpts == nil {
		r.jsonEncOpts = &jsonEncOpts{escapeHTML: true}
	}
	return r.jsonEncOpts
}

func (r *R) SetJSONEscapeHTML(escape bool) {
	opts := r.getJSONEncOpts()
	opts.escapeHTML = escape
}

func (r *R) SetJSONIndent(prefix, indent string) {
	opts := r.getJSONEncOpts()
	opts.indentPrefix = prefix
	opts.indentValue = indent
}

func (r *R) getXMLEncOpts() *xmlEncOpts {
	if r.xmlEncOpts == nil {
		r.xmlEncOpts = &xmlEncOpts{}
	}
	return r.xmlEncOpts
}

func (r *R) SetXMLIndent(prefix, indent string) {
	opts := r.getXMLEncOpts()
	opts.prefix = prefix
	opts.indent = indent
}

func (r *R) SetSsl(certPath, keyPath, CAPath string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		zlog.Error("load keys fail", err)
		return nil, err
	}

	caData, err := ioutil.ReadFile(CAPath)
	if err != nil {
		zlog.Error("read ca fail", err)
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caData)

	trans := r.getTransport()
	if trans == nil {
		return nil, ErrTransEmpty
	}

	trans.TLSClientConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}
	return trans.TLSClientConfig, nil
}

func (r *Res) autoFormat(s fmt.State) {
	req := r.req
	if r.r.flag&BitTime != 0 {
		fmt.Fprint(s, req.Method, " ", req.URL.String(), " ", r.cost)
	} else {
		fmt.Fprint(s, req.Method, " ", req.URL.String())
	}

	var pretty bool
	var parts []string
	addPart := func(part string) {
		if part == "" {
			return
		}
		parts = append(parts, part)
		if !pretty && regNewline.MatchString(part) {
			pretty = true
		}
	}
	if r.r.flag&BitReqBody != 0 {
		addPart(string(r.requesterBody))
	}
	if r.r.flag&BitRespBody != 0 {
		addPart(r.String())
	}

	for _, part := range parts {
		if pretty {
			fmt.Fprint(s, "\n")
		}
		fmt.Fprint(s, " ", part)
	}
}

func (r *Res) miniFormat(s fmt.State) {
	req := r.req
	if r.r.flag&BitTime != 0 {
		fmt.Fprint(s, req.Method, " ", req.URL.String(), " ", r.cost)
	} else {
		fmt.Fprint(s, req.Method, " ", req.URL.String())
	}
	if r.r.flag&BitReqBody != 0 && len(r.requesterBody) > 0 {
		str := regNewline.ReplaceAllString(string(r.requesterBody), " ")
		fmt.Fprint(s, " ", str)
	}
	if r.r.flag&BitRespBody != 0 && r.String() != "" {
		str := regNewline.ReplaceAllString(r.String(), " ")
		fmt.Fprint(s, " ", str)
	}
}

func (r *Res) Format(s fmt.State, verb rune) {
	if r == nil || r.req == nil {
		return
	}
	if s.Flag('+') {
		fmt.Fprint(s, r.Dump())
	} else if s.Flag('-') {
		r.miniFormat(s)
	} else {
		r.autoFormat(s)
	}
}
