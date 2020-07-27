package main

import (
	"fmt"

	gobia "github.com/bialang/go-bia"
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
