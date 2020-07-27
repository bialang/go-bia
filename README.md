# gobia

A CGo wrapper for Bia

## Usage

```go
package main

import (
	"fmt"

	"github.com/bialang/gobia"
)

func main() {
	engine, err := gobia.NewEngine()
  
	// don't forget to close the engine after usage!
	defer engine.Close()

	engine.UseBSL()
	engine.PutFunction("hello_world", func(params *gobia.Parameters) {
		fmt.Println("Hello, World! - Go")
	})
	engine.Run([]byte(`

	  import io

	  io.print("Hello, World! - Bia")
	  hello_world()

	`))
}
```

## Installation

### Requirements

- Bia
- Go *>= 1.12*
- GCC

### Downloading

```sh
go get github.com/bialang/gobia
```
