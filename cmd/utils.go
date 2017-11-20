package cmd

import (
	"os"
	"path/filepath"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/ludwieg/compiler/models"
	"github.com/ludwieg/compiler/parser"
	"github.com/ludwieg/compiler/validation"
)

// ProcessFiles attempts to load all files provided on the input array and
// returns an array of `models.Package` objects ready for use. In case of error,
// those are printed to the stdout and nil is returned
func ProcessFiles(toProcess []string) models.PackageList {
	allPackages := models.PackageList{}
	for _, p := range toProcess {
		f := filepath.Base(p)
		logger := log.WithField("file", f)
		fd, err := os.Open(p)
		if err != nil {
			log.Errorf("Error reading: %s", err)
			return nil
		}

		out, err := parser.ParseReader("", fd)
		fd.Close()
		if err != nil {
			logger.Errorf("Error parsing %s: %s", p, err)
			return nil
		}
		result := out.([]interface{})

		for _, subject := range result {
			if pkg, ok := subject.(parser.Package); ok {
				errs := validation.Validate(pkg)
				if len(errs) > 0 {
					logger.Warn("Found problems validating package:")
					for _, err := range errs {
						logger.Errorf("error: %s", err)
					}
					return nil
				}
				defer func() {
					if err := recover(); err != nil {
						logger.Errorf("Unexpected error analysing %s (%s): %s", pkg.Name, f, err)
					}
				}()
				allPackages = append(allPackages, *models.ConvertASTPackage(pkg))
			}
		}
	}

	packages := map[string]bool{}
	for _, p := range allPackages {
		if _, exists := packages[p.Name]; exists {
			log.Warn("Found problems reading source files:")
			log.Errorf("error: duplicated package definition %s", p.Name)
			return nil
		}
		packages[p.Name] = true
	}

	sort.Sort(allPackages)

	return allPackages
}
