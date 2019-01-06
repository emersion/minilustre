package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/llir/llvm/ir"

	"github.com/emersion/minilustre"
)

var (
	noop = flag.Bool("n", false, "don't compile, just print AST")
)

func main() {
	flag.Parse()

	f, err := minilustre.Parse(os.Stdin)
	if err != nil {
		panic(err)
	}

	if *noop {
		fmt.Println(f)
		return
	}

	m := ir.NewModule()
	if err := minilustre.Compile(f, m); err != nil {
		panic(err)
	}

	fmt.Println(m)
}
