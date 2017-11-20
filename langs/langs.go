package langs

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"

	"github.com/ludwieg/compiler/models"
)

type Compiler interface {
	Compile(in, out, pkgName, prefix string, packages *models.PackageList)
}

type templateData map[string]interface{}

func processTemplate(facility, templateString string, data templateData) []byte {
	var tpl bytes.Buffer
	t := template.Must(template.New(facility).Parse(templateString))
	err := t.Execute(&tpl, data)
	if err != nil {
		log.WithField("facility", "template-processor").Errorf("Failed processing template: %s", err)
		os.Exit(1)
	}
	return tpl.Bytes()
}

func convertToPascalCase(val string) string {
	arr := []string{}
	for _, s := range strings.Split(val, "_") {
		rs := []rune(s)
		arr = append(arr, strings.ToUpper(string(rs[0]))+string(rs[1:]))
	}
	return strings.Join(arr, "")
}

func convertToCamelCase(val string) string {
	arr := []string{}
	for i, s := range strings.Split(val, "_") {
		rs := []rune(s)
		s = ""
		if i == 0 {
			s = strings.ToLower(string(rs[0]))
		} else {
			s = strings.ToUpper(string(rs[0]))
		}
		arr = append(arr, s+string(rs[1:]))
	}
	return strings.Join(arr, "")
}
