package http

import (
	"net/http"

	"github.com/zopsmart/ezgo/pkg/gofr/logging"

	"github.com/zopsmart/ezgo/pkg/gofr/http/middleware"

	"github.com/gorilla/mux"
)

type Router struct {
	mux.Router
}

func NewRouter() *Router {
	muxRouter := mux.NewRouter().StrictSlash(false)
	muxRouter.Use(
		middleware.Tracer,
		middleware.Logging(logging.NewLogger(logging.INFO)),
	)

	return &Router{
		Router: *muxRouter,
	}
}

func (rou *Router) Add(method, pattern string, handler http.Handler) {
	rou.Router.NewRoute().Methods(method).Path(pattern).Handler(handler)
}
