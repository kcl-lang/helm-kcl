package main

import (
	"os"

	"kusionstack.io/helm-kcl/cmd"
	_ "kusionstack.io/kclvm-go"
	_ "kusionstack.io/kpt-kcl-sdk/pkg/config"
)

func main() {
	if err := cmd.New().Execute(); err != nil {
		os.Exit(1)
	}
}
