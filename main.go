package main

import (
	"os"
	"runtime"

	"github.com/urfave/cli/v3"
	"i2pgit.org/idk/reseed-tools/cmd"
)

func main() {
	// TLS 1.3 is available only on an opt-in basis in Go 1.12.
	// To enable it, set the GODEBUG environment variable (comma-separated key=value options) such that it includes "tls13=1".
	// To enable it from within the process, set the environment variable before any use of TLS:
	os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1")

	// use at most half the cpu cores
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)

	app := cli.NewApp()
	app.Name = "reseed-tools"
	app.Version = "0.2.4"
	app.Usage = "I2P tools and reseed server"
	auth := &cli.Author{
		Name:  "eyedeekay",
		Email: "hankhill19580@gmail.com",
	}
	app.Authors = append(app.Authors, auth)
	app.Flags = []cli.Flag{}
	app.Commands = []*cli.Command{
		cmd.NewReseedCommand(),
		cmd.NewSu3VerifyCommand(),
		cmd.NewKeygenCommand(),
		cmd.NewShareCommand(),
		// cmd.NewSu3VerifyPublicCommand(),
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
