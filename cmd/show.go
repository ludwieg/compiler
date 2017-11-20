package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/disiqueira/gotree"
	"github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/ludwieg/compiler/models"
)

func printField(tree *gotree.GTStructure, index int, f models.Field) {
	name := aurora.Gray(fmt.Sprintf("[%d] ", index)).String()
	if f.Type.Source == models.SourceNative {
		name += aurora.Cyan(f.Type.NativeType).String()
	} else {
		name += fmt.Sprintf("@%s", aurora.Green(f.Type.CustomType))
	}
	if f.ObjectType == models.ObjectTypeArray {
		name += fmt.Sprintf("[%s]", aurora.Gray(f.Size))
	}
	name += fmt.Sprintf(" %s", f.Name)
	if f.HasAttribute(models.AttributeDeprecated) {
		name += fmt.Sprintf(" %s", aurora.Red("[Deprecated]"))
	}

	var item gotree.GTStructure
	item.Name = name
	tree.Items = append(tree.Items, item)
}

func printStructure(tree *gotree.GTStructure, s models.Struct) {
	var item gotree.GTStructure
	item.Name = fmt.Sprintf("%s", aurora.Green(s.Name))
	structsCount := len(s.Structs)

	var fTree gotree.GTStructure
	fTree.Name = "Fields"
	cursor := 1
	for _, field := range s.Fields {
		printField(&fTree, cursor-1, field)
		cursor++
	}
	item.Items = append(item.Items, fTree)

	if structsCount > 0 {
		var sTree gotree.GTStructure
		sTree.Name = "Structures"
		for _, str := range s.Structs {
			printStructure(&sTree, str)
		}
		item.Items = append(item.Items, sTree)
	}
	tree.Items = append(tree.Items, item)
}

var Show = cli.Command{
	Name:    "show",
	Aliases: []string{"s"},
	Usage:   "Shows the structure of Ludwieg packages",
	Action: func(c *cli.Context) error {
		input := c.Args().First()
		toProcess := []string{}
		if input == "" {
			log.Errorf("Please specify project path. ludgo show <path>")
			return nil
		}
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

		fmt.Println()

		for _, pkg := range allPackages {
			var tree gotree.GTStructure
			tree.Name = fmt.Sprintf("%s (%s)", aurora.Bold(pkg.Name), aurora.Gray(pkg.Identifier))

			fieldsCount := len(pkg.Fields)
			structsCount := len(pkg.Structs)
			cursor := 1

			if fieldsCount > 0 {
				var fTree gotree.GTStructure
				fTree.Name = "Fields"
				for _, field := range pkg.Fields {
					printField(&fTree, cursor-1, field)
					cursor++
				}
				tree.Items = append(tree.Items, fTree)
			}

			if structsCount > 0 {
				var sTree gotree.GTStructure
				sTree.Name = "Structures"
				for _, str := range pkg.Structs {
					printStructure(&sTree, str)
				}
				tree.Items = append(tree.Items, sTree)
			}
			if fieldsCount == 0 && structsCount == 0 {
				var empty gotree.GTStructure
				empty.Name = aurora.Gray("(Empty Package)").String()
				tree.Items = append(tree.Items, empty)
			}

			gotree.PrintTree(tree)
			fmt.Println()
		}

		return nil
	},
}
