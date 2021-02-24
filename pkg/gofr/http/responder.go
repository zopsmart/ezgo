package http

import (
	"encoding/json"
	"net/http"

	resTypes "github.com/vikash/gofr/pkg/gofr/http/response"
)

func NewResponder(w http.ResponseWriter) *Responder {
	return &Responder{w: w}
}

type Responder struct {
	w http.ResponseWriter
}

func (r Responder) Respond(data interface{}, err error) {
	r.w.Header().Set("Content-type", "application/json")

	statusCode, errorObj := r.HTTPStatusFromError(err)
	r.w.WriteHeader(statusCode)

	var resp interface{}
	switch v := data.(type) {
	case resTypes.Raw:
		resp = response{
			Data: v.Data,
			Error: errorObj,
			MetaData: v.MetaData,
		}
	default:
		resp = response{
			Data:  v,
			Error: errorObj,
		}
	}

	_ = json.NewEncoder(r.w).Encode(resp)
}

func (r Responder) HTTPStatusFromError(err error) (int, interface{}) {
	if err == nil {
		return http.StatusOK, nil
	}

	return http.StatusInternalServerError, map[string]interface{}{
		"message": err.Error(),
	}
}

type response struct {
	Error interface{} `json:"error,omitempty" xml:"error,omitempty"`
	Data  interface{} `json:"data,omitempty" xml:"data,omitempty"`
	MetaData interface{} `json:"meta,omitempty" xml:"meta,omitempty"`
}
