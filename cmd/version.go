package cmd

import (
	"fmt"

	"github.com/urfave/cli/v3"
	"i2pgit.org/idk/reseed-tools/reseed"
)

func NewVersionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print the version number of reseed-tools",
		Action: func(c *cli.Context) error {
			fmt.Printf("%s\n", reseed.Version)
			return nil
		},
	}
}
