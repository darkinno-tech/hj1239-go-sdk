package main

import (
	"fmt"
	"os"

	"github.com/darkinno-tech/hj1239-go-sdk/gen"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: hj1239-gen <package_path>\n")
		fmt.Fprintf(os.Stderr, "Example: hj1239-gen ./model\n")
		os.Exit(1)
	}

	pkgPath := os.Args[1]
	if err := gen.Run(pkgPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
