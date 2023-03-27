package main

import (
	"os"

	"kusionstack.io/helm-kcl/cmd"
)

func main() {
	if err := cmd.New().Execute(); err != nil {
		os.Exit(1)
	}
}
