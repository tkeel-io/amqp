package main

import (
	"fmt"
	"os"
)

func main() {
	ParseFlags(Server)
	if err := Server.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
