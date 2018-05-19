package main

import (
	"os"

	"github.com/urfave/cli"

	"github.com/ludwieg/ludco/cmd"
	"fmt"
)

var build_date string

func main() {
	app := cli.NewApp()
	app.Name = "ludco"
	app.Usage = "Ludwieg compiler"
	app.Version = fmt.Sprintf("0.1.0 (%s)", build_date)
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
