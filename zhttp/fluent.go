package zhttp

import (
	"context"
	"io"
	"net/http"
	"strings"
)

// NewRequest create new request
func NewRequest() *Request {
	return std.NewRequest()
}

// GetRequest get request from pool
func (e *Engine) GetRequest() *Request {
	return e.NewRequest()
}

// GetRequest get request from pool
func GetRequest() *Request {
	return std.GetRequest()
}

// URL set request url
func (r *Request) URL(url string) *Request {
	r.url = url
	return r
}

// Method set request method
func (r *Request) Method(method string) *Request {
	r.method = strings.ToUpper(method)
	return r
}

// Header set request header
func (r *Request) Header(key, value string) *Request {
	r.headers[key] = value
	return r
}

// Headers set request headers
func (r *Request) Headers(headers Header) *Request {
	for k, v := range headers {
		r.headers[k] = v
	}
	return r
}

// Query set query param
func (r *Request) Query(key string, value interface{}) *Request {
	r.queryParams[key] = value
	return r
}

// QueryMap set query params
func (r *Request) QueryMap(params QueryParam) *Request {
	for k, v := range params {
		r.queryParams[k] = v
	}
	return r
}

// Form set form param
func (r *Request) Form(key string, value interface{}) *Request {
	r.formParams[key] = value
	return r
}

// FormMap set form params
func (r *Request) FormMap(params Param) *Request {
	for k, v := range params {
		r.formParams[k] = v
	}
	return r
}

// Body set request body
func (r *Request) Body(body interface{}) *Request {
	r.body = body
	return r
}

// JSON set json request body
func (r *Request) JSON(v interface{}) *Request {
	r.body = &bodyJson{v}
	return r
}

// XML set xml request body
func (r *Request) XML(v interface{}) *Request {
	r.body = &bodyXml{v}
	return r
}

// File set file upload
func (r *Request) File(fieldName, fileName string, file io.ReadCloser) *Request {
	r.uploads = append(r.uploads, FileUpload{
		FieldName: fieldName,
		FileName:  fileName,
		File:      file,
	})
	return r
}

// Client set http client
func (r *Request) Client(client *http.Client) *Request {
	r.client = client
	return r
}

// Cookie set cookie
func (r *Request) Cookie(cookie *http.Cookie) *Request {
	r.cookies = append(r.cookies, cookie)
	return r
}

// Context set context
func (r *Request) Context(ctx context.Context) *Request {
	r.ctx = ctx
	return r
}

// Host set host
func (r *Request) Host(host string) *Request {
	r.host = host
	return r
}

// UploadProgress set upload progress callback
func (r *Request) UploadProgress(progress UploadProgress) *Request {
	r.uploadProg = progress
	return r
}

// DownloadProgress set download progress callback
func (r *Request) DownloadProgress(progress DownloadProgress) *Request {
	r.downloadProg = progress
	return r
}

// NoRedirect disable redirect
func (r *Request) NoRedirect(disable bool) *Request {
	r.noRedirect = disable
	return r
}

// Custom custom request handler
func (r *Request) Custom(fn CustomReq) *Request {
	r.customReq = fn
	return r
}

// Do do request
func (r *Request) Do() (*Res, error) {
	if r.url == "" {
		return nil, ErrUrlNotSpecified
	}

	args := &RequestArgs{
		Headers:      r.headers,
		QueryParams:  r.queryParams,
		FormParams:   r.formParams,
		Body:         r.body,
		Uploads:      r.uploads,
		Client:       r.client,
		Cookies:      r.cookies,
		Ctx:          r.ctx,
		Host:         r.host,
		UploadProg:   r.uploadProg,
		DownloadProg: r.downloadProg,
		NoRedirect:   r.noRedirect,
		CustomReq:    r.customReq,
	}

	method := r.method
	if method == "" {
		method = "GET"
	}

	return r.engine.DoWithArgs(method, r.url, args)
}

// GET do get request
func (r *Request) GET() (*Res, error) {
	r.method = "GET"
	return r.Do()
}

// POST do post request
func (r *Request) POST() (*Res, error) {
	r.method = "POST"
	return r.Do()
}

// PUT do put request
func (r *Request) PUT() (*Res, error) {
	r.method = "PUT"
	return r.Do()
}

// PATCH do patch request
func (r *Request) PATCH() (*Res, error) {
	r.method = "PATCH"
	return r.Do()
}

// DELETE do delete request
func (r *Request) DELETE() (*Res, error) {
	r.method = "DELETE"
	return r.Do()
}

// HEAD do head request
func (r *Request) HEAD() (*Res, error) {
	r.method = "HEAD"
	return r.Do()
}

// OPTIONS do options request
func (r *Request) OPTIONS() (*Res, error) {
	r.method = "OPTIONS"
	return r.Do()
}

// Reset reset request
func (r *Request) Reset() *Request {
	r.url = ""
	r.method = ""

	if len(r.headers) > 0 {
		r.headers = make(Header, 8)
	}
	if len(r.queryParams) > 0 {
		r.queryParams = make(QueryParam, 8)
	}
	if len(r.formParams) > 0 {
		r.formParams = make(Param, 8)
	}

	r.body = nil
	r.uploads = r.uploads[:0]
	r.cookies = r.cookies[:0]
	r.ctx = nil
	r.host = ""
	r.uploadProg = nil
	r.downloadProg = nil
	r.noRedirect = false
	r.customReq = nil

	return r
}

// Release release request to pool
func (r *Request) Release() {
	r.Reset()
	requestPool.Put(r)
}

// DoAndRelease do request and release to pool
func (r *Request) DoAndRelease() (*Res, error) {
	result, err := r.Do()
	r.Release()
	return result, err
}

// GetAndRelease do get request and release to pool
func (r *Request) GetAndRelease() (*Res, error) {
	result, err := r.GET()
	r.Release()
	return result, err
}

// PostAndRelease do post request and release to pool
func (r *Request) PostAndRelease() (*Res, error) {
	result, err := r.POST()
	r.Release()
	return result, err
}
