package main

import (
	"fmt"

	"github.com/mshroyer/gcedns/internal/gcedns"
)

func main() {
	if err := gcedns.Example(); err != nil {
		fmt.Printf("Got an error: %e\n", err)
	}
	fmt.Println("Hello, world!")
}
