package cmd

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/ludwieg/ludco/langs"
)

var Compile = cli.Command{
	Name:    "compile",
	Aliases: []string{"c"},
	Usage:   "Compiles a Ludwieg project",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "lang",
			Usage: "Destination language. Currently supported languages are go, objc, and java",
		},
		cli.StringFlag{
			Name:  "package",
			Usage: "Package name to use when generating files. Required when compiling to Java or Go. When omitted, ludco uses the input folder's name",
		},
		cli.StringFlag{
			Name:  "prefix",
			Usage: "Class prefix used by the Objective-C compiler",
		},
	},
	Action: func(c *cli.Context) error {
		lang := strings.ToLower(c.String("lang"))
		if c.NArg() != 2 {
			log.Errorf("Please specify input and output paths. ludgo compile --lang <lang> <input> <output>")
			return nil
		}

		input, err := filepath.Abs(c.Args()[0])
		if err != nil {
			log.Errorf("Error reading input path: %s", err)
			return nil
		}
		output, err := filepath.Abs(c.Args()[1])
		if err != nil {
			log.Errorf("Error reading output path: %s", err)
			return nil
		}

		if lang == "" {
			log.Errorf("Error: You must define which language must be used as output")
			return nil
		}
		if lang != "objc" && lang != "go" && lang != "java" {
			log.Errorf("Error: Supported languages are go, objc and java")
			return nil
		}

		if input == "" {
			log.Errorf("Please provide input files.")
			return nil
		}

		if output == "" {
			log.Errorf("Please provide a path for output files. Path must point to a directory, that will be created in case it does not already exist.")
			return nil
		}

		if stat, err := os.Stat(output); err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(output, 0700)
				if err != nil {
					log.Errorf("Error: %s", err)
					return nil
				}
			} else {
				log.Errorf("Error: %s", err)
				return nil
			}
		} else {
			if !stat.IsDir() {
				log.Errorf("%s already exists and is not a directory.", output)
			}
		}

		toProcess := []string{}
		if stat, err := os.Stat(input); err != nil {
			if os.IsNotExist(err) {
				log.Errorf("Input path %s does not exist.", input)
				return nil
			}
		} else {
			if !stat.IsDir() {
				log.Errorf("Input path is not a directory.")
				return nil
			}

			glob, err := filepath.Glob(input + "/*.lud")
			if err != nil {
				log.Errorf("Error enumerating files: %s", err)
				return nil
			}
			toProcess = glob
		}

		allPackages := ProcessFiles(toProcess)
		if allPackages == nil {
			return nil
		}

		var compiler langs.Compiler

		switch lang {
		case "go":
			compiler = langs.Go{}
		case "objc":
			compiler = langs.ObjC{}
		case "java":
			compiler = langs.Java{}
		}

		compiler.Compile(input, output, c.String("package"), c.String("prefix"), &allPackages)
		return nil
	},
}
