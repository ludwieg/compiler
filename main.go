package main

import (
	"os"

	"github.com/ludwieg/compiler/cmd"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "ludco"
	app.Usage = "Ludwieg compiler"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		cmd.Compile,
		cmd.Show,
	}

	app.Action = func(c *cli.Context) error {
		cli.ShowAppHelpAndExit(c, 1)
		return nil
	}

	app.Run(os.Args)
}
