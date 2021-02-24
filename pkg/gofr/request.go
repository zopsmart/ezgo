package gofr

import (
	"context"
	"mime/multipart"
	"net/http"
)

type Request interface {
	Context() context.Context
	Param(string) string
	PathParam(string) string
	Bind(interface{}) error
	Header() http.Header
	FormFile(string) (multipart.File, *multipart.FileHeader, error)
}
