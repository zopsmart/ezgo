package main

import (
	"github.com/zopsmart/ezgo/examples/grpc-server/grpc"
	"github.com/zopsmart/ezgo/pkg/gofr"
)

func main() {
	app := gofr.New()

	grpc.RegisterHelloServer(app, grpc.Server{})

	app.Run()
}
