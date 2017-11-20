package validation

import (
	"fmt"
	"math"
	"strconv"

	"github.com/ludwieg/compiler/parser"
)

func hasKey(key string, obj map[string]parser.Object) bool {
	for k := range obj {
		if k == key {
			return true
		}
	}
	return false
}

func validateStruct(obj parser.Object) []error {
	if obj.ObjectType != parser.ObjStruct {
		return []error{}
	}
	errors := []error{}
	fields := map[string]parser.Object{}
	structs := map[string]parser.Object{}

	for _, data := range obj.Contents {
		switch data.ObjectType {
		case parser.ObjID:
			errors = append(errors, fmt.Errorf("struct `%s' has prohibited identifier declaration", obj.Name))
		case parser.ObjStruct:
			if hasKey(data.Name, structs) {
				errors = append(errors, fmt.Errorf("duplicated struct definition `%s'", data.Name))
			} else {
				structs[data.Name] = data
				errors = append(errors, validateStruct(data)...)
			}
		case parser.ObjField:
			if hasKey(data.Name, fields) {
				errors = append(errors, fmt.Errorf("struct `%s' has duplicated field definition for `%s'", obj.Name, data.Name))
			} else {
				fields[data.Name] = data
			}
			if data.Source == parser.SourceUser {
				errors = append(errors, fmt.Errorf("struct `%s' has field with prohibited custom type `%s'", obj.Name, data.Name))
			}
		case parser.ObjArray:
			if hasKey(data.Name, fields) {
				errors = append(errors, fmt.Errorf("struct `%s' has duplicated field definition `%s'", obj.Name, data.Name))
			} else {
				fields[data.Name] = data
			}
			errors = append(errors, validateArrayField(data)...)
		}
	}
	return errors
}

func validateArrayField(obj parser.Object) []error {
	if obj.ObjectType != parser.ObjArray {
		return []error{}
	}
	errors := []error{}
	if obj.Size != "*" {
		size, err := strconv.Atoi(obj.Size)
		if err != nil {
			errors = append(errors, fmt.Errorf("invalid size for array field `%s'", obj.Name))
		} else {
			if size < 1 {
				errors = append(errors, fmt.Errorf("invalid size for array field `%s' (minimum allowed is 1)", obj.Name))
			} else if size >= math.MaxUint32-1 {
				errors = append(errors, fmt.Errorf("invalid size for array field `%s' (maximum allowed is %d)", obj.Name, math.MaxUint32-1))
			}
		}
	}
	return errors
}

func Validate(p parser.Package) []error {
	hasIdentifier := false
	errors := []error{}
	fields := map[string]parser.Object{}
	structs := map[string]parser.Object{}

	for _, data := range p.Contents {
		switch data.ObjectType {
		case parser.ObjID:
			if hasIdentifier {
				errors = append(errors, fmt.Errorf("duplicated identifier declaration"))
			}
			hasIdentifier = true
		case parser.ObjField:
			if hasKey(data.Name, fields) {
				errors = append(errors, fmt.Errorf("duplicated field definition `%s'", data.Name))
			} else {
				fields[data.Name] = data
			}
		case parser.ObjArray:
			if hasKey(data.Name, fields) {
				errors = append(errors, fmt.Errorf("duplicated field definition `%s'", data.Name))
			} else {
				fields[data.Name] = data
			}
			errors = append(errors, validateArrayField(data)...)
		case parser.ObjStruct:
			if hasKey(data.Name, structs) {
				errors = append(errors, fmt.Errorf("duplicated struct definition `%s'", data.Name))
			} else {
				structs[data.Name] = data
				errors = append(errors, validateStruct(data)...)
			}
		}
	}

	for n, i := range fields {
		if i.Source != parser.SourceUser {
			continue
		}

		if !hasKey(i.Kind, structs) {
			errors = append(errors, fmt.Errorf("field `%s' references unknown type `%s'", n, i.Kind))
		}
	}

	return errors
}
