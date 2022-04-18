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

func (r *Engine) Client() *http.Client {
	if r.client == nil {
		r.client = newClient()
	}
	return r.client
}

func (r *Engine) SetClient(client *http.Client) {
	r.client = client
}

func (r *Engine) DisableChunke(enable ...bool) {
	state := true
	if len(enable) > 0 && enable[0] {
		state = false
	}
	r.disableChunke = state
}

func (r *Engine) Get(url string, v ...interface{}) (*Res, error) {
	return r.Do(http.MethodGet, url, v...)
}

func (r *Engine) Post(url string, v ...interface{}) (*Res, error) {
	return r.Do(http.MethodPost, url, v...)
}

func (r *Engine) Put(url string, v ...interface{}) (*Res, error) {
	return r.Do(http.MethodPut, url, v...)
}

func (r *Engine) Patch(url string, v ...interface{}) (*Res, error) {
	return r.Do(http.MethodPatch, url, v...)
}

func (r *Engine) Delete(url string, v ...interface{}) (*Res, error) {
	return r.Do(http.MethodDelete, url, v...)
}

func (r *Engine) Head(url string, v ...interface{}) (*Res, error) {
	return r.Do(http.MethodHead, url, v...)
}

func (r *Engine) Options(url string, v ...interface{}) (*Res, error) {
	return r.Do(http.MethodOptions, url, v...)
}

func (r *Engine) Trace(url string, v ...interface{}) (*Res, error) {
	return r.Do(http.MethodTrace, url, v...)
}

func (r *Engine) Connect(url string, v ...interface{}) (*Res, error) {
	return r.Do(http.MethodConnect, url, v...)
}

func (r *Engine) DoRetry(attempt int, sleep time.Duration, fn func() (*Res, error)) (*Res, error) {
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

func (r *Engine) EnableInsecureTLS(enable bool) {
	trans := r.getTransport()
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

func (r *Engine) TlsCertificate(certs ...Certificate) error {
	trans := r.getTransport()
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

func (r *Engine) EnableCookie(enable bool) {
	if enable {
		jar, _ := cookiejar.New(nil)
		r.Client().Jar = jar
	} else {
		r.Client().Jar = nil
	}
}

func (r *Engine) CheckRedirect(fn ...func(req *http.Request, via []*http.Request) error) {
	if len(fn) > 0 {
		r.Client().CheckRedirect = fn[0]
	} else {
		r.Client().CheckRedirect = func(_ *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
}

func (r *Engine) SetTimeout(d time.Duration) {
	r.Client().Timeout = d
}

func (r *Engine) SetTransport(transport func(*http.Transport)) error {
	trans := r.getTransport()
	if trans == nil {
		return ErrNoTransport
	}
	transport(trans)
	return nil
}

func (r *Engine) SetProxyUrl(proxyUrl ...string) error {
	l := len(proxyUrl)
	if l == 0 {
		return errors.New("proxy url cannot be empty")
	}
	u := proxyUrl[0]
	return r.SetProxy(func(request *http.Request) (*url.URL, error) {
		if l > 1 {
			u = proxyUrl[zstring.RandInt(0, l-1)]
		}
		return url.Parse(u)
	})
}

func (r *Engine) SetProxy(proxy func(*http.Request) (*url.URL, error)) error {
	return r.SetTransport(func(transport *http.Transport) {
		transport.Proxy = proxy
	})
}

func (r *Engine) RemoveProxy() error {
	trans := r.getTransport()
	if trans == nil {
		return ErrNoTransport
	}
	trans.Proxy = http.ProxyFromEnvironment
	return nil
}

func (r *Engine) getJSONEncOpts() *jsonEncOpts {
	if r.jsonEncOpts == nil {
		r.jsonEncOpts = &jsonEncOpts{escapeHTML: true}
	}
	return r.jsonEncOpts
}

func (r *Engine) SetJSONEscapeHTML(escape bool) {
	opts := r.getJSONEncOpts()
	opts.escapeHTML = escape
}

func (r *Engine) SetJSONIndent(prefix, indent string) {
	opts := r.getJSONEncOpts()
	opts.indentPrefix = prefix
	opts.indentValue = indent
}

func (r *Engine) SetXMLIndent(prefix, indent string) {
	opts := r.getXMLEncOpts()
	opts.prefix = prefix
	opts.indent = indent
}

func (r *Engine) SetSsl(certPath, keyPath, CAPath string) (*tls.Config, error) {
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

func (r *Engine) getTransport() *http.Transport {
	trans, _ := r.Client().Transport.(*http.Transport)
	return trans
}

func (r *Engine) getXMLEncOpts() *xmlEncOpts {
	if r.xmlEncOpts == nil {
		r.xmlEncOpts = &xmlEncOpts{}
	}
	return r.xmlEncOpts
}
