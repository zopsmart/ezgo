package gofr

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/zopsmart/ezgo/pkg/gofr/config"
	gofrHTTP "github.com/zopsmart/ezgo/pkg/gofr/http"
	"github.com/zopsmart/ezgo/pkg/gofr/logging"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

// App is the main application in the gofr framework.
type App struct {
	// Config can be used by applications to fetch custom configurations from environment or file.
	Config Config // If we directly embed, unnecessary confusion between app.Get and app.GET will happen.

	grpcServer *grpcServer
	HttpServer *httpServer

	cmd *cmd

	// container is unexported because this is an internal implementation and applications are provided access to it via Context
	container *Container

	grpcRegistered bool
	httpRegistered bool
}

// RegisterService adds a grpc service to the gofr application.
func (a *App) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	a.grpcRegistered = true
	a.container.Logger.Infof("Registering GRPC Server: %s", desc.ServiceName)
	a.grpcServer.server.RegisterService(desc, impl)
}

// New creates a HTTP Server Application and returns that App.
func New() *App {
	app := &App{}
	app.readConfig()
	app.container = newContainer(app.Config)

	app.initTracer()

	// HTTP Server
	port, err := strconv.Atoi(app.Config.Get("HTTP_PORT"))
	if err != nil || port <= 0 {
		port = defaultHTTPPort
	}

	app.HttpServer = &httpServer{
		Router: gofrHTTP.NewRouter(),
		port:   port,
	}

	// GRPC Server
	port, err = strconv.Atoi(app.Config.Get("GRPC_PORT"))
	if err != nil || port <= 0 {
		port = defaultGRPCPort
	}

	app.grpcServer = newGRPCServer(app.container, port)

	return app
}

// NewCMD creates a command line application.
func NewCMD() *App {
	app := &App{}
	app.readConfig()

	app.container = newContainer(app.Config)
	app.cmd = &cmd{}
	app.container.Logger = logging.NewSilentLogger() // TODO - figure out a proper way to log in CMD

	app.initTracer()

	return app
}

// Run starts the application. If it is a HTTP server, it will start the server.
func (a *App) Run() {
	if a.cmd != nil {
		a.cmd.Run(a.container)
	}

	wg := sync.WaitGroup{}

	// Start HTTP Server
	if a.httpRegistered {
		wg.Add(1)

		// Add Default routes
		a.add(http.MethodGet, "/.well-known/health", healthHandler)
		a.add(http.MethodGet, "/favicon.ico", faviconHandler)

		go func(s *httpServer) {
			defer wg.Done()
			s.Run(a.container)
		}(a.HttpServer)
	}

	// Start GRPC Server only if a service is registered
	if a.grpcRegistered {
		wg.Add(1)

		go func(s *grpcServer) {
			defer wg.Done()
			s.Run(a.container)
		}(a.grpcServer)
	}

	wg.Wait()
}

// readConfig reads the configuration from the default location.
func (a *App) readConfig() {
	var configLocation string
	if _, err := os.Stat("./configs"); err == nil {
		configLocation = "./configs"
	}

	a.Config = config.NewEnvFile(configLocation)
}

// GET adds a Handler for http GET method for a route pattern.
func (a *App) GET(pattern string, handler Handler) {
	a.add("GET", pattern, handler)
}

// PUT adds a Handler for http PUT method for a route pattern.
func (a *App) PUT(pattern string, handler Handler) {
	a.add("PUT", pattern, handler)
}

// POST adds a Handler for http POST method for a route pattern.
func (a *App) POST(pattern string, handler Handler) {
	a.add("POST", pattern, handler)
}

// DELETE adds a Handler for http DELETE method for a route pattern.
func (a *App) DELETE(pattern string, handler Handler) {
	a.add("DELETE", pattern, handler)
}

func (a *App) add(method, pattern string, h Handler) {
	a.httpRegistered = true
	a.HttpServer.Router.Add(method, pattern, handler{
		function:  h,
		container: a.container,
	})
}

// SubCommand adds a sub-command to the CLI application.
// Can be used to create commands like "kubectl get" or "kubectl get ingress".
func (a *App) SubCommand(pattern string, handler Handler) {
	a.cmd.addRoute(pattern, handler)
}

func (a *App) initTracer() {
	// If zipkin is running on default port - start tracing
	// TODO - remove gomnd and do proper check here.
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", "9411"), 2*time.Second) //nolint:gomnd // 2 is reasonable to check.

	if err != nil || conn == nil {
		a.container.Logf("Tracer detection error: %v", err)
		a.container.Log("To enable tracing locally, Run: docker run -d -p 9411:9411 openzipkin/zipkin")

		return
	}

	a.container.Log("Exporting traces to zipkin.")

	err = zipkin.InstallNewPipeline(
		"http://localhost:9411/api/v2/spans",
		"gofr-App",
		zipkin.WithSDK(&trace.Config{DefaultSampler: trace.AlwaysSample()}),
	)

	if err != nil {
		a.container.Error(err)
	}
}
