// Code generated from specification version 6.7.0 (f77342646af): DO NOT EDIT

package esapi

import (
	"context"
	"strings"
	"time"
)

func newGetScriptFunc(t Transport) GetScript {
	return func(id string, o ...func(*GetScriptRequest)) (*Response, error) {
		var r = GetScriptRequest{DocumentID: id}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// GetScript returns a script.
//
// See full documentation at http://www.elastic.co/guide/en/elasticsearch/reference/master/modules-scripting.html.
//
type GetScript func(id string, o ...func(*GetScriptRequest)) (*Response, error)

// GetScriptRequest configures the Get Script API request.
//
type GetScriptRequest struct {
	DocumentID string

	MasterTimeout time.Duration

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r GetScriptRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len("_scripts") + 1 + len(r.DocumentID))
	path.WriteString("/")
	path.WriteString("_scripts")
	path.WriteString("/")
	path.WriteString(r.DocumentID)

	params = make(map[string]string)

	if r.MasterTimeout != 0 {
		params["master_timeout"] = time.Duration(r.MasterTimeout * time.Millisecond).String()
	}

	if r.Pretty {
		params["pretty"] = "true"
	}

	if r.Human {
		params["human"] = "true"
	}

	if r.ErrorTrace {
		params["error_trace"] = "true"
	}

	if len(r.FilterPath) > 0 {
		params["filter_path"] = strings.Join(r.FilterPath, ",")
	}

	req, _ := newRequest(method, path.String(), nil)

	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	res, err := transport.Perform(req)
	if err != nil {
		return nil, err
	}

	response := Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	return &response, nil
}

// WithContext sets the request context.
//
func (f GetScript) WithContext(v context.Context) func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
		r.ctx = v
	}
}

// WithMasterTimeout - specify timeout for connection to master.
//
func (f GetScript) WithMasterTimeout(v time.Duration) func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
		r.MasterTimeout = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f GetScript) WithPretty() func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f GetScript) WithHuman() func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f GetScript) WithErrorTrace() func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f GetScript) WithFilterPath(v ...string) func(*GetScriptRequest) {
	return func(r *GetScriptRequest) {
		r.FilterPath = v
	}
}
