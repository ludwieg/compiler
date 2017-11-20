package langs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/logrusorgru/aurora"
	"github.com/ludwieg/compiler/models"
	log "github.com/sirupsen/logrus"
)

type ObjC struct {
	prefix string
	out    string
}

func (c ObjC) Compile(in, out, pkgName, prefix string, packages *models.PackageList) {
	log.Infof("Initialising %s compiler", aurora.Blue("objc"))
	if pkgName != "" {
		log.Warn("Ignoring unecessary --pacakge option")
	}
	if prefix == "" {
		log.Warn("You did not specify a prefix using the --prefix option. Although optional, its use is advised.")
		log.Warn(`Quoting Apple documentation: 
	Your own classes should use three letter prefixes. These might relate to a 
	combination of your company name and your app name, or even a specific 
	component within your app. As an example, if your company were called 
	Whispering Oak, and you were developing a game called Zebra Surprise, you 
	might choose WZS or WOZ as your class prefix."`)
	}

	c.prefix = strings.ToUpper(prefix)
	c.out = out

	for _, p := range *packages {
		c.writePackage(&p)
	}
	log.Info("Succeeded")
	fmt.Println()
	fmt.Println(c.integrationInstructions(packages))
}

func (c ObjC) output(name string, contents []byte) {
	log.Infof("Writing %s/%s", aurora.Magenta(filepath.Base(c.out)), aurora.Magenta(name))
	err := ioutil.WriteFile(filepath.Join(c.out, name), contents, 0644)
	if err != nil {
		log.Errorf("Operation failed: %s", err)
		os.Exit(1)
	}
}

func (c ObjC) writePackage(p *models.Package) {
	if len(p.Fields) == 0 {
		c.writeEmptyPackage(p)
	} else {
		c.writeNormalPackage(p)
	}
}

func (c ObjC) writeEmptyPackage(p *models.Package) {
	c.output(p.Name+".h", processTemplate("emptyPackageHeader", objcEmptyPackageHeader, templateData{
		"prefix": c.prefix,
		"name":   convertToPascalCase(p.Name),
	}))

	c.output(p.Name+".m", processTemplate("emptyPackageImplementation", objcEmptyPackageImplementation, templateData{
		"prefix": c.prefix,
		"id":     p.Identifier,
		"name":   convertToPascalCase(p.Name),
	}))
}

func (c ObjC) writeNormalPackage(p *models.Package) {
	pkgName := convertToPascalCase(p.Name)

	c.output(p.Name+".h", processTemplate("objcPackageHeader", objcPackageHeader, templateData{
		"prefix":     c.prefix,
		"name":       pkgName,
		"fields":     c.generateFields(p.Fields, pkgName),
		"structures": c.generateStructsHeaders(p.Structs, pkgName),
	}))

	c.output(p.Name+".m", processTemplate("objcPackageImplementation", objcPackageImplementation, templateData{
		"prefix":      c.prefix,
		"id":          p.Identifier,
		"name":        pkgName,
		"annotations": c.generateAnnotations(p.Fields, pkgName),
		"structures":  c.generateStructsImplementation(p.Structs, pkgName),
	}))
}

func (c ObjC) generateStructsHeaders(sArr []models.Struct, prefix string) string {
	var val []byte
	for _, s := range sArr {
		val = append(val, c.writeStructHeaders(&s, prefix)...)
	}
	return string(val)
}

func (c ObjC) generateStructsImplementation(sArr []models.Struct, pkgName string) string {
	var structs []byte
	for _, s := range sArr {
		structs = append(structs, c.writeStructsImplementation(&s, pkgName)...)
	}
	return string(structs)
}

func (c ObjC) generateFields(fArr []models.Field, pkgName string) string {
	var fields []byte
	for _, f := range fArr {
		fields = append(fields, c.writeField(&f, pkgName)...)
	}
	return string(fields)
}

func (c ObjC) generateAnnotations(fArr []models.Field, pkgName string) string {
	var annotationsArr []string
	for _, f := range fArr {
		annotationsArr = append(annotationsArr, string(c.writeAnnotation(&f, pkgName)))
	}

	annotations := strings.Join(annotationsArr, "\n")
	return annotations
}

func (c ObjC) writeStructsImplementation(s *models.Struct, prefix string) []byte {
	pkgName := c.prefix + prefix + convertToPascalCase(s.Name)

	return processTemplate("objcStructImplementation", objcStructImplementation, templateData{
		"name":        pkgName,
		"annotations": c.generateAnnotations(s.Fields, pkgName),
		"structures":  c.generateStructsImplementation(s.Structs, pkgName),
	})

}

func (c ObjC) writeStructHeaders(s *models.Struct, prefix string) []byte {
	pkgName := c.prefix + prefix + convertToPascalCase(s.Name)

	return processTemplate("objcStruct", objcStructHeader, templateData{
		"name":       pkgName,
		"fields":     c.generateFields(s.Fields, pkgName),
		"structures": c.generateStructsHeaders(s.Structs, pkgName),
	})
}

func (c ObjC) writeAnnotation(f *models.Field, prefix string) []byte {
	isArray := f.Size != ""
	var v, kind string
	name := convertToCamelCase(f.Name)
	if f.Type.Source == "native" {
		kind = convertToPascalCase(string(f.Type.NativeType))
		if isArray {
			v = "[LUDTypeAnnotation arrayAnnotationWithName:@\"" + name + "\" type:LUDProtocolType" + kind + " andArraySize:@\"" + f.Size + "\"],"
		} else {
			v = "[LUDTypeAnnotation annotationWithName:@\"" + name + "\" type:LUDProtocolType" + kind + "],"
		}
	} else {
		kind = c.prefix + prefix + convertToPascalCase(string(f.Type.CustomType))
		if isArray {
			v = "[LUDTypeAnnotation arrayAnnotationWithName:@\"" + name + "\" userType:[" + kind + " class] andArraySize:@\"" + f.Size + "\"],"
		} else {
			v = "[LUDTypeAnnotation annotationWithName:@\"" + name + "\" andType:LUDProtocolTypeStruct],"
		}
	}

	return []byte("		" + v)
}

func (c ObjC) writeField(f *models.Field, prefix string) []byte {
	isArray := f.Size != ""
	var t string
	switch f.Type.Source {
	case "native":
		t = "LUDType" + convertToPascalCase(string(f.Type.NativeType)) + " *"
	case "user":
		// User-based types are always relative to the current package. Given
		// our naming rules, we can just append the name of the struct to the
		// current package name without needing to resort to any name-resolution
		// techniniques.
		t = c.prefix + prefix + convertToPascalCase(f.Type.CustomType) + " *"
	}
	if isArray {
		t = "NSArray<" + t + "> *"
	}

	t = "@property (nullable, nonatomic, retain) " + t + convertToCamelCase(f.Name) + ";\n"
	return []byte(t)
}

func (c ObjC) formatCode(code string) string {
	var buf bytes.Buffer
	err := quick.Highlight(&buf, code, "objectivec", "terminal", "pygments")
	if err != nil {
		log.Fatalf("BUG: Error processing ObjC source: %s", err)
	}

	return string(buf.Bytes())
}

func (c ObjC) integrationInstructions(pList *models.PackageList) string {
	list := []string{}
	for _, p := range *pList {
		list = append(list, "["+convertToPascalCase(p.Name)+" class]")
	}
	sort.Strings(list)

	separator := ", "
	if len(list) > 2 {
		separator = ",\n" + strings.Repeat(" ", 39)
	}

	data := templateData{
		"carthage":        aurora.Bold("Carthage"),
		"cartfile":        aurora.Magenta("Cartfile"),
		"cartfileRequire": fmt.Sprintf("%s %s", aurora.Magenta("github"), aurora.Red(`"ludwieg/objc"`)),
		"objcImport":      c.formatCode("#import <Ludwieg/Ludwieg.h>"),
		"objcRegister":    c.formatCode(fmt.Sprintf("[LUDDeserializer registerPackages:%s, nil];", strings.Join(list, separator))),
	}

	return string(processTemplate("objcIntegration", objcIntegrationSteps, data))
}
