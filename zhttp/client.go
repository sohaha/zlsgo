package zhttp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zstring"
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
		Timeout:   10 * time.Minute,
	}
}

func (e *Engine) Client() *http.Client {
	if e.client == nil {
		e.client = newClient()
	}
	return e.client
}

func (e *Engine) SetClient(client *http.Client) {
	e.client = client
}

func (e *Engine) DisableChunke(enable ...bool) {
	state := true
	if len(enable) > 0 && enable[0] {
		state = false
	}
	e.disableChunke = state
}

func (e *Engine) Get(url string, v ...interface{}) (*Res, error) {
	return e.Do(http.MethodGet, url, v...)
}

func (e *Engine) Post(url string, v ...interface{}) (*Res, error) {
	return e.Do(http.MethodPost, url, v...)
}

func (e *Engine) Put(url string, v ...interface{}) (*Res, error) {
	return e.Do(http.MethodPut, url, v...)
}

func (e *Engine) Patch(url string, v ...interface{}) (*Res, error) {
	return e.Do(http.MethodPatch, url, v...)
}

func (e *Engine) Delete(url string, v ...interface{}) (*Res, error) {
	return e.Do(http.MethodDelete, url, v...)
}

func (e *Engine) Head(url string, v ...interface{}) (*Res, error) {
	return e.Do(http.MethodHead, url, v...)
}

func (e *Engine) Options(url string, v ...interface{}) (*Res, error) {
	return e.Do(http.MethodOptions, url, v...)
}

func (e *Engine) Trace(url string, v ...interface{}) (*Res, error) {
	return e.Do(http.MethodTrace, url, v...)
}

func (e *Engine) Connect(url string, v ...interface{}) (*Res, error) {
	return e.Do(http.MethodConnect, url, v...)
}

func (e *Engine) DoRetry(attempt int, sleep time.Duration, fn func() (*Res, error)) (*Res, error) {
	for attempt >= 0 {
		attempt--
		res, err := fn()
		if err != nil {
			time.Sleep(sleep)
			continue
		}
		return res, nil
	}
	return nil, errors.New("the number of retries has been exhausted")
}

func (e *Engine) EnableInsecureTLS(enable bool) {
	trans := e.getTransport()
	if trans == nil {
		return
	}
	if trans.TLSClientConfig == nil {
		trans.TLSClientConfig = &tls.Config{}
	}
	trans.TLSClientConfig.InsecureSkipVerify = enable
}

type Certificate struct {
	CertFile string
	KeyFile  string
}

func (e *Engine) TlsCertificate(certs ...Certificate) error {
	trans := e.getTransport()
	if trans == nil {
		return nil
	}
	if trans.TLSClientConfig == nil {
		trans.TLSClientConfig = &tls.Config{}
	}
	l := len(certs)
	certificates := make([]tls.Certificate, 0, l)
	for i := 0; i < l; i++ {
		x509KeyPair, err := tls.LoadX509KeyPair(certs[i].CertFile, certs[i].KeyFile)
		if err != nil {
			return err
		}
		certificates = append(certificates, x509KeyPair)
	}
	trans.TLSClientConfig.Certificates = certificates
	return nil
}

func (e *Engine) EnableCookie(enable bool) {
	if enable {
		jar, _ := cookiejar.New(nil)
		e.Client().Jar = jar
	} else {
		e.Client().Jar = nil
	}
}

func (e *Engine) CheckRedirect(fn ...func(req *http.Request, via []*http.Request) error) {
	if len(fn) > 0 {
		e.Client().CheckRedirect = fn[0]
	} else {
		e.Client().CheckRedirect = func(_ *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
}

func (e *Engine) SetTimeout(d time.Duration) {
	e.Client().Timeout = d
}

func (e *Engine) SetTransport(transport func(*http.Transport)) error {
	trans := e.getTransport()
	if trans == nil {
		return ErrNoTransport
	}
	transport(trans)
	return nil
}

func (e *Engine) SetProxyUrl(proxyUrl ...string) error {
	l := len(proxyUrl)
	if l == 0 {
		return errors.New("proxy url cannot be empty")
	}
	u := proxyUrl[0]
	return e.SetProxy(func(request *http.Request) (*url.URL, error) {
		if l > 1 {
			u = proxyUrl[zstring.RandInt(0, l-1)]
		}
		return url.Parse(u)
	})
}

func (e *Engine) SetProxy(proxy func(*http.Request) (*url.URL, error)) error {
	return e.SetTransport(func(transport *http.Transport) {
		transport.Proxy = proxy
	})
}

func (e *Engine) RemoveProxy() error {
	trans := e.getTransport()
	if trans == nil {
		return ErrNoTransport
	}
	trans.Proxy = http.ProxyFromEnvironment
	return nil
}

func (e *Engine) getJSONEncOpts() *jsonEncOpts {
	if e.jsonEncOpts == nil {
		e.jsonEncOpts = &jsonEncOpts{escapeHTML: true}
	}
	return e.jsonEncOpts
}

func (e *Engine) SetJSONEscapeHTML(escape bool) {
	opts := e.getJSONEncOpts()
	opts.escapeHTML = escape
}

func (e *Engine) SetJSONIndent(prefix, indent string) {
	opts := e.getJSONEncOpts()
	opts.indentPrefix = prefix
	opts.indentValue = indent
}

func (e *Engine) SetXMLIndent(prefix, indent string) {
	opts := e.getXMLEncOpts()
	opts.prefix = prefix
	opts.indent = indent
}

func (e *Engine) SetSsl(certPath, keyPath, CAPath string) (*tls.Config, error) {
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

	trans := e.getTransport()
	if trans == nil {
		return nil, ErrTransEmpty
	}

	trans.TLSClientConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}
	return trans.TLSClientConfig, nil
}

func (e *Engine) getTransport() *http.Transport {
	trans, _ := e.Client().Transport.(*http.Transport)
	return trans
}

func (e *Engine) getXMLEncOpts() *xmlEncOpts {
	if e.xmlEncOpts == nil {
		e.xmlEncOpts = &xmlEncOpts{}
	}
	return e.xmlEncOpts
}
