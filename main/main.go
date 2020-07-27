package main

import (
	"fmt"

	"github.com/bialang/gobia"
)

func main() {
	engine, err := gobia.NewEngine()

	if err != nil {
		panic(err)
	}

	defer engine.Close()

	engine.UseBSL()
	engine.PutFunction("foo", func(params *gobia.Parameters) {
		fmt.Println("my function")
	})
	engine.Run([]byte(`foo()`))
}
