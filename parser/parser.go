package parser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Type struct {
	Name   string
	Source string
}

type Package struct {
	Name     string
	Contents []Object
}

type Object struct {
	ObjectType string
	Source     string
	Kind       string
	Name       string
	Size       string
	Value      string
	Contents   []Object
	Attributes []string
}

func asString(data interface{}) string {
	if i, ok := data.([]uint8); ok {
		return string([]byte(i))
	} else if i, ok := data.(string); ok {
		return i
	} else {
		i := data.([]interface{})
		arr := make([]uint8, len(i))
		for x, v := range i {
			arr[x] = (v.([]uint8))[0]
		}
		return asString(arr)
	}
}

func strSlice(rawSlice interface{}) []string {
	switch t := rawSlice.(type) {
	case []interface{}:
		arr := make([]string, len(t))
		for i, s := range t {
			arr[i] = s.(string)
		}
		return arr
	case []string:
		return t
	}
	return []string{}
}

func objSlice(rawSlice interface{}) []Object {
	switch t := rawSlice.(type) {
	case []interface{}:
		arr := make([]Object, len(t))
		for i, s := range t {
			arr[i] = s.(Object)
		}
		return arr
	case []Object:
		return t
	}
	return []Object{}
}

const (
	SourceNative        = "native"
	SourceUser          = "user"
	ObjID               = "id"
	ObjField            = "field"
	ObjArray            = "array"
	ObjStruct           = "struct"
	AttributeDeprecated = "deprecated"
)

var g = &grammar{
	rules: []*rule{
		{
			name: "start",
			pos:  position{line: 87, col: 1, offset: 1830},
			expr: &actionExpr{
				pos: position{line: 88, col: 7, offset: 1842},
				run: (*parser).callonstart1,
				expr: &labeledExpr{
					pos:   position{line: 88, col: 7, offset: 1842},
					label: "val",
					expr: &ruleRefExpr{
						pos:  position{line: 88, col: 11, offset: 1846},
						name: "contents",
					},
				},
			},
		},
		{
			name: "whitespace",
			pos:  position{line: 90, col: 1, offset: 1876},
			expr: &charClassMatcher{
				pos:        position{line: 91, col: 7, offset: 1893},
				val:        "[ \\t]",
				chars:      []rune{' ', '\t'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "EOL",
			pos:  position{line: 93, col: 1, offset: 1900},
			expr: &seqExpr{
				pos: position{line: 94, col: 7, offset: 1910},
				exprs: []interface{}{
					&oneOrMoreExpr{
						pos: position{line: 94, col: 7, offset: 1910},
						expr: &charClassMatcher{
							pos:        position{line: 94, col: 7, offset: 1910},
							val:        "[ \\t\\r\\n]",
							chars:      []rune{' ', '\t', '\r', '\n'},
							ignoreCase: false,
							inverted:   false,
						},
					},
					&zeroOrOneExpr{
						pos: position{line: 94, col: 18, offset: 1921},
						expr: &ruleRefExpr{
							pos:  position{line: 94, col: 18, offset: 1921},
							name: "comment",
						},
					},
				},
			},
		},
		{
			name: "EOF",
			pos:  position{line: 96, col: 1, offset: 1931},
			expr: &notExpr{
				pos: position{line: 97, col: 7, offset: 1941},
				expr: &anyMatcher{
					line: 97, col: 8, offset: 1942,
				},
			},
		},
		{
			name:        "_",
			displayName: "\"whitespace\"",
			pos:         position{line: 99, col: 1, offset: 1945},
			expr: &actionExpr{
				pos: position{line: 100, col: 7, offset: 1966},
				run: (*parser).callon_1,
				expr: &zeroOrMoreExpr{
					pos: position{line: 100, col: 7, offset: 1966},
					expr: &ruleRefExpr{
						pos:  position{line: 100, col: 7, offset: 1966},
						name: "whitespace",
					},
				},
			},
		},
		{
			name:        "__",
			displayName: "\"eol\"",
			pos:         position{line: 102, col: 1, offset: 1999},
			expr: &actionExpr{
				pos: position{line: 103, col: 7, offset: 2014},
				run: (*parser).callon__1,
				expr: &zeroOrMoreExpr{
					pos: position{line: 103, col: 7, offset: 2014},
					expr: &ruleRefExpr{
						pos:  position{line: 103, col: 7, offset: 2014},
						name: "EOL",
					},
				},
			},
		},
		{
			name: "digit",
			pos:  position{line: 105, col: 1, offset: 2040},
			expr: &charClassMatcher{
				pos:        position{line: 106, col: 7, offset: 2052},
				val:        "[0-9]",
				ranges:     []rune{'0', '9'},
				ignoreCase: false,
				inverted:   false,
			},
		},
		{
			name: "digits",
			pos:  position{line: 108, col: 1, offset: 2059},
			expr: &actionExpr{
				pos: position{line: 109, col: 7, offset: 2072},
				run: (*parser).callondigits1,
				expr: &labeledExpr{
					pos:   position{line: 109, col: 7, offset: 2072},
					label: "digits",
					expr: &zeroOrMoreExpr{
						pos: position{line: 109, col: 14, offset: 2079},
						expr: &ruleRefExpr{
							pos:  position{line: 109, col: 14, offset: 2079},
							name: "digit",
						},
					},
				},
			},
		},
		{
			name: "hexValue",
			pos:  position{line: 111, col: 1, offset: 2120},
			expr: &actionExpr{
				pos: position{line: 112, col: 7, offset: 2135},
				run: (*parser).callonhexValue1,
				expr: &seqExpr{
					pos: position{line: 112, col: 7, offset: 2135},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 112, col: 7, offset: 2135},
							label: "first",
							expr: &litMatcher{
								pos:        position{line: 112, col: 13, offset: 2141},
								val:        "0x",
								ignoreCase: false,
							},
						},
						&labeledExpr{
							pos:   position{line: 112, col: 18, offset: 2146},
							label: "rest",
							expr: &oneOrMoreExpr{
								pos: position{line: 112, col: 23, offset: 2151},
								expr: &charClassMatcher{
									pos:        position{line: 112, col: 23, offset: 2151},
									val:        "[a-fA-F0-9]",
									ranges:     []rune{'a', 'f', 'A', 'F', '0', '9'},
									ignoreCase: false,
									inverted:   false,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "itemName",
			pos:  position{line: 114, col: 1, offset: 2214},
			expr: &actionExpr{
				pos: position{line: 115, col: 7, offset: 2229},
				run: (*parser).callonitemName1,
				expr: &labeledExpr{
					pos:   position{line: 115, col: 7, offset: 2229},
					label: "value",
					expr: &oneOrMoreExpr{
						pos: position{line: 115, col: 13, offset: 2235},
						expr: &charClassMatcher{
							pos:        position{line: 115, col: 13, offset: 2235},
							val:        "[a-z_]",
							chars:      []rune{'_'},
							ranges:     []rune{'a', 'z'},
							ignoreCase: false,
							inverted:   false,
						},
					},
				},
			},
		},
		{
			name: "openCurlyBrace",
			pos:  position{line: 117, col: 1, offset: 2276},
			expr: &litMatcher{
				pos:        position{line: 118, col: 7, offset: 2297},
				val:        "{",
				ignoreCase: false,
			},
		},
		{
			name: "closeCurlyBrace",
			pos:  position{line: 120, col: 1, offset: 2302},
			expr: &litMatcher{
				pos:        position{line: 121, col: 7, offset: 2324},
				val:        "}",
				ignoreCase: false,
			},
		},
		{
			name: "openSquareBrace",
			pos:  position{line: 123, col: 1, offset: 2329},
			expr: &litMatcher{
				pos:        position{line: 124, col: 7, offset: 2351},
				val:        "[",
				ignoreCase: false,
			},
		},
		{
			name: "closeSquareBrace",
			pos:  position{line: 126, col: 1, offset: 2356},
			expr: &litMatcher{
				pos:        position{line: 127, col: 7, offset: 2379},
				val:        "]",
				ignoreCase: false,
			},
		},
		{
			name: "comment",
			pos:  position{line: 129, col: 1, offset: 2384},
			expr: &actionExpr{
				pos: position{line: 130, col: 7, offset: 2398},
				run: (*parser).calloncomment1,
				expr: &seqExpr{
					pos: position{line: 130, col: 7, offset: 2398},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 130, col: 7, offset: 2398},
							val:        "//",
							ignoreCase: false,
						},
						&zeroOrMoreExpr{
							pos: position{line: 130, col: 12, offset: 2403},
							expr: &charClassMatcher{
								pos:        position{line: 130, col: 12, offset: 2403},
								val:        "[^\\n]",
								chars:      []rune{'\n'},
								ignoreCase: false,
								inverted:   true,
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 130, col: 19, offset: 2410},
							expr: &choiceExpr{
								pos: position{line: 130, col: 20, offset: 2411},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 130, col: 20, offset: 2411},
										name: "EOL",
									},
									&ruleRefExpr{
										pos:  position{line: 130, col: 24, offset: 2415},
										name: "EOF",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "attribute",
			pos:  position{line: 132, col: 1, offset: 2442},
			expr: &actionExpr{
				pos: position{line: 133, col: 7, offset: 2458},
				run: (*parser).callonattribute1,
				expr: &seqExpr{
					pos: position{line: 133, col: 7, offset: 2458},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 133, col: 7, offset: 2458},
							name: "_",
						},
						&litMatcher{
							pos:        position{line: 133, col: 9, offset: 2460},
							val:        "!",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 133, col: 13, offset: 2464},
							label: "flag",
							expr: &litMatcher{
								pos:        position{line: 133, col: 19, offset: 2470},
								val:        "deprecated",
								ignoreCase: false,
							},
						},
					},
				},
			},
		},
		{
			name: "attributeList",
			pos:  position{line: 135, col: 1, offset: 2516},
			expr: &actionExpr{
				pos: position{line: 136, col: 4, offset: 2533},
				run: (*parser).callonattributeList1,
				expr: &labeledExpr{
					pos:   position{line: 136, col: 4, offset: 2533},
					label: "attr",
					expr: &oneOrMoreExpr{
						pos: position{line: 136, col: 9, offset: 2538},
						expr: &ruleRefExpr{
							pos:  position{line: 136, col: 9, offset: 2538},
							name: "attribute",
						},
					},
				},
			},
		},
		{
			name: "arraySize",
			pos:  position{line: 140, col: 1, offset: 2612},
			expr: &actionExpr{
				pos: position{line: 141, col: 7, offset: 2628},
				run: (*parser).callonarraySize1,
				expr: &seqExpr{
					pos: position{line: 141, col: 7, offset: 2628},
					exprs: []interface{}{
						&ruleRefExpr{
							pos:  position{line: 141, col: 7, offset: 2628},
							name: "openSquareBrace",
						},
						&labeledExpr{
							pos:   position{line: 141, col: 23, offset: 2644},
							label: "val",
							expr: &choiceExpr{
								pos: position{line: 141, col: 28, offset: 2649},
								alternatives: []interface{}{
									&litMatcher{
										pos:        position{line: 141, col: 28, offset: 2649},
										val:        "*",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 141, col: 34, offset: 2655},
										name: "digits",
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 141, col: 42, offset: 2663},
							name: "closeSquareBrace",
						},
					},
				},
			},
		},
		{
			name: "nativeType",
			pos:  position{line: 143, col: 1, offset: 2711},
			expr: &actionExpr{
				pos: position{line: 144, col: 7, offset: 2728},
				run: (*parser).callonnativeType1,
				expr: &labeledExpr{
					pos:   position{line: 144, col: 7, offset: 2728},
					label: "val",
					expr: &choiceExpr{
						pos: position{line: 144, col: 12, offset: 2733},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 144, col: 12, offset: 2733},
								val:        "uint8",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 144, col: 22, offset: 2743},
								val:        "uint32",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 144, col: 33, offset: 2754},
								val:        "uint64",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 144, col: 44, offset: 2765},
								val:        "byte",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 144, col: 53, offset: 2774},
								val:        "double",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 144, col: 64, offset: 2785},
								val:        "string",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 144, col: 75, offset: 2796},
								val:        "blob",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 144, col: 84, offset: 2805},
								val:        "bool",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 144, col: 93, offset: 2814},
								val:        "uuid",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 144, col: 102, offset: 2823},
								val:        "any",
								ignoreCase: false,
							},
						},
					},
				},
			},
		},
		{
			name: "userType",
			pos:  position{line: 151, col: 1, offset: 2942},
			expr: &actionExpr{
				pos: position{line: 152, col: 7, offset: 2957},
				run: (*parser).callonuserType1,
				expr: &seqExpr{
					pos: position{line: 152, col: 7, offset: 2957},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 152, col: 7, offset: 2957},
							val:        "@",
							ignoreCase: false,
						},
						&labeledExpr{
							pos:   position{line: 152, col: 11, offset: 2961},
							label: "val",
							expr: &ruleRefExpr{
								pos:  position{line: 152, col: 15, offset: 2965},
								name: "itemName",
							},
						},
					},
				},
			},
		},
		{
			name: "idDefinition",
			pos:  position{line: 159, col: 1, offset: 3083},
			expr: &actionExpr{
				pos: position{line: 160, col: 7, offset: 3102},
				run: (*parser).callonidDefinition1,
				expr: &seqExpr{
					pos: position{line: 160, col: 7, offset: 3102},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 160, col: 7, offset: 3102},
							val:        "id",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 160, col: 12, offset: 3107},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 160, col: 14, offset: 3109},
							label: "val",
							expr: &ruleRefExpr{
								pos:  position{line: 160, col: 18, offset: 3113},
								name: "hexValue",
							},
						},
					},
				},
			},
		},
		{
			name: "fieldDefinition",
			pos:  position{line: 167, col: 1, offset: 3234},
			expr: &actionExpr{
				pos: position{line: 168, col: 7, offset: 3256},
				run: (*parser).callonfieldDefinition1,
				expr: &seqExpr{
					pos: position{line: 168, col: 7, offset: 3256},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 168, col: 7, offset: 3256},
							label: "t",
							expr: &choiceExpr{
								pos: position{line: 168, col: 10, offset: 3259},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 168, col: 10, offset: 3259},
										name: "nativeType",
									},
									&ruleRefExpr{
										pos:  position{line: 168, col: 23, offset: 3272},
										name: "userType",
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 168, col: 33, offset: 3282},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 168, col: 35, offset: 3284},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 168, col: 40, offset: 3289},
								name: "itemName",
							},
						},
						&labeledExpr{
							pos:   position{line: 168, col: 49, offset: 3298},
							label: "attributes",
							expr: &zeroOrOneExpr{
								pos: position{line: 168, col: 60, offset: 3309},
								expr: &ruleRefExpr{
									pos:  position{line: 168, col: 60, offset: 3309},
									name: "attributeList",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "arrayDefinition",
			pos:  position{line: 178, col: 1, offset: 3554},
			expr: &actionExpr{
				pos: position{line: 179, col: 7, offset: 3576},
				run: (*parser).callonarrayDefinition1,
				expr: &seqExpr{
					pos: position{line: 179, col: 7, offset: 3576},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 179, col: 7, offset: 3576},
							label: "t",
							expr: &choiceExpr{
								pos: position{line: 179, col: 10, offset: 3579},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 179, col: 10, offset: 3579},
										name: "nativeType",
									},
									&ruleRefExpr{
										pos:  position{line: 179, col: 23, offset: 3592},
										name: "userType",
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 179, col: 33, offset: 3602},
							label: "size",
							expr: &ruleRefExpr{
								pos:  position{line: 179, col: 38, offset: 3607},
								name: "arraySize",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 179, col: 48, offset: 3617},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 179, col: 50, offset: 3619},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 179, col: 55, offset: 3624},
								name: "itemName",
							},
						},
						&labeledExpr{
							pos:   position{line: 179, col: 64, offset: 3633},
							label: "attributes",
							expr: &zeroOrOneExpr{
								pos: position{line: 179, col: 75, offset: 3644},
								expr: &ruleRefExpr{
									pos:  position{line: 179, col: 75, offset: 3644},
									name: "attributeList",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "contents",
			pos:  position{line: 192, col: 1, offset: 3946},
			expr: &actionExpr{
				pos: position{line: 193, col: 7, offset: 3961},
				run: (*parser).calloncontents1,
				expr: &labeledExpr{
					pos:   position{line: 193, col: 7, offset: 3961},
					label: "val",
					expr: &oneOrMoreExpr{
						pos: position{line: 193, col: 11, offset: 3965},
						expr: &ruleRefExpr{
							pos:  position{line: 193, col: 11, offset: 3965},
							name: "fileContents",
						},
					},
				},
			},
		},
		{
			name: "fileContents",
			pos:  position{line: 195, col: 1, offset: 4000},
			expr: &actionExpr{
				pos: position{line: 196, col: 7, offset: 4019},
				run: (*parser).callonfileContents1,
				expr: &seqExpr{
					pos: position{line: 196, col: 7, offset: 4019},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 196, col: 7, offset: 4019},
							expr: &ruleRefExpr{
								pos:  position{line: 196, col: 7, offset: 4019},
								name: "__",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 196, col: 11, offset: 4023},
							expr: &ruleRefExpr{
								pos:  position{line: 196, col: 11, offset: 4023},
								name: "comment",
							},
						},
						&labeledExpr{
							pos:   position{line: 196, col: 20, offset: 4032},
							label: "val",
							expr: &ruleRefExpr{
								pos:  position{line: 196, col: 24, offset: 4036},
								name: "pkg",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 196, col: 28, offset: 4040},
							expr: &ruleRefExpr{
								pos:  position{line: 196, col: 28, offset: 4040},
								name: "__",
							},
						},
					},
				},
			},
		},
		{
			name: "str",
			pos:  position{line: 200, col: 1, offset: 4080},
			expr: &actionExpr{
				pos: position{line: 201, col: 7, offset: 4090},
				run: (*parser).callonstr1,
				expr: &seqExpr{
					pos: position{line: 201, col: 7, offset: 4090},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 201, col: 7, offset: 4090},
							label: "header",
							expr: &ruleRefExpr{
								pos:  position{line: 201, col: 14, offset: 4097},
								name: "strHeader",
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 201, col: 24, offset: 4107},
							expr: &ruleRefExpr{
								pos:  position{line: 201, col: 24, offset: 4107},
								name: "_",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 27, offset: 4110},
							name: "openCurlyBrace",
						},
						&ruleRefExpr{
							pos:  position{line: 201, col: 42, offset: 4125},
							name: "__",
						},
						&zeroOrOneExpr{
							pos: position{line: 202, col: 5, offset: 4132},
							expr: &ruleRefExpr{
								pos:  position{line: 202, col: 5, offset: 4132},
								name: "__",
							},
						},
						&labeledExpr{
							pos:   position{line: 203, col: 5, offset: 4140},
							label: "contents",
							expr: &oneOrMoreExpr{
								pos: position{line: 203, col: 14, offset: 4149},
								expr: &ruleRefExpr{
									pos:  position{line: 203, col: 14, offset: 4149},
									name: "strContents",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 204, col: 5, offset: 4166},
							expr: &ruleRefExpr{
								pos:  position{line: 204, col: 5, offset: 4166},
								name: "__",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 205, col: 5, offset: 4174},
							name: "closeCurlyBrace",
						},
					},
				},
			},
		},
		{
			name: "strHeader",
			pos:  position{line: 213, col: 1, offset: 4365},
			expr: &actionExpr{
				pos: position{line: 214, col: 7, offset: 4381},
				run: (*parser).callonstrHeader1,
				expr: &seqExpr{
					pos: position{line: 214, col: 7, offset: 4381},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 214, col: 7, offset: 4381},
							val:        "struct",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 214, col: 16, offset: 4390},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 214, col: 18, offset: 4392},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 214, col: 23, offset: 4397},
								name: "itemName",
							},
						},
					},
				},
			},
		},
		{
			name: "strContents",
			pos:  position{line: 216, col: 1, offset: 4428},
			expr: &actionExpr{
				pos: position{line: 217, col: 7, offset: 4446},
				run: (*parser).callonstrContents1,
				expr: &seqExpr{
					pos: position{line: 217, col: 7, offset: 4446},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 217, col: 7, offset: 4446},
							expr: &ruleRefExpr{
								pos:  position{line: 217, col: 7, offset: 4446},
								name: "_",
							},
						},
						&labeledExpr{
							pos:   position{line: 217, col: 10, offset: 4449},
							label: "val",
							expr: &choiceExpr{
								pos: position{line: 217, col: 15, offset: 4454},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 217, col: 15, offset: 4454},
										name: "fieldDefinition",
									},
									&ruleRefExpr{
										pos:  position{line: 218, col: 19, offset: 4488},
										name: "arrayDefinition",
									},
									&ruleRefExpr{
										pos:  position{line: 219, col: 19, offset: 4522},
										name: "comment",
									},
									&ruleRefExpr{
										pos:  position{line: 220, col: 19, offset: 4548},
										name: "str",
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 220, col: 24, offset: 4553},
							expr: &ruleRefExpr{
								pos:  position{line: 220, col: 24, offset: 4553},
								name: "__",
							},
						},
					},
				},
			},
		},
		{
			name: "pkg",
			pos:  position{line: 224, col: 1, offset: 4591},
			expr: &actionExpr{
				pos: position{line: 225, col: 7, offset: 4601},
				run: (*parser).callonpkg1,
				expr: &seqExpr{
					pos: position{line: 225, col: 7, offset: 4601},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 225, col: 7, offset: 4601},
							expr: &ruleRefExpr{
								pos:  position{line: 225, col: 7, offset: 4601},
								name: "_",
							},
						},
						&labeledExpr{
							pos:   position{line: 225, col: 10, offset: 4604},
							label: "header",
							expr: &ruleRefExpr{
								pos:  position{line: 225, col: 17, offset: 4611},
								name: "pkgHeader",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 225, col: 27, offset: 4621},
							name: "_",
						},
						&ruleRefExpr{
							pos:  position{line: 225, col: 29, offset: 4623},
							name: "openCurlyBrace",
						},
						&ruleRefExpr{
							pos:  position{line: 225, col: 44, offset: 4638},
							name: "__",
						},
						&zeroOrOneExpr{
							pos: position{line: 226, col: 5, offset: 4645},
							expr: &ruleRefExpr{
								pos:  position{line: 226, col: 5, offset: 4645},
								name: "__",
							},
						},
						&labeledExpr{
							pos:   position{line: 227, col: 5, offset: 4653},
							label: "contents",
							expr: &oneOrMoreExpr{
								pos: position{line: 227, col: 14, offset: 4662},
								expr: &ruleRefExpr{
									pos:  position{line: 227, col: 14, offset: 4662},
									name: "pkgContents",
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 228, col: 5, offset: 4679},
							expr: &ruleRefExpr{
								pos:  position{line: 228, col: 5, offset: 4679},
								name: "__",
							},
						},
						&ruleRefExpr{
							pos:  position{line: 229, col: 5, offset: 4687},
							name: "closeCurlyBrace",
						},
					},
				},
			},
		},
		{
			name: "pkgHeader",
			pos:  position{line: 237, col: 1, offset: 4845},
			expr: &actionExpr{
				pos: position{line: 238, col: 7, offset: 4861},
				run: (*parser).callonpkgHeader1,
				expr: &seqExpr{
					pos: position{line: 238, col: 7, offset: 4861},
					exprs: []interface{}{
						&litMatcher{
							pos:        position{line: 238, col: 7, offset: 4861},
							val:        "package",
							ignoreCase: false,
						},
						&ruleRefExpr{
							pos:  position{line: 238, col: 17, offset: 4871},
							name: "_",
						},
						&labeledExpr{
							pos:   position{line: 238, col: 19, offset: 4873},
							label: "name",
							expr: &ruleRefExpr{
								pos:  position{line: 238, col: 24, offset: 4878},
								name: "itemName",
							},
						},
					},
				},
			},
		},
		{
			name: "pkgContents",
			pos:  position{line: 240, col: 1, offset: 4909},
			expr: &actionExpr{
				pos: position{line: 241, col: 7, offset: 4927},
				run: (*parser).callonpkgContents1,
				expr: &seqExpr{
					pos: position{line: 241, col: 7, offset: 4927},
					exprs: []interface{}{
						&zeroOrOneExpr{
							pos: position{line: 241, col: 7, offset: 4927},
							expr: &ruleRefExpr{
								pos:  position{line: 241, col: 7, offset: 4927},
								name: "_",
							},
						},
						&labeledExpr{
							pos:   position{line: 241, col: 10, offset: 4930},
							label: "val",
							expr: &choiceExpr{
								pos: position{line: 241, col: 15, offset: 4935},
								alternatives: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 241, col: 15, offset: 4935},
										name: "idDefinition",
									},
									&ruleRefExpr{
										pos:  position{line: 242, col: 19, offset: 4966},
										name: "fieldDefinition",
									},
									&ruleRefExpr{
										pos:  position{line: 243, col: 19, offset: 5000},
										name: "arrayDefinition",
									},
									&ruleRefExpr{
										pos:  position{line: 244, col: 19, offset: 5034},
										name: "comment",
									},
									&ruleRefExpr{
										pos:  position{line: 245, col: 19, offset: 5060},
										name: "str",
									},
								},
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 245, col: 24, offset: 5065},
							expr: &ruleRefExpr{
								pos:  position{line: 245, col: 24, offset: 5065},
								name: "__",
							},
						},
					},
				},
			},
		},
	},
}

func (c *current) onstart1(val interface{}) (interface{}, error) {
	return val, nil
}

func (p *parser) callonstart1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onstart1(stack["val"])
}

func (c *current) on_1() (interface{}, error) {
	return nil, nil
}

func (p *parser) callon_1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.on_1()
}

func (c *current) on__1() (interface{}, error) {
	return nil, nil
}

func (p *parser) callon__1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.on__1()
}

func (c *current) ondigits1(digits interface{}) (interface{}, error) {
	return asString(digits), nil
}

func (p *parser) callondigits1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.ondigits1(stack["digits"])
}

func (c *current) onhexValue1(first, rest interface{}) (interface{}, error) {
	return asString(first) + asString(rest), nil
}

func (p *parser) callonhexValue1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onhexValue1(stack["first"], stack["rest"])
}

func (c *current) onitemName1(value interface{}) (interface{}, error) {
	return asString(value), nil
}

func (p *parser) callonitemName1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onitemName1(stack["value"])
}

func (c *current) oncomment1() (interface{}, error) {
	return nil, nil
}

func (p *parser) calloncomment1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.oncomment1()
}

func (c *current) onattribute1(flag interface{}) (interface{}, error) {
	return asString(flag), nil
}

func (p *parser) callonattribute1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onattribute1(stack["flag"])
}

func (c *current) onattributeList1(attr interface{}) (interface{}, error) {
	return strSlice(attr), nil
}

func (p *parser) callonattributeList1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onattributeList1(stack["attr"])
}

func (c *current) onarraySize1(val interface{}) (interface{}, error) {
	return asString(val), nil
}

func (p *parser) callonarraySize1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onarraySize1(stack["val"])
}

func (c *current) onnativeType1(val interface{}) (interface{}, error) {
	return Type{
		Name:   asString(val),
		Source: SourceNative,
	}, nil

}

func (p *parser) callonnativeType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onnativeType1(stack["val"])
}

func (c *current) onuserType1(val interface{}) (interface{}, error) {
	return Type{
		Name:   val.(string),
		Source: SourceUser,
	}, nil

}

func (p *parser) callonuserType1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onuserType1(stack["val"])
}

func (c *current) onidDefinition1(val interface{}) (interface{}, error) {
	return Object{
		ObjectType: ObjID,
		Value:      asString(val),
	}, nil

}

func (p *parser) callonidDefinition1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onidDefinition1(stack["val"])
}

func (c *current) onfieldDefinition1(t, name, attributes interface{}) (interface{}, error) {
	return Object{
		ObjectType: ObjField,
		Source:     t.(Type).Source,
		Kind:       t.(Type).Name,
		Name:       name.(string),
		Attributes: strSlice(attributes),
	}, nil

}

func (p *parser) callonfieldDefinition1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onfieldDefinition1(stack["t"], stack["name"], stack["attributes"])
}

func (c *current) onarrayDefinition1(t, size, name, attributes interface{}) (interface{}, error) {
	return Object{
		ObjectType: ObjArray,
		Source:     t.(Type).Source,
		Kind:       t.(Type).Name,
		Name:       name.(string),
		Size:       size.(string),
		Attributes: strSlice(attributes),
	}, nil

}

func (p *parser) callonarrayDefinition1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onarrayDefinition1(stack["t"], stack["size"], stack["name"], stack["attributes"])
}

func (c *current) oncontents1(val interface{}) (interface{}, error) {
	return val, nil
}

func (p *parser) calloncontents1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.oncontents1(stack["val"])
}

func (c *current) onfileContents1(val interface{}) (interface{}, error) {
	return val, nil
}

func (p *parser) callonfileContents1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onfileContents1(stack["val"])
}

func (c *current) onstr1(header, contents interface{}) (interface{}, error) {
	return Object{
		ObjectType: ObjStruct,
		Name:       header.(string),
		Contents:   objSlice(contents.([]interface{})),
	}, nil

}

func (p *parser) callonstr1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onstr1(stack["header"], stack["contents"])
}

func (c *current) onstrHeader1(name interface{}) (interface{}, error) {
	return name, nil
}

func (p *parser) callonstrHeader1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onstrHeader1(stack["name"])
}

func (c *current) onstrContents1(val interface{}) (interface{}, error) {
	return val, nil
}

func (p *parser) callonstrContents1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onstrContents1(stack["val"])
}

func (c *current) onpkg1(header, contents interface{}) (interface{}, error) {
	return Package{
		Name:     header.(string),
		Contents: objSlice(contents.([]interface{})),
	}, nil

}

func (p *parser) callonpkg1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onpkg1(stack["header"], stack["contents"])
}

func (c *current) onpkgHeader1(name interface{}) (interface{}, error) {
	return name, nil
}

func (p *parser) callonpkgHeader1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onpkgHeader1(stack["name"])
}

func (c *current) onpkgContents1(val interface{}) (interface{}, error) {
	return val, nil
}

func (p *parser) callonpkgContents1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onpkgContents1(stack["val"])
}

var (
	// errNoRule is returned when the grammar to parse has no rule.
	errNoRule = errors.New("grammar has no rule")

	// errInvalidEntrypoint is returned when the specified entrypoint rule
	// does not exit.
	errInvalidEntrypoint = errors.New("invalid entrypoint")

	// errInvalidEncoding is returned when the source is not properly
	// utf8-encoded.
	errInvalidEncoding = errors.New("invalid encoding")

	// errMaxExprCnt is used to signal that the maximum number of
	// expressions have been parsed.
	errMaxExprCnt = errors.New("max number of expresssions parsed")
)

// Option is a function that can set an option on the parser. It returns
// the previous setting as an Option.
type Option func(*parser) Option

// MaxExpressions creates an Option to stop parsing after the provided
// number of expressions have been parsed, if the value is 0 then the parser will
// parse for as many steps as needed (possibly an infinite number).
//
// The default for maxExprCnt is 0.
func MaxExpressions(maxExprCnt uint64) Option {
	return func(p *parser) Option {
		oldMaxExprCnt := p.maxExprCnt
		p.maxExprCnt = maxExprCnt
		return MaxExpressions(oldMaxExprCnt)
	}
}

// Entrypoint creates an Option to set the rule name to use as entrypoint.
// The rule name must have been specified in the -alternate-entrypoints
// if generating the parser with the -optimize-grammar flag, otherwise
// it may have been optimized out. Passing an empty string sets the
// entrypoint to the first rule in the grammar.
//
// The default is to start parsing at the first rule in the grammar.
func Entrypoint(ruleName string) Option {
	return func(p *parser) Option {
		oldEntrypoint := p.entrypoint
		p.entrypoint = ruleName
		if ruleName == "" {
			p.entrypoint = g.rules[0].name
		}
		return Entrypoint(oldEntrypoint)
	}
}

// Statistics adds a user provided Stats struct to the parser to allow
// the user to process the results after the parsing has finished.
// Also the key for the "no match" counter is set.
//
// Example usage:
//
//     input := "input"
//     stats := Stats{}
//     _, err := Parse("input-file", []byte(input), Statistics(&stats, "no match"))
//     if err != nil {
//         log.Panicln(err)
//     }
//     b, err := json.MarshalIndent(stats.ChoiceAltCnt, "", "  ")
//     if err != nil {
//         log.Panicln(err)
//     }
//     fmt.Println(string(b))
//
func Statistics(stats *Stats, choiceNoMatch string) Option {
	return func(p *parser) Option {
		oldStats := p.Stats
		p.Stats = stats
		oldChoiceNoMatch := p.choiceNoMatch
		p.choiceNoMatch = choiceNoMatch
		if p.Stats.ChoiceAltCnt == nil {
			p.Stats.ChoiceAltCnt = make(map[string]map[string]int)
		}
		return Statistics(oldStats, oldChoiceNoMatch)
	}
}

// Debug creates an Option to set the debug flag to b. When set to true,
// debugging information is printed to stdout while parsing.
//
// The default is false.
func Debug(b bool) Option {
	return func(p *parser) Option {
		old := p.debug
		p.debug = b
		return Debug(old)
	}
}

// Memoize creates an Option to set the memoize flag to b. When set to true,
// the parser will cache all results so each expression is evaluated only
// once. This guarantees linear parsing time even for pathological cases,
// at the expense of more memory and slower times for typical cases.
//
// The default is false.
func Memoize(b bool) Option {
	return func(p *parser) Option {
		old := p.memoize
		p.memoize = b
		return Memoize(old)
	}
}

// AllowInvalidUTF8 creates an Option to allow invalid UTF-8 bytes.
// Every invalid UTF-8 byte is treated as a utf8.RuneError (U+FFFD)
// by character class matchers and is matched by the any matcher.
// The returned matched value, c.text and c.offset are NOT affected.
//
// The default is false.
func AllowInvalidUTF8(b bool) Option {
	return func(p *parser) Option {
		old := p.allowInvalidUTF8
		p.allowInvalidUTF8 = b
		return AllowInvalidUTF8(old)
	}
}

// Recover creates an Option to set the recover flag to b. When set to
// true, this causes the parser to recover from panics and convert it
// to an error. Setting it to false can be useful while debugging to
// access the full stack trace.
//
// The default is true.
func Recover(b bool) Option {
	return func(p *parser) Option {
		old := p.recover
		p.recover = b
		return Recover(old)
	}
}

// GlobalStore creates an Option to set a key to a certain value in
// the globalStore.
func GlobalStore(key string, value interface{}) Option {
	return func(p *parser) Option {
		old := p.cur.globalStore[key]
		p.cur.globalStore[key] = value
		return GlobalStore(key, old)
	}
}

// InitState creates an Option to set a key to a certain value in
// the global "state" store.
func InitState(key string, value interface{}) Option {
	return func(p *parser) Option {
		old := p.cur.state[key]
		p.cur.state[key] = value
		return InitState(key, old)
	}
}

// ParseFile parses the file identified by filename.
func ParseFile(filename string, opts ...Option) (i interface{}, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = closeErr
		}
	}()
	return ParseReader(filename, f, opts...)
}

// ParseReader parses the data from r using filename as information in the
// error messages.
func ParseReader(filename string, r io.Reader, opts ...Option) (interface{}, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return Parse(filename, b, opts...)
}

// Parse parses the data from b using filename as information in the
// error messages.
func Parse(filename string, b []byte, opts ...Option) (interface{}, error) {
	return newParser(filename, b, opts...).parse(g)
}

// position records a position in the text.
type position struct {
	line, col, offset int
}

func (p position) String() string {
	return fmt.Sprintf("%d:%d [%d]", p.line, p.col, p.offset)
}

// savepoint stores all state required to go back to this point in the
// parser.
type savepoint struct {
	position
	rn rune
	w  int
}

type current struct {
	pos  position // start position of the match
	text []byte   // raw text of the match

	// state is a store for arbitrary key,value pairs that the user wants to be
	// tied to the backtracking of the parser.
	// This is always rolled back if a parsing rule fails.
	state storeDict

	// globalStore is a general store for the user to store arbitrary key-value
	// pairs that they need to manage and that they do not want tied to the
	// backtracking of the parser. This is only modified by the user and never
	// rolled back by the parser. It is always up to the user to keep this in a
	// consistent state.
	globalStore storeDict
}

type storeDict map[string]interface{}

// the AST types...

type grammar struct {
	pos   position
	rules []*rule
}

type rule struct {
	pos         position
	name        string
	displayName string
	expr        interface{}
}

type choiceExpr struct {
	pos          position
	alternatives []interface{}
}

type actionExpr struct {
	pos  position
	expr interface{}
	run  func(*parser) (interface{}, error)
}

type recoveryExpr struct {
	pos          position
	expr         interface{}
	recoverExpr  interface{}
	failureLabel []string
}

type seqExpr struct {
	pos   position
	exprs []interface{}
}

type throwExpr struct {
	pos   position
	label string
}

type labeledExpr struct {
	pos   position
	label string
	expr  interface{}
}

type expr struct {
	pos  position
	expr interface{}
}

type andExpr expr
type notExpr expr
type zeroOrOneExpr expr
type zeroOrMoreExpr expr
type oneOrMoreExpr expr

type ruleRefExpr struct {
	pos  position
	name string
}

type stateCodeExpr struct {
	pos position
	run func(*parser) error
}

type andCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type notCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type litMatcher struct {
	pos        position
	val        string
	ignoreCase bool
}

type charClassMatcher struct {
	pos             position
	val             string
	basicLatinChars [128]bool
	chars           []rune
	ranges          []rune
	classes         []*unicode.RangeTable
	ignoreCase      bool
	inverted        bool
}

type anyMatcher position

// errList cumulates the errors found by the parser.
type errList []error

func (e *errList) add(err error) {
	*e = append(*e, err)
}

func (e errList) err() error {
	if len(e) == 0 {
		return nil
	}
	e.dedupe()
	return e
}

func (e *errList) dedupe() {
	var cleaned []error
	set := make(map[string]bool)
	for _, err := range *e {
		if msg := err.Error(); !set[msg] {
			set[msg] = true
			cleaned = append(cleaned, err)
		}
	}
	*e = cleaned
}

func (e errList) Error() string {
	switch len(e) {
	case 0:
		return ""
	case 1:
		return e[0].Error()
	default:
		var buf bytes.Buffer

		for i, err := range e {
			if i > 0 {
				buf.WriteRune('\n')
			}
			buf.WriteString(err.Error())
		}
		return buf.String()
	}
}

// parserError wraps an error with a prefix indicating the rule in which
// the error occurred. The original error is stored in the Inner field.
type parserError struct {
	Inner    error
	pos      position
	prefix   string
	expected []string
}

// Error returns the error message.
func (p *parserError) Error() string {
	return p.prefix + ": " + p.Inner.Error()
}

// newParser creates a parser with the specified input source and options.
func newParser(filename string, b []byte, opts ...Option) *parser {
	stats := Stats{
		ChoiceAltCnt: make(map[string]map[string]int),
	}

	p := &parser{
		filename: filename,
		errs:     new(errList),
		data:     b,
		pt:       savepoint{position: position{line: 1}},
		recover:  true,
		cur: current{
			state:       make(storeDict),
			globalStore: make(storeDict),
		},
		maxFailPos:      position{col: 1, line: 1},
		maxFailExpected: make([]string, 0, 20),
		Stats:           &stats,
		// start rule is rule [0] unless an alternate entrypoint is specified
		entrypoint: g.rules[0].name,
		emptyState: make(storeDict),
	}
	p.setOptions(opts)

	if p.maxExprCnt == 0 {
		p.maxExprCnt = math.MaxUint64
	}

	return p
}

// setOptions applies the options to the parser.
func (p *parser) setOptions(opts []Option) {
	for _, opt := range opts {
		opt(p)
	}
}

type resultTuple struct {
	v   interface{}
	b   bool
	end savepoint
}

const choiceNoMatch = -1

// Stats stores some statistics, gathered during parsing
type Stats struct {
	// ExprCnt counts the number of expressions processed during parsing
	// This value is compared to the maximum number of expressions allowed
	// (set by the MaxExpressions option).
	ExprCnt uint64

	// ChoiceAltCnt is used to count for each ordered choice expression,
	// which alternative is used how may times.
	// These numbers allow to optimize the order of the ordered choice expression
	// to increase the performance of the parser
	//
	// The outer key of ChoiceAltCnt is composed of the name of the rule as well
	// as the line and the column of the ordered choice.
	// The inner key of ChoiceAltCnt is the number (one-based) of the matching alternative.
	// For each alternative the number of matches are counted. If an ordered choice does not
	// match, a special counter is incremented. The name of this counter is set with
	// the parser option Statistics.
	// For an alternative to be included in ChoiceAltCnt, it has to match at least once.
	ChoiceAltCnt map[string]map[string]int
}

type parser struct {
	filename string
	pt       savepoint
	cur      current

	data []byte
	errs *errList

	depth   int
	recover bool
	debug   bool

	memoize bool
	// memoization table for the packrat algorithm:
	// map[offset in source] map[expression or rule] {value, match}
	memo map[int]map[interface{}]resultTuple

	// rules table, maps the rule identifier to the rule node
	rules map[string]*rule
	// variables stack, map of label to value
	vstack []map[string]interface{}
	// rule stack, allows identification of the current rule in errors
	rstack []*rule

	// parse fail
	maxFailPos            position
	maxFailExpected       []string
	maxFailInvertExpected bool

	// max number of expressions to be parsed
	maxExprCnt uint64
	// entrypoint for the parser
	entrypoint string

	allowInvalidUTF8 bool

	*Stats

	choiceNoMatch string
	// recovery expression stack, keeps track of the currently available recovery expression, these are traversed in reverse
	recoveryStack []map[string]interface{}

	// emptyState contains an empty storeDict, which is used to optimize cloneState if global "state" store is not used.
	emptyState storeDict
}

// push a variable set on the vstack.
func (p *parser) pushV() {
	if cap(p.vstack) == len(p.vstack) {
		// create new empty slot in the stack
		p.vstack = append(p.vstack, nil)
	} else {
		// slice to 1 more
		p.vstack = p.vstack[:len(p.vstack)+1]
	}

	// get the last args set
	m := p.vstack[len(p.vstack)-1]
	if m != nil && len(m) == 0 {
		// empty map, all good
		return
	}

	m = make(map[string]interface{})
	p.vstack[len(p.vstack)-1] = m
}

// pop a variable set from the vstack.
func (p *parser) popV() {
	// if the map is not empty, clear it
	m := p.vstack[len(p.vstack)-1]
	if len(m) > 0 {
		// GC that map
		p.vstack[len(p.vstack)-1] = nil
	}
	p.vstack = p.vstack[:len(p.vstack)-1]
}

// push a recovery expression with its labels to the recoveryStack
func (p *parser) pushRecovery(labels []string, expr interface{}) {
	if cap(p.recoveryStack) == len(p.recoveryStack) {
		// create new empty slot in the stack
		p.recoveryStack = append(p.recoveryStack, nil)
	} else {
		// slice to 1 more
		p.recoveryStack = p.recoveryStack[:len(p.recoveryStack)+1]
	}

	m := make(map[string]interface{}, len(labels))
	for _, fl := range labels {
		m[fl] = expr
	}
	p.recoveryStack[len(p.recoveryStack)-1] = m
}

// pop a recovery expression from the recoveryStack
func (p *parser) popRecovery() {
	// GC that map
	p.recoveryStack[len(p.recoveryStack)-1] = nil

	p.recoveryStack = p.recoveryStack[:len(p.recoveryStack)-1]
}

func (p *parser) print(prefix, s string) string {
	if !p.debug {
		return s
	}

	fmt.Printf("%s %d:%d:%d: %s [%#U]\n",
		prefix, p.pt.line, p.pt.col, p.pt.offset, s, p.pt.rn)
	return s
}

func (p *parser) in(s string) string {
	p.depth++
	return p.print(strings.Repeat(" ", p.depth)+">", s)
}

func (p *parser) out(s string) string {
	p.depth--
	return p.print(strings.Repeat(" ", p.depth)+"<", s)
}

func (p *parser) addErr(err error) {
	p.addErrAt(err, p.pt.position, []string{})
}

func (p *parser) addErrAt(err error, pos position, expected []string) {
	var buf bytes.Buffer
	if p.filename != "" {
		buf.WriteString(p.filename)
	}
	if buf.Len() > 0 {
		buf.WriteString(":")
	}
	buf.WriteString(fmt.Sprintf("%d:%d (%d)", pos.line, pos.col, pos.offset))
	if len(p.rstack) > 0 {
		if buf.Len() > 0 {
			buf.WriteString(": ")
		}
		rule := p.rstack[len(p.rstack)-1]
		if rule.displayName != "" {
			buf.WriteString("rule " + rule.displayName)
		} else {
			buf.WriteString("rule " + rule.name)
		}
	}
	pe := &parserError{Inner: err, pos: pos, prefix: buf.String(), expected: expected}
	p.errs.add(pe)
}

func (p *parser) failAt(fail bool, pos position, want string) {
	// process fail if parsing fails and not inverted or parsing succeeds and invert is set
	if fail == p.maxFailInvertExpected {
		if pos.offset < p.maxFailPos.offset {
			return
		}

		if pos.offset > p.maxFailPos.offset {
			p.maxFailPos = pos
			p.maxFailExpected = p.maxFailExpected[:0]
		}

		if p.maxFailInvertExpected {
			want = "!" + want
		}
		p.maxFailExpected = append(p.maxFailExpected, want)
	}
}

// read advances the parser to the next rune.
func (p *parser) read() {
	p.pt.offset += p.pt.w
	rn, n := utf8.DecodeRune(p.data[p.pt.offset:])
	p.pt.rn = rn
	p.pt.w = n
	p.pt.col++
	if rn == '\n' {
		p.pt.line++
		p.pt.col = 0
	}

	if rn == utf8.RuneError && n == 1 { // see utf8.DecodeRune
		if !p.allowInvalidUTF8 {
			p.addErr(errInvalidEncoding)
		}
	}
}

// restore parser position to the savepoint pt.
func (p *parser) restore(pt savepoint) {
	if p.debug {
		defer p.out(p.in("restore"))
	}
	if pt.offset == p.pt.offset {
		return
	}
	p.pt = pt
}

// Cloner is implemented by any value that has a Clone method, which returns a
// copy of the value. This is mainly used for types which are not passed by
// value (e.g map, slice, chan) or structs that contain such types.
//
// This is used in conjunction with the global state feature to create proper
// copies of the state to allow the parser to properly restore the state in
// the case of backtracking.
type Cloner interface {
	Clone() interface{}
}

// clone and return parser current state.
func (p *parser) cloneState() storeDict {
	if p.debug {
		defer p.out(p.in("cloneState"))
	}

	if len(p.cur.state) == 0 {
		return p.emptyState
	}

	state := make(storeDict, len(p.cur.state))
	for k, v := range p.cur.state {
		if c, ok := v.(Cloner); ok {
			state[k] = c.Clone()
		} else {
			state[k] = v
		}
	}
	return state
}

// restore parser current state to the state storeDict.
// every restoreState should applied only one time for every cloned state
func (p *parser) restoreState(state storeDict) {
	if p.debug {
		defer p.out(p.in("restoreState"))
	}
	p.cur.state = state
}

// get the slice of bytes from the savepoint start to the current position.
func (p *parser) sliceFrom(start savepoint) []byte {
	return p.data[start.position.offset:p.pt.position.offset]
}

func (p *parser) getMemoized(node interface{}) (resultTuple, bool) {
	if len(p.memo) == 0 {
		return resultTuple{}, false
	}
	m := p.memo[p.pt.offset]
	if len(m) == 0 {
		return resultTuple{}, false
	}
	res, ok := m[node]
	return res, ok
}

func (p *parser) setMemoized(pt savepoint, node interface{}, tuple resultTuple) {
	if p.memo == nil {
		p.memo = make(map[int]map[interface{}]resultTuple)
	}
	m := p.memo[pt.offset]
	if m == nil {
		m = make(map[interface{}]resultTuple)
		p.memo[pt.offset] = m
	}
	m[node] = tuple
}

func (p *parser) buildRulesTable(g *grammar) {
	p.rules = make(map[string]*rule, len(g.rules))
	for _, r := range g.rules {
		p.rules[r.name] = r
	}
}

func (p *parser) parse(g *grammar) (val interface{}, err error) {
	if len(g.rules) == 0 {
		p.addErr(errNoRule)
		return nil, p.errs.err()
	}

	// TODO : not super critical but this could be generated
	p.buildRulesTable(g)

	if p.recover {
		// panic can be used in action code to stop parsing immediately
		// and return the panic as an error.
		defer func() {
			if e := recover(); e != nil {
				if p.debug {
					defer p.out(p.in("panic handler"))
				}
				val = nil
				switch e := e.(type) {
				case error:
					p.addErr(e)
				default:
					p.addErr(fmt.Errorf("%v", e))
				}
				err = p.errs.err()
			}
		}()
	}

	startRule, ok := p.rules[p.entrypoint]
	if !ok {
		p.addErr(errInvalidEntrypoint)
		return nil, p.errs.err()
	}

	p.read() // advance to first rune
	val, ok = p.parseRule(startRule)
	if !ok {
		if len(*p.errs) == 0 {
			// If parsing fails, but no errors have been recorded, the expected values
			// for the farthest parser position are returned as error.
			maxFailExpectedMap := make(map[string]struct{}, len(p.maxFailExpected))
			for _, v := range p.maxFailExpected {
				maxFailExpectedMap[v] = struct{}{}
			}
			expected := make([]string, 0, len(maxFailExpectedMap))
			eof := false
			if _, ok := maxFailExpectedMap["!."]; ok {
				delete(maxFailExpectedMap, "!.")
				eof = true
			}
			for k := range maxFailExpectedMap {
				expected = append(expected, k)
			}
			sort.Strings(expected)
			if eof {
				expected = append(expected, "EOF")
			}
			p.addErrAt(errors.New("no match found, expected: "+listJoin(expected, ", ", "or")), p.maxFailPos, expected)
		}

		return nil, p.errs.err()
	}
	return val, p.errs.err()
}

func listJoin(list []string, sep string, lastSep string) string {
	switch len(list) {
	case 0:
		return ""
	case 1:
		return list[0]
	default:
		return fmt.Sprintf("%s %s %s", strings.Join(list[:len(list)-1], sep), lastSep, list[len(list)-1])
	}
}

func (p *parser) parseRule(rule *rule) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRule " + rule.name))
	}

	if p.memoize {
		res, ok := p.getMemoized(rule)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
	}

	start := p.pt
	p.rstack = append(p.rstack, rule)
	p.pushV()
	val, ok := p.parseExpr(rule.expr)
	p.popV()
	p.rstack = p.rstack[:len(p.rstack)-1]
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}

	if p.memoize {
		p.setMemoized(start, rule, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseExpr(expr interface{}) (interface{}, bool) {
	var pt savepoint

	if p.memoize {
		res, ok := p.getMemoized(expr)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
		pt = p.pt
	}

	p.ExprCnt++
	if p.ExprCnt > p.maxExprCnt {
		panic(errMaxExprCnt)
	}

	var val interface{}
	var ok bool
	switch expr := expr.(type) {
	case *actionExpr:
		val, ok = p.parseActionExpr(expr)
	case *andCodeExpr:
		val, ok = p.parseAndCodeExpr(expr)
	case *andExpr:
		val, ok = p.parseAndExpr(expr)
	case *anyMatcher:
		val, ok = p.parseAnyMatcher(expr)
	case *charClassMatcher:
		val, ok = p.parseCharClassMatcher(expr)
	case *choiceExpr:
		val, ok = p.parseChoiceExpr(expr)
	case *labeledExpr:
		val, ok = p.parseLabeledExpr(expr)
	case *litMatcher:
		val, ok = p.parseLitMatcher(expr)
	case *notCodeExpr:
		val, ok = p.parseNotCodeExpr(expr)
	case *notExpr:
		val, ok = p.parseNotExpr(expr)
	case *oneOrMoreExpr:
		val, ok = p.parseOneOrMoreExpr(expr)
	case *recoveryExpr:
		val, ok = p.parseRecoveryExpr(expr)
	case *ruleRefExpr:
		val, ok = p.parseRuleRefExpr(expr)
	case *seqExpr:
		val, ok = p.parseSeqExpr(expr)
	case *stateCodeExpr:
		val, ok = p.parseStateCodeExpr(expr)
	case *throwExpr:
		val, ok = p.parseThrowExpr(expr)
	case *zeroOrMoreExpr:
		val, ok = p.parseZeroOrMoreExpr(expr)
	case *zeroOrOneExpr:
		val, ok = p.parseZeroOrOneExpr(expr)
	default:
		panic(fmt.Sprintf("unknown expression type %T", expr))
	}
	if p.memoize {
		p.setMemoized(pt, expr, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseActionExpr(act *actionExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseActionExpr"))
	}

	start := p.pt
	val, ok := p.parseExpr(act.expr)
	if ok {
		p.cur.pos = start.position
		p.cur.text = p.sliceFrom(start)
		state := p.cloneState()
		actVal, err := act.run(p)
		if err != nil {
			p.addErrAt(err, start.position, []string{})
		}
		p.restoreState(state)

		val = actVal
	}
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}
	return val, ok
}

func (p *parser) parseAndCodeExpr(and *andCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndCodeExpr"))
	}

	state := p.cloneState()

	ok, err := and.run(p)
	if err != nil {
		p.addErr(err)
	}
	p.restoreState(state)

	return nil, ok
}

func (p *parser) parseAndExpr(and *andExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(and.expr)
	p.popV()
	p.restore(pt)
	return nil, ok
}

func (p *parser) parseAnyMatcher(any *anyMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAnyMatcher"))
	}

	if p.pt.rn == utf8.RuneError && p.pt.w == 0 {
		// EOF - see utf8.DecodeRune
		p.failAt(false, p.pt.position, ".")
		return nil, false
	}
	start := p.pt
	p.read()
	p.failAt(true, start.position, ".")
	return p.sliceFrom(start), true
}

func (p *parser) parseCharClassMatcher(chr *charClassMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseCharClassMatcher"))
	}

	cur := p.pt.rn
	start := p.pt

	// can't match EOF
	if cur == utf8.RuneError && p.pt.w == 0 { // see utf8.DecodeRune
		p.failAt(false, start.position, chr.val)
		return nil, false
	}

	if chr.ignoreCase {
		cur = unicode.ToLower(cur)
	}

	// try to match in the list of available chars
	for _, rn := range chr.chars {
		if rn == cur {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of ranges
	for i := 0; i < len(chr.ranges); i += 2 {
		if cur >= chr.ranges[i] && cur <= chr.ranges[i+1] {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of Unicode classes
	for _, cl := range chr.classes {
		if unicode.Is(cl, cur) {
			if chr.inverted {
				p.failAt(false, start.position, chr.val)
				return nil, false
			}
			p.read()
			p.failAt(true, start.position, chr.val)
			return p.sliceFrom(start), true
		}
	}

	if chr.inverted {
		p.read()
		p.failAt(true, start.position, chr.val)
		return p.sliceFrom(start), true
	}
	p.failAt(false, start.position, chr.val)
	return nil, false
}

func (p *parser) incChoiceAltCnt(ch *choiceExpr, altI int) {
	choiceIdent := fmt.Sprintf("%s %d:%d", p.rstack[len(p.rstack)-1].name, ch.pos.line, ch.pos.col)
	m := p.ChoiceAltCnt[choiceIdent]
	if m == nil {
		m = make(map[string]int)
		p.ChoiceAltCnt[choiceIdent] = m
	}
	// We increment altI by 1, so the keys do not start at 0
	alt := strconv.Itoa(altI + 1)
	if altI == choiceNoMatch {
		alt = p.choiceNoMatch
	}
	m[alt]++
}

func (p *parser) parseChoiceExpr(ch *choiceExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseChoiceExpr"))
	}

	for altI, alt := range ch.alternatives {
		// dummy assignment to prevent compile error if optimized
		_ = altI

		state := p.cloneState()
		p.pushV()
		val, ok := p.parseExpr(alt)
		p.popV()
		if ok {
			p.incChoiceAltCnt(ch, altI)
			return val, ok
		}
		p.restoreState(state)
	}
	p.incChoiceAltCnt(ch, choiceNoMatch)
	return nil, false
}

func (p *parser) parseLabeledExpr(lab *labeledExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLabeledExpr"))
	}

	p.pushV()
	val, ok := p.parseExpr(lab.expr)
	p.popV()
	if ok && lab.label != "" {
		m := p.vstack[len(p.vstack)-1]
		m[lab.label] = val
	}
	return val, ok
}

func (p *parser) parseLitMatcher(lit *litMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLitMatcher"))
	}

	ignoreCase := ""
	if lit.ignoreCase {
		ignoreCase = "i"
	}
	val := fmt.Sprintf("%q%s", lit.val, ignoreCase)
	start := p.pt
	for _, want := range lit.val {
		cur := p.pt.rn
		if lit.ignoreCase {
			cur = unicode.ToLower(cur)
		}
		if cur != want {
			p.failAt(false, start.position, val)
			p.restore(start)
			return nil, false
		}
		p.read()
	}
	p.failAt(true, start.position, val)
	return p.sliceFrom(start), true
}

func (p *parser) parseNotCodeExpr(not *notCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotCodeExpr"))
	}

	state := p.cloneState()

	ok, err := not.run(p)
	if err != nil {
		p.addErr(err)
	}
	p.restoreState(state)

	return nil, !ok
}

func (p *parser) parseNotExpr(not *notExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotExpr"))
	}

	pt := p.pt
	p.pushV()
	p.maxFailInvertExpected = !p.maxFailInvertExpected
	_, ok := p.parseExpr(not.expr)
	p.maxFailInvertExpected = !p.maxFailInvertExpected
	p.popV()
	p.restore(pt)
	return nil, !ok
}

func (p *parser) parseOneOrMoreExpr(expr *oneOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseOneOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			if len(vals) == 0 {
				// did not match once, no match
				return nil, false
			}
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseRecoveryExpr(recover *recoveryExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRecoveryExpr (" + strings.Join(recover.failureLabel, ",") + ")"))
	}

	p.pushRecovery(recover.failureLabel, recover.recoverExpr)
	val, ok := p.parseExpr(recover.expr)
	p.popRecovery()

	return val, ok
}

func (p *parser) parseRuleRefExpr(ref *ruleRefExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRuleRefExpr " + ref.name))
	}

	if ref.name == "" {
		panic(fmt.Sprintf("%s: invalid rule: missing name", ref.pos))
	}

	rule := p.rules[ref.name]
	if rule == nil {
		p.addErr(fmt.Errorf("undefined rule: %s", ref.name))
		return nil, false
	}
	return p.parseRule(rule)
}

func (p *parser) parseSeqExpr(seq *seqExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseSeqExpr"))
	}

	vals := make([]interface{}, 0, len(seq.exprs))

	pt := p.pt
	for _, expr := range seq.exprs {
		val, ok := p.parseExpr(expr)
		if !ok {
			p.restore(pt)
			return nil, false
		}
		vals = append(vals, val)
	}
	return vals, true
}

func (p *parser) parseStateCodeExpr(state *stateCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseStateCodeExpr"))
	}

	err := state.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, true
}

func (p *parser) parseThrowExpr(expr *throwExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseThrowExpr"))
	}

	for i := len(p.recoveryStack) - 1; i >= 0; i-- {
		if recoverExpr, ok := p.recoveryStack[i][expr.label]; ok {
			if val, ok := p.parseExpr(recoverExpr); ok {
				return val, ok
			}
		}
	}

	return nil, false
}

func (p *parser) parseZeroOrMoreExpr(expr *zeroOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseZeroOrOneExpr(expr *zeroOrOneExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrOneExpr"))
	}

	p.pushV()
	val, _ := p.parseExpr(expr.expr)
	p.popV()
	// whether it matched or not, consider it a match
	return val, true
}
