package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version identifier populated via the CI/CD process.
var Version = "HEAD"

// NewVersionCmd returns the version command.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version of the helm kcl plugin",
		Run: func(*cobra.Command, []string) {
			fmt.Println(Version)
		},
		SilenceUsage: true,
	}
}
