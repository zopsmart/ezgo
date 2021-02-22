package http

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"mime/multipart"
	"net/http"

	"github.com/gorilla/mux"
)

// Request is an abstraction over the underlying http.Request. This abstraction is useful because it allows us
// to create applications without being aware of the transport. cmd.Request is another such abstraction.
type Request struct {
	req        *http.Request
	pathParams map[string]string
	formData   map[string]interface{}
	filesData  map[string]*multipart.FileHeader
}

func NewRequest(r *http.Request) *Request {
	return &Request{
		req:        r,
		pathParams: mux.Vars(r),
		formData:   make(map[string]interface{}),
		filesData:  make(map[string]*multipart.FileHeader),
	}
}

func (r *Request) Param(key string) string {
	return r.req.URL.Query().Get(key)
}

func (r *Request) Context() context.Context {
	return r.req.Context()
}

func (r *Request) PathParam(key string) string {
	return r.pathParams[key]
}

func (r *Request) Bind(i interface{}) error {
	body, err := r.body()
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &i)
}

func (r *Request) body() ([]byte, error) {
	bodyBytes, err := ioutil.ReadAll(r.req.Body)
	if err != nil {
		return nil, err
	}

	r.req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes, nil
}

func (r *Request) GetMultipartFormData(maxMemory int64, input interface{}) error {
	err := r.req.ParseMultipartForm(maxMemory)
	if err != nil {
		return err
	}

	for key, val := range r.req.MultipartForm.Value {
		r.formData[key] = val[0]
	}

	for fieldName, file := range r.req.MultipartForm.File {
		r.filesData[fieldName] = file[0]
	}

	err = mapstructure.Decode(r.formData, &input)
	if err != nil {
		return err
	}

	err = mapstructure.Decode(r.filesData, &input)
	if err != nil {
		return err
	}

	return nil
}
