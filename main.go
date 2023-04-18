package main

import (
	"os"

	_ "github.com/KusionStack/krm-kcl/pkg/config"
	"kusionstack.io/helm-kcl/cmd"
	_ "kusionstack.io/kclvm-go"
)

func main() {
	if err := cmd.New().Execute(); err != nil {
		os.Exit(1)
	}
}
