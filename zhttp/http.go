// Package zhttp provides http client related operations
package zhttp

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zcache/fast"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

const (
	textContentType = "Content-Type"
)

const (
	BitReqHead = 1 << iota
	BitReqBody
	BitRespHead
	BitRespBody
	BitTime
	BitStdFlags = BitReqHead | BitReqBody | BitRespHead | BitRespBody
)

type (
	Header     map[string]string
	Param      map[string]interface{}
	QueryParam map[string]interface{}
	Host       string
	FileUpload struct {
		File      io.ReadCloser
		FileName  string
		FieldName string
	}
	DownloadProgress func(current, total int64)
	UploadProgress   func(current, total int64)
	Engine           struct {
		client         *zutil.Pointer
		jsonEncOpts    *jsonEncOpts
		xmlEncOpts     *xmlEncOpts
		getUserAgent   func() string
		urlCache       *fast.FastCache
		flag           int
		debug          bool
		disableChunked bool
	}

	bodyJson struct {
		v interface{}
	}
	bodyXml struct {
		v interface{}
	}

	NoRedirect bool

	CustomReq func(req *http.Request)

	param struct {
		url.Values
	}

	bodyWrapper struct {
		io.ReadCloser
		buf   bytes.Buffer
		limit int
	}

	multipartHelper struct {
		form           url.Values
		uploadProgress UploadProgress
		uploads        []FileUpload
		dump           []byte
	}

	jsonEncOpts struct {
		indentPrefix string
		indentValue  string
		escapeHTML   bool
	}

	xmlEncOpts struct {
		prefix string
		indent string
	}
)

var std = New()

var (
	ErrNoTransport     = errors.New("no transport")
	ErrUrlNotSpecified = errors.New("url not specified")
	ErrTransEmpty      = errors.New("trans is empty")
	ErrNoMatched       = errors.New("no file have been matched")
)

type Request struct {
	body         interface{}
	ctx          context.Context
	client       *http.Client
	uploadProg   UploadProgress
	queryParams  QueryParam
	formParams   Param
	headers      Header
	customReq    CustomReq
	engine       *Engine
	downloadProg DownloadProgress
	method       string
	host         string
	url          string
	cookies      []*http.Cookie
	uploads      []FileUpload
	noRedirect   bool
}

type RequestArgs struct {
	Body         interface{}
	Ctx          context.Context
	UploadProg   UploadProgress
	FormParams   Param
	Client       *http.Client
	QueryParams  QueryParam
	Headers      Header
	DownloadProg DownloadProgress
	CustomReq    CustomReq
	Host         string
	Uploads      []FileUpload
	Cookies      []*http.Cookie
	NoRedirect   bool
}

var requestPool = sync.Pool{
	New: func() interface{} {
		return &Request{
			headers:     make(Header, 8),
			queryParams: make(QueryParam, 8),
			formParams:  make(Param, 8),
			cookies:     make([]*http.Cookie, 0, 4),
			uploads:     make([]FileUpload, 0, 2),
		}
	},
}

// New create a new Engine
func New() *Engine {
	e := &Engine{
		flag:  BitStdFlags,
		debug: Debug.Load(),
		urlCache: fast.NewFast(func(o *fast.Options) {
			o.Bucket = uint16(4)
			o.Expiration = 0
			o.AutoCleaner = false
		}),
		client: zutil.NewPointer(nil),
	}
	e.SetClient(newClient())
	return e
}

func (e *Engine) parseURL(rawurl string) (*url.URL, error) {
	if u, ok := e.urlCache.Get(rawurl); ok {
		return u.(*url.URL), nil
	}

	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	e.urlCache.Set(rawurl, u)
	return u, nil
}

func (p *param) getValues() url.Values {
	if p.Values == nil {
		p.Values = make(url.Values)
	}
	return p.Values
}

func (p *param) Copy(pp param) {
	if pp.Values == nil {
		return
	}
	vs := p.getValues()
	for key, values := range pp.Values {
		for _, value := range values {
			vs.Add(key, value)
		}
	}
}

// Adds Add multiple parameters
func (p *param) Adds(m map[string]interface{}) {
	if len(m) == 0 {
		return
	}
	vs := p.getValues()
	for k, v := range m {
		var str string
		switch val := v.(type) {
		case string:
			str = val
		case int:
			str = strconv.Itoa(val)
		case int64:
			str = strconv.FormatInt(val, 10)
		case uint:
			str = strconv.FormatUint(uint64(val), 10)
		case uint64:
			str = strconv.FormatUint(val, 10)
		case float64:
			str = strconv.FormatFloat(val, 'f', -1, 64)
		case float32:
			str = strconv.FormatFloat(float64(val), 'f', -1, 32)
		case bool:
			str = strconv.FormatBool(val)
		default:
			str = fmt.Sprint(v)
		}
		vs.Add(k, str)
	}
}

func (p *param) Empty() bool {
	return p.Values == nil
}

func (e *Engine) Do(method, rawurl string, vs ...interface{}) (resp *Res, err error) {
	if rawurl == "" {
		return nil, ErrUrlNotSpecified
	}

	var (
		queryParam     param
		formParam      param
		uploads        []FileUpload
		uploadProgress UploadProgress
		progress       func(int64, int64)
		delayedFunc    []func()
		lastFunc       []func()
	)

	req := &http.Request{
		Method:     strings.ToUpper(method),
		Header:     make(http.Header, 8),
		Proto:      "Engine/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	resp = &Res{req: req, r: e}
	if e.getUserAgent != nil {
		ua := e.getUserAgent()
		if ua == "" {
			ua = UserAgentLists[zstring.RandInt(0, len(UserAgentLists)-1)]
		}
		req.Header.Add("User-Agent", ua)
	}
	for _, v := range vs {
		switch vv := v.(type) {
		case NoRedirect:
			if vv {
				r := e.Client().CheckRedirect
				e.Client().CheckRedirect = func(_ *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
				defer func() {
					e.Client().CheckRedirect = r
				}()
			}
		case CustomReq:
			vv(req)
		case Header:
			for key, value := range vv {
				req.Header.Add(key, value)
			}
		case http.Header:
			for key, values := range vv {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}
		case *bodyJson:
			fn, err := setBodyJson(req, resp, e.jsonEncOpts, vv.v)
			if err != nil {
				return nil, err
			}
			delayedFunc = append(delayedFunc, fn)
		case *bodyXml:
			fn, err := setBodyXml(req, resp, e.xmlEncOpts, vv.v)
			if err != nil {
				return nil, err
			}
			delayedFunc = append(delayedFunc, fn)
		case url.Values:
			p := param{vv}
			if method == "GET" || method == "HEAD" {
				queryParam.Copy(p)
			} else {
				formParam.Copy(p)
			}
		case Param:
			if method == "GET" || method == "HEAD" {
				queryParam.Adds(vv)
			} else {
				formParam.Adds(vv)
			}
		case QueryParam:
			queryParam.Adds(vv)
		case string:
			setBodyBytes(req, resp, []byte(vv))
		case []byte:
			setBodyBytes(req, resp, vv)
		case bytes.Buffer:
			setBodyBytes(req, resp, vv.Bytes())
		case *http.Client:
			resp.client = vv
		case FileUpload:
			uploads = append(uploads, vv)
		case []FileUpload:
			uploads = append(uploads, vv...)
		case map[string]*http.Cookie:
			for i := range vv {
				req.AddCookie(vv[i])
			}
		case *http.Cookie:
			req.AddCookie(vv)
		case Host:
			req.Host = string(vv)
		case io.Reader:
			fn := setBodyReader(req, resp, vv)
			lastFunc = append(lastFunc, fn)
		case UploadProgress:
			uploadProgress = vv
		case DownloadProgress:
			resp.downloadProgress = vv
		case func(int64, int64):
			progress = vv
		case context.Context:
			req = req.WithContext(vv)
			resp.req = req
		case error:
			return resp, vv
		}
	}

	if length := req.Header.Get("Content-Length"); length != "" {
		if l, err := strconv.ParseInt(length, 10, 64); err == nil {
			req.ContentLength = l
		}
	}

	if len(uploads) > 0 && (req.Method == "POST" || req.Method == "PUT") {
		var up UploadProgress
		if uploadProgress != nil {
			up = uploadProgress
		} else if progress != nil {
			up = UploadProgress(progress)
		}
		multipartHelper := &multipartHelper{
			form:           formParam.Values,
			uploads:        uploads,
			uploadProgress: up,
		}
		if e.disableChunked {
			multipartHelper.Upload(req)
		} else {
			multipartHelper.UploadChunke(req)
		}
		resp.multipartHelper = multipartHelper
	} else {
		if progress != nil {
			resp.downloadProgress = DownloadProgress(progress)
		}
		if !formParam.Empty() {
			if req.Body != nil {
				queryParam.Copy(formParam)
			} else {
				setBodyBytes(req, resp, []byte(formParam.Encode()))
				setContentType(req, "application/x-www-form-urlencoded; charset=UTF-8")
			}
		}
	}

	if !queryParam.Empty() {
		paramStr := queryParam.Encode()
		requiredSize := len(rawurl) + 1 + len(paramStr)

		if strings.IndexByte(rawurl, '?') == -1 {
			sb := zutil.GetBuff(uint(requiredSize))
			sb.WriteString(rawurl)
			sb.WriteByte('?')
			sb.WriteString(paramStr)
			rawurl = sb.String()
			zutil.PutBuff(sb)
		} else {
			sb := zutil.GetBuff(uint(requiredSize))
			sb.WriteString(rawurl)
			sb.WriteByte('&')
			sb.WriteString(paramStr)
			rawurl = sb.String()
			zutil.PutBuff(sb)
		}
	}
	var u *url.URL
	u, err = e.parseURL(rawurl) // 使用缓存的 URL 解析
	if err != nil {
		return
	}
	req.URL = u

	if host := req.Header.Get("Host"); host != "" {
		req.Host = host
	}

	for _, fn := range delayedFunc {
		fn()
	}

	if resp.client == nil {
		resp.client = e.Client()
	}

	var response *http.Response

	if e.flag&BitTime != 0 {
		before := time.Now()
		response, err = resp.client.Do(req)
		after := time.Now()
		resp.cost = after.Sub(before)
	} else {
		response, err = resp.client.Do(req)
	}

	if err != nil {
		return
	}

	for _, fn := range lastFunc {
		fn()
	}

	resp.resp = response

	if _, ok := resp.client.Transport.(*http.Transport); ok && response.Header.Get("Content-Encoding") == "gzip" && req.Header.Get("Accept-Encoding") != "" {
		var body *gzip.Reader
		body, err = gzip.NewReader(response.Body)
		if err != nil {
			return
		}
		response.Body = body
	}

	if //noinspection GoBoolExpressions
	Debug.Load() || e.debug {
		zlog.Println(resp.Dump())
	}

	return
}

func setBodyBytes(req *http.Request, resp *Res, data []byte) {
	resp.requesterBody = data
	req.Body = io.NopCloser(bytes.NewReader(data))
	req.ContentLength = int64(len(data))
}

func setBodyJson(req *http.Request, resp *Res, opts *jsonEncOpts, v interface{}) (func(), error) {
	var data []byte
	switch vv := v.(type) {
	case string:
		data = []byte(vv)
	case []byte:
		data = vv
	case *bytes.Buffer:
		data = vv.Bytes()
	default:
		if opts != nil {
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.SetIndent(opts.indentPrefix, opts.indentValue)
			enc.SetEscapeHTML(opts.escapeHTML)
			err := enc.Encode(v)
			if err != nil {
				return nil, err
			}
			data = buf.Bytes()
		} else {
			var err error
			data, err = zjson.Marshal(v)
			if err != nil {
				return nil, err
			}
		}
	}
	setBodyBytes(req, resp, data)
	delayedFunc := func() {
		setContentType(req, "application/json; charset=UTF-8")
	}
	return delayedFunc, nil
}

func setBodyXml(req *http.Request, resp *Res, opts *xmlEncOpts, v interface{}) (func(), error) {
	var data []byte
	switch vv := v.(type) {
	case string:
		data = []byte(vv)
	case []byte:
		data = vv
	case *bytes.Buffer:
		data = vv.Bytes()
	default:
		if opts != nil {
			var buf bytes.Buffer
			enc := xml.NewEncoder(&buf)
			enc.Indent(opts.prefix, opts.indent)
			err := enc.Encode(v)
			if err != nil {
				return nil, err
			}
			data = buf.Bytes()
		} else {
			var err error
			data, err = xml.Marshal(v)
			if err != nil {
				return nil, err
			}
		}
	}
	setBodyBytes(req, resp, data)
	delayedFunc := func() {
		setContentType(req, "application/xml; charset=UTF-8")
	}
	return delayedFunc, nil
}

func setContentType(req *http.Request, contentType string) {
	if req.Header.Get(textContentType) == "" {
		req.Header.Set(textContentType, contentType)
	}
}

func setBodyReader(req *http.Request, resp *Res, rd io.Reader) func() {
	var rc io.ReadCloser
	switch r := rd.(type) {
	case *os.File:
		stat, err := r.Stat()
		if err == nil {
			req.ContentLength = stat.Size()
		}
		rc = r

	case io.ReadCloser:
		rc = r
	default:
		rc = io.NopCloser(rd)
	}
	bw := &bodyWrapper{
		ReadCloser: rc,
		limit:      102400,
	}
	req.Body = bw
	lastFunc := func() {
		resp.requesterBody = bw.buf.Bytes()
	}
	return lastFunc
}

func (b *bodyWrapper) Read(p []byte) (n int, err error) {
	n, err = b.ReadCloser.Read(p)
	if left := b.limit - b.buf.Len(); left > 0 && n > 0 {
		if n <= left {
			b.buf.Write(p[:n])
		} else {
			b.buf.Write(p[:left])
		}
	}
	return
}

func (m *multipartHelper) upload(req *http.Request, upload func(io.Writer, io.Reader) error, bodyWriter *multipart.Writer) {
	for key, values := range m.form {
		for _, value := range values {
			_ = bodyWriter.WriteField(key, value)
		}
	}

	i := 0
	for _, up := range m.uploads {
		if up.FieldName == "" {
			i++
			up.FieldName = "file" + strconv.Itoa(i)
		}

		fileWriter, err := bodyWriter.CreateFormFile(up.FieldName, up.FileName)
		if err != nil {
			_ = up.File.Close()
			continue
		}

		if upload == nil {
			_, _ = io.Copy(fileWriter, up.File)
		} else {
			if _, ok := up.File.(*os.File); ok {
				_ = upload(fileWriter, up.File)
			} else {
				_, _ = io.Copy(fileWriter, up.File)
			}
		}
		_ = up.File.Close()
	}
}

func (m *multipartHelper) Upload(req *http.Request) {
	bodyBuf := zutil.GetBuff(1048576)
	defer zutil.PutBuff(bodyBuf)

	bodyWriter := multipart.NewWriter(bodyBuf)

	m.upload(req, nil, bodyWriter)
	_ = bodyWriter.Close()

	req.Header.Set(textContentType, bodyWriter.FormDataContentType())

	bodyBytes := make([]byte, bodyBuf.Len())
	copy(bodyBytes, bodyBuf.Bytes())
	b := bytes.NewReader(bodyBytes)

	req.Body = io.NopCloser(b)
	req.ContentLength = int64(b.Len())
}

func (m *multipartHelper) UploadChunke(req *http.Request) {
	pr, pw := io.Pipe()
	bodyWriter := multipart.NewWriter(pw)
	go func() {
		defer func() {
			_ = bodyWriter.Close()
			_ = pw.Close()
		}()

		var upload func(io.Writer, io.Reader) error

		if m.uploadProgress != nil {
			var (
				total    int64
				current  int64
				lastTime time.Time
			)
			for _, up := range m.uploads {
				if file, ok := up.File.(*os.File); ok {
					stat, err := file.Stat()
					if err != nil {
						continue
					}
					total += stat.Size()
				}
			}
			duration, buf := 200*time.Millisecond, make([]byte, 1024)
			upload = func(w io.Writer, r io.Reader) error {
				for {
					n, err := r.Read(buf)
					if n > 0 {
						_, _err := w.Write(buf[:n])
						if _err != nil {
							return _err
						}
						current += int64(n)
						if now := time.Now(); now.Sub(lastTime) > duration {
							lastTime = now
							m.uploadProgress(current, total)
						}
					}
					if err == io.EOF {
						m.uploadProgress(total, total)
						return nil
					}
					if err != nil {
						return err
					}
				}
			}
		}
		m.upload(req, upload, bodyWriter)
	}()
	req.Header.Set(textContentType, bodyWriter.FormDataContentType())
	req.Body = io.NopCloser(pr)
}

func (m *multipartHelper) Dump() []byte {
	if m.dump != nil {
		return m.dump
	}
	var buf bytes.Buffer
	bodyWriter := multipart.NewWriter(&buf)
	for key, values := range m.form {
		for _, value := range values {
			_ = m.writeField(bodyWriter, key, value)
		}
	}
	for _, up := range m.uploads {
		_ = m.writeFile(bodyWriter, up.FieldName, up.FileName)
	}
	_ = bodyWriter.Close()
	m.dump = buf.Bytes()
	return m.dump
}

func (m *multipartHelper) writeField(w *multipart.Writer, fieldname, value string) error {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"`, fieldname))
	p, err := w.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = p.Write([]byte(value))
	return err
}

func (m *multipartHelper) writeFile(w *multipart.Writer, fieldname, filename string) error {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			fieldname, filename))
	h.Set(textContentType, "application/octet-stream")
	p, err := w.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = p.Write([]byte("******"))
	return err
}

func Client() *http.Client {
	return std.Client()
}

func SetClient(client *http.Client) {
	std.SetClient(client)
}

func (e *Engine) SetFlags(flags int) {
	e.flag = flags
}

func (e *Engine) GetFlags() int {
	return e.flag
}

func (e *Engine) SetUserAgent(fn func() string) {
	e.getUserAgent = fn
}

func SetFlags(flags int) {
	std.SetFlags(flags)
}

func Flags() int {
	return std.GetFlags()
}

func EnableInsecureTLS(enable bool) {
	std.EnableInsecureTLS(enable)
}

func TlsCertificate(certs ...Certificate) error {
	return std.TlsCertificate(certs...)
}

func EnableCookie(enable bool) error {
	return std.EnableCookie(enable)
}

func SetTimeout(d time.Duration) {
	std.SetTimeout(d)
}

func RemoveProxy() error {
	return std.RemoveProxy()
}

// SetUserAgent returning an empty array means random built-in User Agent
func SetUserAgent(fn func() string) {
	std.SetUserAgent(fn)
}

// SetTransport SetTransport
func SetTransport(transport func(*http.Transport)) error {
	return std.SetTransport(transport)
}

// SetProxyUrl SetProxyUrl
func SetProxyUrl(proxyUrl ...string) error {
	return std.SetProxyUrl(proxyUrl...)
}

// SetProxy SetProxy
func SetProxy(proxy func(*http.Request) (*url.URL, error)) error {
	return std.SetProxy(proxy)
}

func SetJSONEscapeHTML(escape bool) {
	std.SetJSONEscapeHTML(escape)
}

func SetJSONIndent(prefix, indent string) {
	std.SetJSONIndent(prefix, indent)
}

func SetXMLIndent(prefix, indent string) {
	std.SetXMLIndent(prefix, indent)
}

// DoWithArgs use struct args to do request
func (e *Engine) DoWithArgs(method, rawurl string, args *RequestArgs) (*Res, error) {
	if rawurl == "" {
		return nil, ErrUrlNotSpecified
	}

	var (
		queryParam     param
		formParam      param
		uploads        []FileUpload
		uploadProgress UploadProgress
		progress       func(int64, int64)
		delayedFunc    []func()
		lastFunc       []func()
	)

	req := &http.Request{
		Method:     strings.ToUpper(method),
		Header:     make(http.Header, 8),
		Proto:      "Engine/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}
	resp := &Res{req: req, r: e}

	if e.getUserAgent != nil {
		ua := e.getUserAgent()
		if ua == "" {
			ua = UserAgentLists[zstring.RandInt(0, len(UserAgentLists)-1)]
		}
		req.Header.Add("User-Agent", ua)
	}

	if args != nil {
		if args.NoRedirect {
			r := e.Client().CheckRedirect
			e.Client().CheckRedirect = func(_ *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
			defer func() {
				e.Client().CheckRedirect = r
			}()
		}

		if args.CustomReq != nil {
			args.CustomReq(req)
		}

		if len(args.Headers) > 0 {
			for key, value := range args.Headers {
				req.Header.Add(key, value)
			}
		}

		if args.Body != nil {
			switch vv := args.Body.(type) {
			case *bodyJson:
				fn, err := setBodyJson(req, resp, e.jsonEncOpts, vv.v)
				if err != nil {
					return nil, err
				}
				delayedFunc = append(delayedFunc, fn)
			case *bodyXml:
				fn, err := setBodyXml(req, resp, e.xmlEncOpts, vv.v)
				if err != nil {
					return nil, err
				}
				delayedFunc = append(delayedFunc, fn)
			case string:
				setBodyBytes(req, resp, []byte(vv))
			case []byte:
				setBodyBytes(req, resp, vv)
			case bytes.Buffer:
				setBodyBytes(req, resp, vv.Bytes())
			case io.Reader:
				fn := setBodyReader(req, resp, vv)
				lastFunc = append(lastFunc, fn)
			}
		}

		if len(args.QueryParams) > 0 {
			queryParam.Adds(args.QueryParams)
		}
		if len(args.FormParams) > 0 {
			if method == "GET" || method == "HEAD" {
				queryParam.Adds(args.FormParams)
			} else {
				formParam.Adds(args.FormParams)
			}
		}

		if args.Client != nil {
			resp.client = args.Client
		}

		if len(args.Uploads) > 0 {
			uploads = args.Uploads
		}

		for _, cookie := range args.Cookies {
			if cookie != nil {
				req.AddCookie(cookie)
			}
		}

		if args.Host != "" {
			req.Host = args.Host
		}
		if args.Ctx != nil {
			req = req.WithContext(args.Ctx)
			resp.req = req
		}

		if args.UploadProg != nil {
			uploadProgress = args.UploadProg
		}
		if args.DownloadProg != nil {
			resp.downloadProgress = args.DownloadProg
		}
	}

	if length := req.Header.Get("Content-Length"); length != "" {
		if l, err := strconv.ParseInt(length, 10, 64); err == nil {
			req.ContentLength = l
		}
	}

	if len(uploads) > 0 && (req.Method == "POST" || req.Method == "PUT") {
		var up UploadProgress
		if uploadProgress != nil {
			up = uploadProgress
		} else if progress != nil {
			up = UploadProgress(progress)
		}
		multipartHelper := &multipartHelper{
			form:           formParam.Values,
			uploads:        uploads,
			uploadProgress: up,
		}
		if e.disableChunked {
			multipartHelper.Upload(req)
		} else {
			multipartHelper.UploadChunke(req)
		}
		resp.multipartHelper = multipartHelper
	} else {
		if progress != nil {
			resp.downloadProgress = DownloadProgress(progress)
		}
		if !formParam.Empty() {
			if req.Body != nil {
				queryParam.Copy(formParam)
			} else {
				setBodyBytes(req, resp, []byte(formParam.Encode()))
				setContentType(req, "application/x-www-form-urlencoded; charset=UTF-8")
			}
		}
	}

	if !queryParam.Empty() {
		paramStr := queryParam.Encode()
		requiredSize := len(rawurl) + 1 + len(paramStr)

		if strings.IndexByte(rawurl, '?') == -1 {
			sb := zutil.GetBuff(uint(requiredSize))
			sb.WriteString(rawurl)
			sb.WriteByte('?')
			sb.WriteString(paramStr)
			rawurl = sb.String()
			zutil.PutBuff(sb)
		} else {
			sb := zutil.GetBuff(uint(requiredSize))
			sb.WriteString(rawurl)
			sb.WriteByte('&')
			sb.WriteString(paramStr)
			rawurl = sb.String()
			zutil.PutBuff(sb)
		}
	}

	var u *url.URL
	var err error
	u, err = e.parseURL(rawurl)
	if err != nil {
		return nil, err
	}
	req.URL = u

	if host := req.Header.Get("Host"); host != "" {
		req.Host = host
	}

	for _, fn := range delayedFunc {
		fn()
	}

	if resp.client == nil {
		resp.client = e.Client()
	}

	var response *http.Response

	if e.flag&BitTime != 0 {
		before := time.Now()
		response, err = resp.client.Do(req)
		after := time.Now()
		resp.cost = after.Sub(before)
	} else {
		response, err = resp.client.Do(req)
	}

	if err != nil {
		return nil, err
	}

	for _, fn := range lastFunc {
		fn()
	}

	resp.resp = response

	if _, ok := resp.client.Transport.(*http.Transport); ok && response.Header.Get("Content-Encoding") == "gzip" && req.Header.Get("Accept-Encoding") != "" {
		var body *gzip.Reader
		body, err = gzip.NewReader(response.Body)
		if err != nil {
			return nil, err
		}
		response.Body = body
	}

	if //noinspection GoBoolExpressions
	Debug.Load() || e.debug {
		zlog.Println(resp.Dump())
	}
	return resp, nil
}

// DoWithArgs use struct args to do request
func DoWithArgs(method, rawurl string, args *RequestArgs) (*Res, error) {
	return std.DoWithArgs(method, rawurl, args)
}

// NewRequest create new request
func (e *Engine) NewRequest() *Request {
	req := requestPool.Get().(*Request)
	req.engine = e
	req.Reset()
	return req
}
