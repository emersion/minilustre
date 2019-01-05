package main

import (
	"fmt"
	"os"

	"github.com/emersion/minilustre"
)

func main() {
	f, err := minilustre.Parse(os.Stdin)
	if err != nil {
		panic(err)
	}

	fmt.Println(f)
}
