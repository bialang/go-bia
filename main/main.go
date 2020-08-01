package main

import (
	"fmt"
	"os"

	"github.com/bialang/gobia"
)

func main() {
	engine, err := gobia.NewEngine()

	if err != nil {
		panic(err)
	}

	defer engine.Close()

	engine.UseBSL(os.Args)
	engine.PutFunction("foo", func(params *gobia.Parameters) interface{} {
		fmt.Println("my function")

		return 99
	})
	engine.Run([]byte(`import io; io.print(foo())`))
}
