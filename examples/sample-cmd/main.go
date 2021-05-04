package main

import (
	"fmt"

	"github.com/zopsmart/ezgo/pkg/gofr"
)

func main() {
	app := gofr.NewCMD()

	app.SubCommand("hello", func(c *gofr.Context) (interface{}, error) {
		return "Hello World!", nil
	})

	app.SubCommand("params", func(c *gofr.Context) (interface{}, error) {
		return fmt.Sprintf("Hello %s!", c.Param("name")), nil
	})

	app.Run()

}
