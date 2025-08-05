package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"i2pgit.org/idk/reseed-tools/reseed"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of reseed-tools",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", reseed.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
