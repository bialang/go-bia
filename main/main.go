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

	i := 0

	engine.UseBSL(os.Args)
	engine.Put("hey", "ho")
	engine.PutFunction("foo", func(params *gobia.Parameters) interface{} {
		if m, e := params.Get("as"); e == nil {
			var s string
			m.Cast(&s)
			fmt.Printf("%s? this is what i get?\n", s)
		}

		fmt.Println("my function")

		i++

		return i
	})
	engine.Run([]byte(`import io; io.print(foo(as="hi")); io.print(foo()); io.print(hey)`))
}
