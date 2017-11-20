package langs

import (
	"encoding/base64"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"

	"github.com/ludwieg/compiler/models"
)

type Go struct {
	pkgName string
	out     string
}

func (c Go) Compile(in, out, pkgName, prefix string, packages *models.PackageList) {
	log.Infof("Initialising %s compiler", aurora.Blue("Golang"))
	if prefix != "" {
		log.Warn("Ignoring unecessary --prefix option.")
		prefix = ""
	}
	if pkgName == "" {
		// This means that no package name was provided. So we can assume
		// it using `out` basename.
		r := regexp.MustCompile("[^a-z]")
		pkgName = string(r.ReplaceAll([]byte(strings.ToLower(filepath.Base(out))), []byte{}))
		log.Warnf("No package name was provided. Assumed %s based on output path.", aurora.Magenta(pkgName))
		log.Warn("Please use the --package argument to define a custom package name")
	}
	c.pkgName = pkgName
	c.out = out

	if err := c.writeBaseFile(); err != nil {
		log.Errorf("Operation failed: %s", err)
		os.Exit(1)
	}

	for _, p := range *packages {
		c.writePackage(&p)
	}

	c.writeInitializer(packages)
	log.Info("Succeeded")
}

func (c Go) writeInitializer(pkgs *models.PackageList) {
	result := []string{}
	for _, p := range *pkgs {
		result = append(result, convertToPascalCase(p.Name)+"{}")
	}
	separator := ","
	if len(result) > 1 {
		separator = ",\n"
	}

	c.output(c.pkgName, processTemplate("golangInitializer", goInitializer, templateData{
		"pkg":      c.pkgName,
		"packages": strings.Join(result, separator),
	}))
}

func (c Go) output(name string, contents []byte) {
	log.Infof("Writing %s/%s", aurora.Magenta(filepath.Base(c.out)), aurora.Magenta(name+".go"))
	contents, err := format.Source(contents)
	if err != nil {
		log.Errorf("Error formatting output source file: %s", err)
		log.Errorf("This is a serious bug. Please report it and attach the input lud files.")
		os.Exit(1)
	}
	err = ioutil.WriteFile(filepath.Join(c.out, name+".go"), contents, 0644)
	if err != nil {
		log.Errorf("Operation failed: %s", err)
		os.Exit(1)
	}
}

func (c Go) writePackage(p *models.Package) {
	if len(p.Fields) == 0 {
		c.writeEmptyPackage(p)
	} else {
		c.writeNormalPackage(p)
	}
}

func (c Go) writeEmptyPackage(p *models.Package) {
	c.output(p.Name, processTemplate("emptyPackage", goEmptyPackage, templateData{
		"pkg":  c.pkgName,
		"id":   p.Identifier,
		"name": convertToPascalCase(p.Name),
	}))
}

func (c Go) writeNormalPackage(p *models.Package) {
	pkgName := convertToPascalCase(p.Name)

	c.output(p.Name, processTemplate("package", goPackage, templateData{
		"pkg":         c.pkgName,
		"id":          p.Identifier,
		"name":        pkgName,
		"fields":      c.generateFields(p.Fields, pkgName),
		"annotations": c.generateAnnotations(p.Fields, pkgName),
		"structures":  c.generateStructs(p.Structs, pkgName),
	}))
}

func (c Go) generateStructs(sArr []models.Struct, pkgName string) string {
	var structs []byte
	for _, s := range sArr {
		structs = append(structs, c.writeStruct(&s, pkgName)...)
	}
	return string(structs)
}

func (c Go) generateFields(fArr []models.Field, pkgName string) string {
	var fields []byte
	for _, f := range fArr {
		fields = append(fields, c.writeField(&f, pkgName)...)
	}
	return string(fields)
}

func (c Go) generateAnnotations(fArr []models.Field, pkgName string) string {
	var annotationsArr []string
	for _, f := range fArr {
		annotationsArr = append(annotationsArr, string(c.writeAnnotation(&f, pkgName)))
	}

	annotations := strings.Join(annotationsArr, ",")
	if len(annotationsArr) > 1 {
		annotations = annotations + ","
		annotations = strings.Replace(annotations, ",", ",\n", -1)
	}
	return annotations
}

func (c Go) writeStruct(s *models.Struct, prefix string) []byte {
	pkgName := prefix + convertToPascalCase(s.Name)
	return processTemplate("struct", goStruct, templateData{
		"name":        pkgName,
		"fields":      c.generateFields(s.Fields, pkgName),
		"annotations": c.generateAnnotations(s.Fields, pkgName),
		"structures":  c.generateStructs(s.Structs, pkgName),
	})
}

func (c Go) writeAnnotation(f *models.Field, prefix string) []byte {
	isArray := f.Size != ""
	if f.Type.Source == "native" {
		val := []string{}
		if isArray {
			val = append(val, "Type: TypeArray", `ArraySize: "`+f.Size+`"`)
			val = append(val, "ArrayType: Type"+convertToPascalCase(string(f.Type.NativeType)))
		} else {
			val = append(val, "Type: Type"+convertToPascalCase(string(f.Type.NativeType)))
		}
		return []byte("{" + strings.Join(val, ",") + "}")
	}

	// At this point, it's a non-native type, but maybe an array.
	if isArray {
		basicType := prefix + convertToPascalCase(f.Type.CustomType) + "{}"
		var val string
		if f.Size == "*" {
			val = "ArrayOf(" + basicType + ")"
		} else {
			val = "ArrayOfWithSize(" + basicType + ", " + f.Size + ")"
		}
		return []byte(val)
	}
	return []byte("{Type: TypeStruct }")
}

func (c Go) writeField(f *models.Field, prefix string) []byte {
	isArray := f.Size != ""
	var t string
	n := convertToPascalCase(f.Name)
	switch f.Type.Source {
	case "native":
		if f.Type.NativeType == "blob" {
			t = "[]byte"
			if isArray {
				t = "[]" + t
			}
		} else {
			t = "*Ludwieg" + convertToPascalCase(string(f.Type.NativeType))
			if isArray {
				t = "[](" + t + ")"
			}
		}
	case "user":
		// User-based types are always relative to the current package. Given
		// our naming rules, we can just append the name of the struct to the
		// current package name without needing to resort to any name-resolution
		// techniniques.
		t = "*" + prefix + convertToPascalCase(f.Type.CustomType)
		if isArray {
			t = "[](" + t + ")"
		}
	}
	return processTemplate("field", goField, templateData{
		"name": n,
		"type": t,
	})
}

func (c Go) writeBaseFile() error {
	base, err := base64.StdEncoding.DecodeString(golangBaseFile)
	if err != nil {
		return err
	}
	base = append([]byte("// WARNING: Automatically generated by ludco. DO NOT EDIT.\n\npackage "+c.pkgName+"\n"), base...)
	c.output("ludwieg_base", base)
	return nil
}
