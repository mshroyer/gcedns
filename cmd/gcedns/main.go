package main

import (
	"context"
	"fmt"

	"github.com/mshroyer/gcedns/internal/gcedns"
)

func main() {
	ctx := context.Background()

	info, err := gcedns.GetHostVMInfo(ctx, "foo")
	if err != nil {
		fmt.Printf("error getting VM info: %e\n", err)
		return
	}

	fmt.Printf("Got VM info: %+v\n", info)
}
