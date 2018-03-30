package langs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"

	"github.com/ludwieg/ludco/models"
)

type Java struct {
	pkgName string
	out     string
}

func (c Java) Compile(in, out, pkgName, prefix string, packages *models.PackageList) {
	log.Infof("Initialising %s compiler", aurora.Red("Java"))
	if prefix != "" {
		log.Warn("Ignoring unnecessary --prefix option.")
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

	for _, p := range *packages {
		c.writePackage(&p)
	}

	log.Info("Succeeded")
	fmt.Println()
	fmt.Println(c.integrationInstructions(packages))
}

func (c Java) output(name string, contents []byte) {
	log.Infof("Writing %s/%s", aurora.Magenta(filepath.Base(c.out)), aurora.Magenta(name+".java"))
	err := ioutil.WriteFile(filepath.Join(c.out, name+".java"), contents, 0644)
	if err != nil {
		log.Errorf("Operation failed: %s", err)
		os.Exit(1)
	}
}

func (c Java) writePackage(p *models.Package) {
	if len(p.Fields) == 0 {
		c.writeEmptyPackage(p)
	} else {
		c.writeNormalPackage(p)
	}
}

func (c Java) writeEmptyPackage(p *models.Package) {
	name := convertToPascalCase(p.Name)
	c.output(name, processTemplate("emptyPackage", javaEmptyPackage, templateData{
		"pkg":        c.pkgName,
		"name":       name,
		"annotation": c.getClassAnnotationFor(p),
	}))
}

func (c Java) writeNormalPackage(p *models.Package) {
	pkgName := convertToPascalCase(p.Name)

	c.output(pkgName, processTemplate("package", javaPackage, templateData{
		"pkg":        c.pkgName,
		"annotation": c.getClassAnnotationFor(p),
		"name":       pkgName,
		"fields":     c.generateFields(p.Fields, pkgName),
		"getters":    c.generateGetters(p.Fields, pkgName),
		"setters":    c.generateSetters(p.Fields, pkgName),
	}))

	c.generateStructs(p.Structs, pkgName)
}

func (c Java) generateStructs(sArr []models.Struct, pkgName string) {
	for _, s := range sArr {
		name := pkgName + convertToPascalCase(s.Name)
		c.output(name, processTemplate("package", javaPackage, templateData{
			"pkg":        c.pkgName,
			"annotation": c.getClassAnnotationFor(&s),
			"name":       pkgName,
			"fields":     c.generateFields(s.Fields, pkgName),
			"getters":    c.generateGetters(s.Fields, pkgName),
			"setters":    c.generateSetters(s.Fields, pkgName),
		}))
		c.generateStructs(s.Structs, name)
	}
}

func (c Java) getClassAnnotationFor(item interface{}) string {
	var template string
	data := templateData{}
	switch item.(type) {
	case *models.Package:
		m := item.(*models.Package)
		data["id"] = m.Identifier
		template = javaAnnotationPackage
	case *models.Struct:
		template = javaAnnotationStruct
	default:
		log.Fatalf("BUG: getClassAnnotationFor failed for invalid type %#v", item)
	}
	return string(processTemplate("classAnnotation", template, data))
}

func (c Java) generateFields(fArr []models.Field, pkgName string) string {
	result := []string{}
	for i, f := range fArr {
		result = append(result, c.generateFieldAnnotation(i, &f, pkgName))
		result = append(result, c.generateField(&f, pkgName))
	}
	return strings.Join(result, "\n")
}

func (c Java) generateFieldAnnotation(index int, f *models.Field, pkgName string) string {
	data := templateData{
		"index": index,
	}
	template := ""

	if f.Type.Source == "native" {
		data["type"] = strings.ToUpper(string(f.Type.NativeType))
		if f.IsArray() {
			template = javaFieldAnnotationNativeArray
		} else {
			template = javaFieldAnnotationNative
		}
	} else {
		data["type"] = pkgName + convertToPascalCase(f.Type.CustomType)
		if f.IsArray() {
			template = javaFieldAnnotationCustomArray
		} else {
			template = javaFieldAnnotationCustom
		}
	}

	return strings.Repeat(" ", 4) + string(processTemplate("fieldAnnotation", template, data))
}

func (c Java) generateField(f *models.Field, pkgName string) string {
	var kind string
	if f.Type.Source == "native" {
		kind = "Type" + convertToPascalCase(string(f.Type.NativeType))
		if f.IsArray() {
			kind = "TypeArray<" + kind + ">"
		}
	} else {
		kind = "TypeStruct<" + pkgName + convertToPascalCase(string(f.Type.CustomType)) + ">"
		if f.IsArray() {
			kind = "TypeArray<" + kind + ">"
		}
	}

	return strings.Repeat(" ", 4) + string(processTemplate("fieldImplementation", javaField, templateData{
		"type":        kind,
		"name":        convertToCamelCase(f.Name),
		"initializer": c.generateInitializer(f, pkgName),
	}))
}

func (c Java) generateInitializer(f *models.Field, pkgName string) string {
	item := ""
	if f.Type.Source == "native" {
		baseType := "Type" + convertToPascalCase(string(f.Type.NativeType))
		if f.IsArray() {
			item = item + "TypeArray<>(" + baseType + ".class)"
		} else {
			item = item + baseType + "()"
		}
	} else {
		if f.IsArray() {
			item = item + "TypeArray<>(TypeStruct.class)"
		} else {
			item = item + "TypeStruct<>(" + pkgName + convertToPascalCase(string(f.Type.CustomType)) + ".class)"
		}
	}
	return item
}

func (c Java) generateGetters(fArr []models.Field, pkgName string) string {
	items := []string{}
	for _, f := range fArr {
		var template string
		var data = templateData{
			"name":      convertToPascalCase(f.Name),
			"fieldName": convertToCamelCase(f.Name),
			"type":      c.nativeTypeForField(&f, pkgName),
		}
		if f.Type.Source == "native" {
			if f.IsArray() {
				data["baseType"] = c.nativeTypeForProtocolType(f.Type.NativeType)
				template = javaGetterNativeArray
			} else {
				if f.Type.NativeType == models.TypeAny {
					template = javaGetterNativeAny
				} else {
					template = javaGetterNative
				}
			}
		} else {
			if f.IsArray() {
				data["baseType"] = pkgName + convertToPascalCase(f.Type.CustomType)
				template = javaGetterCustomArray
			} else {
				template = javaGetterCustom
			}
		}
		items = append(items, strings.Repeat(" ", 4)+string(processTemplate("javaGetter", template, data)))
	}
	return strings.Join(items, "\n")
}

func (c Java) generateSetters(fArr []models.Field, pkgName string) string {
	items := []string{}
	for _, f := range fArr {
		var template string
		var data = templateData{
			"name":      convertToPascalCase(f.Name),
			"fieldName": convertToCamelCase(f.Name),
			"type":      c.nativeTypeForField(&f, pkgName),
			"pkg":       pkgName,
		}
		if f.Type.Source == "native" {
			data["baseType"] = c.nativeTypeForProtocolType(f.Type.NativeType)
			if f.IsArray() {
				template = javaGetterNativeArray
			} else {
				if f.Type.NativeType == models.TypeAny {
					template = javaSetterNativeAny
				} else if f.Type.NativeType == models.TypeDynInt {
					template = javaSetterNativeDynInt
				} else {
					template = javaSetterNative
				}
			}
		} else {
			data["baseType"] = pkgName + convertToPascalCase(f.Type.CustomType)
			if f.IsArray() {
				template = javaSetterCustomArray
			} else {
				template = javaSetterCustom
			}
		}
		items = append(items, strings.Repeat(" ", 4)+string(processTemplate("javaGetter", template, data)))
	}
	return strings.Join(items, "\n")
}

func (c Java) nativeTypeForField(f *models.Field, pkgName string) string {
	var kind string
	if f.Type.Source == "native" {
		kind = c.nativeTypeForProtocolType(f.Type.NativeType)
	} else {
		kind = pkgName + convertToPascalCase(f.Type.CustomType)
	}
	if f.IsArray() {
		kind = "List<" + kind + ">"
	}
	return kind
}

func (c Java) nativeTypeForProtocolType(t models.NativeType) string {
	kind := ""
	switch t {
	case models.TypeUint8, models.TypeUint32, models.TypeByte:
		kind = "Integer"
	case models.TypeUint64:
		kind = "Long"
	case models.TypeDouble:
		kind = "Double"
	case models.TypeString:
		kind = "String"
	case models.TypeBlob:
		kind = "[]byte"
	case models.TypeBool:
		kind = "Boolean"
	case models.TypeUUID:
		kind = "UUID"
	case models.TypeAny:
		kind = "Object"
	case models.TypeDynInt:
		kind = "DynInt"
	}

	if kind == "" {
		log.Fatalf("BUG: Cannot coerce unknown type %#v to native type.", t)
	}

	return kind
}

func (c Java) integrationInstructions(pList *models.PackageList) string {
	list := []string{}
	for _, p := range *pList {
		list = append(list, convertToPascalCase(p.Name)+".class")
	}
	sort.Strings(list)

	separator := ", "
	if len(list) > 2 {
		separator = ",\n" + strings.Repeat(" ", 12)
	}

	jitPackRepository := `      allprojects {
          repositories {
              ...
              maven { url 'https://jitpack.io' }
          }
      }`

	jitPackDependency := `      dependencies {
          dependencies {
              compile 'com.github.ludwieg:kotlin:v0.1.4'
          }
      }`

	data := templateData{
		"dependencies":   aurora.Bold("Dependencies"),
		"integration":    aurora.Bold("Integration"),
		"initialization": aurora.Bold("Initialization"),
		"jitPack":        aurora.Bold("JitPack"),

		"gradle":    aurora.Magenta("Gradle"),
		"maven":     aurora.Magenta("Maven"),
		"sbt":       aurora.Magenta("Sbt"),
		"leiningen": aurora.Magenta("Leiningen"),
		"libPath":   aurora.Magenta("ludwieg/kotlin"),

		"jitPackRepository": c.formatCode(jitPackRepository, "groovy"),
		"jitPackDependency": c.formatCode(jitPackDependency, "groovy"),

		"javaImportLudwieg": c.formatCode("import io.vito.ludwieg;", "java"),
		"javaImportPkg":     c.formatCode("import "+c.pkgName+";", "java"),
		"javaRegister":      c.formatCode("Registry.Companion.getInstance().register("+strings.Join(list, separator)+");", "java"),
	}

	return string(processTemplate("javaIntegration", javaIntegrationSteps, data))
}

func (c Java) formatCode(code, lang string) string {
	var buf bytes.Buffer
	err := quick.Highlight(&buf, code, lang, "terminal", "pygments")
	if err != nil {
		log.Fatalf("BUG: Error processing java source: %s", err)
	}

	return string(buf.Bytes())
}
