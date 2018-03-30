package models

import (
	"fmt"
	"strconv"

	"github.com/ludwieg/ludco/parser"
)

func fail(msg string, args ...interface{}) {
	panic(fmt.Errorf(msg, args...))
}

// Attribute represents a possible attribute used in a Field object
type Attribute string

const (
	// AttributeDeprecated indicates that the field is not used anymore. The
	// field should still be considered when handling internal data structures,
	// but it should not be available for the end-user (either by using the
	// target language deprecation facilities, or completely omitting it from
	// the generated source code)
	AttributeDeprecated Attribute = "deprecated"
)

// NativeType represents a type commonly known by Ludwieg, such as numbers and
// strings.
type NativeType string

const (
	// TypeDynInt automatically detects and processes an integer value based on
	// its size
	TypeDynInt NativeType = "dynint"

	// TypeUint8 represents an 8-bit unsigned integer
	TypeUint8 NativeType = "uint8"

	// TypeUint32 represents a 32-bit unsigned integer
	TypeUint32 NativeType = "uint32"

	// TypeUint64 represents a 64-bit unsigned integer
	TypeUint64 NativeType = "uint64"

	// TypeByte represents an 8-bit unsigned integer. Internally, it works as
	// an alias to TypeUint8.
	TypeByte NativeType = "byte"

	// TypeDouble represents an IEEE-754 64-bit floating-point number
	TypeDouble NativeType = "double"

	// TypeString represents an UTF-8 encoded string
	TypeString NativeType = "string"

	// TypeBlob represents a Binary Large OBject.
	TypeBlob NativeType = "blob"

	// TypeBool represents a true/false value. Internaly, it works as an alias
	// to TypeUint8
	TypeBool NativeType = "bool"

	// TypeUUID represents an Universally Unique IDentifier. It is stored as
	// a 128-bit (16 bytes) value, rather than a 256-bit (32 bytes) string
	// value
	TypeUUID NativeType = "uuid"

	// TypeAny represents a field that can contain any possible value. It is
	// especially useful for dynamic data structures. On the target language
	// it must has a "generic" type, such as interface{} (Go), Object (Java),
	// or id (ObjC). On languages that thread managed and primitive types,
	// the parser must prefer to use managed types (NSNumber for ObjC, Boolean
	// for Java, for instance).
	TypeAny NativeType = "any"
)

// Source identifies the source of a type used in a Field
type Source string

const (
	// SourceUser indicates that the type is declared by the User, and the
	// package holding it contains a Structure with same name
	SourceUser Source = "user"

	// SourceNative indicates that the type is builtin on Ludwieg
	SourceNative = "native"
)

// Type holds information about typing on a Field
type Type struct {
	// Source indicates the source of the field type
	Source Source

	// NativeType holds the NativeType of the field, if Source is SourceNative
	NativeType NativeType

	// CustomType holds the name of the user structure used as the field type.
	// Available when source is SourceUser
	CustomType string
}

// ObjectType identifies whether the field represents a single-value field, or
// an Array
type ObjectType string

const (
	// ObjectTypeField indicates that the object represents a single-value Field
	ObjectTypeField ObjectType = "field"

	// ObjectTypeArray indicates that the object represents an Array with a
	// static or dynamic size.
	ObjectTypeArray = "array"
)

// ------

// Package represents a package defined in a Ludwieg file
type Package struct {
	// Name is used to identify a package. It also yields the Class/Struct name
	// on the target language, when generating sources
	Name string

	// Identifier holds a hexadecimal value identifying the package among other
	// packages. This identifier must be unique among all other packages in the
	// protocol
	Identifier string

	// Structs defines custom user types used in the current package
	Structs []Struct

	// Fields defines all fields that the package carries
	Fields []Field
}

// RawIdentifier returns the raw identifier of a package as a byte
func (p Package) RawIdentifier() byte {
	v, err := strconv.ParseUint(p.Identifier[2:], 16, 8)
	if err != nil {
		fail("error parsing identifier for package %s: %s", p.Name, err)
		return 0
	}
	return byte(v)
}

// PackageList represents a list of packages with sorting capabilities
type PackageList []Package

func (s PackageList) Len() int {
	return len(s)
}
func (s PackageList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s PackageList) Less(i, j int) bool {
	return s[i].RawIdentifier() < s[j].RawIdentifier()
}

// Struct represents a custom user type
type Struct struct {

	// Name identifies the struct among other possible structures on the current
	// package. It is also used to generate classes/structs on the target
	// language then generating sources
	Name string

	// Structs defines custom user types used in the current structure
	Structs []Struct

	// Fields defines all fields that the structure carries
	Fields []Field
}

// Field contains metadata about a field defined in a package.
type Field struct {

	// ObjectType determines whether the field represents an Array or a plain
	// field
	ObjectType ObjectType

	// Type holds metadata about the field type and its source
	Type Type

	// Name indicates the name of the field. Used by the source generator.
	Name string

	// Size indicates the size of an array size. Possible values satisfies the
	// regular expression ^(\*|\d+)$, when * represents an array with arbitrary
	// quantity of values
	Size string

	// Attributes holds any attribute added to the field
	Attributes []Attribute
}

// HasAttribute determines whether a field contains a given attribute
func (f Field) HasAttribute(attr Attribute) bool {
	for _, a := range f.Attributes {
		if a == attr {
			return true
		}
	}
	return false
}

// IsArray determines whether a field represents an Array
func (f Field) IsArray() bool {
	return f.Size != ""
}

func sourceFromParser(source string) Source {
	if source == parser.SourceNative {
		return SourceNative
	} else if source == parser.SourceUser {
		return SourceUser
	} else {
		fail("invalid source %s", source)
		return ""
	}
}

func attributesFromParser(attributes []string) []Attribute {
	arr := []Attribute{}
	for _, att := range attributes {
		if att == parser.AttributeDeprecated {
			arr = append(arr, AttributeDeprecated)
		} else {
			fail("invalid attribute %s", att)
		}
	}
	return arr
}

func typeFromParser(obj parser.Object) Type {
	t := Type{
		Source: sourceFromParser(obj.Source),
	}

	if t.Source == SourceNative {
		t.NativeType = nativeTypeFromParser(obj.Kind)
	} else {
		t.CustomType = obj.Kind
	}

	return t
}

func nativeTypeFromParser(t string) NativeType {
	switch t {
	case "dynint":
		return TypeDynInt
	case "uint8":
		return TypeUint8
	case "uint32":
		return TypeUint32
	case "uint64":
		return TypeUint64
	case "byte":
		return TypeByte
	case "double":
		return TypeDouble
	case "string":
		return TypeString
	case "blob":
		return TypeBlob
	case "bool":
		return TypeBool
	case "uuid":
		return TypeUUID
	case "any":
		return TypeAny
	}
	fail("invalid native type %s", t)
	return ""
}

func structFromParser(obj parser.Object) Struct {
	str := Struct{
		Name:    obj.Name,
		Fields:  []Field{},
		Structs: []Struct{},
	}

	for _, i := range obj.Contents {
		switch i.ObjectType {
		case parser.ObjField, parser.ObjArray:
			str.Fields = append(str.Fields, fieldFromParser(i))
		case parser.ObjStruct:
			str.Structs = append(str.Structs, structFromParser(i))
		}
	}
	return str
}

func fieldFromParser(obj parser.Object) Field {

	result := Field{
		Type:       typeFromParser(obj),
		ObjectType: ObjectTypeField,
		Name:       obj.Name,
		Attributes: attributesFromParser(obj.Attributes),
	}

	if obj.ObjectType == parser.ObjArray {
		result.Size = obj.Size
	}

	return result
}

// ConvertASTPackage attempts to convert a `parser.Package` type into a
// `models.Package` object, which has extra granular options. Please do notice
// that instead of returning errors, this function panics if any inconsistency
// is found. That said, the callee is reponsible for either recovering from the
// panic, or allowing it to crash the whole application.
func ConvertASTPackage(ast parser.Package) *Package {
	pkg := Package{
		Name:    ast.Name,
		Structs: []Struct{},
		Fields:  []Field{},
	}
	for _, i := range ast.Contents {
		switch i.ObjectType {
		case parser.ObjField, parser.ObjArray:
			pkg.Fields = append(pkg.Fields, fieldFromParser(i))
		case parser.ObjStruct:
			pkg.Structs = append(pkg.Structs, structFromParser(i))
		case parser.ObjID:
			pkg.Identifier = i.Value
		}
	}
	return &pkg
}
