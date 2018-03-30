/**
  Simple initial blocks and types
**/

start
    = contents

whitespace
    = [ \t]

EOL
    = [ \t\r\n]+ comment?

EOF
    = !.

_ "whitespace"
    = whitespace*

__ "eol"
    = EOL*

digit
    = [0-9]

digits
    = digits:digit* { return digits.join('') }

hex_value
    = first:"0x" rest:[a-fA-F0-9]+ { return first + rest.join('') }

item_name
    = value:[a-z_]+ { return value.join('') }

open_curly_brace
    = "{"

close_curly_brace
    = "}"

open_square_brace
    = "["

close_square_brace
    = "]"

comment
    = "//" [^\n]* (EOL/EOF)? {}

attribute
    = _ "!" flag:("deprecated") { return flag }

attribute_list
    = attr:attribute+ { return attr }

// Language basic definitions

array_size
    = open_square_brace val:("*" / digits) close_square_brace { return val }

native_type
    = val:("dynint" / "uint8" / "uint32" / "uint64" / "byte" / "double" / "string" / "blob" / "bool" / "uuid" / "array" / "any") { return { source: "native", name: val } }

user_type
    = "@" val:item_name { return { source: "user", name: val } }

id_definition
    = "id" _ val:hex_value { return { object_type: "identifier", value: val } }

field_definition
    = type:(native_type / user_type) _ name:item_name attributes:attribute_list? { return { object_type: "field", source: type.source, kind: type.name, name: name, attributes: attributes } }

array_definition
    = type:(native_type / user_type) size:array_size _ name:item_name attributes:attribute_list? { return { object_type: "array", source: type.source, kind: type.name, name: name, size: size, attributes: attributes } }

// Language structures

contents
    = val:file_contents+ { return val }

file_contents
    = __? comment? val:pkg __? { return val }

// Structures

str
    = header:str_header _? open_curly_brace __
    __?
    content:str_contents+
    __?
    close_curly_brace { return { "object_type": "struct", name: header, contents: content } }

str_header
    = "struct" _ name:item_name { return name }

str_contents
    = _? val:(field_definition
            / array_definition
            / comment
            / str) __? { return val }

// Packages

pkg
    = _? header:pkg_header _ open_curly_brace __
        __?
        contents:pkg_contents+
        __?
        close_curly_brace { return { name: header, contents: contents } }


pkg_header
    = "package" _ name:item_name { return name }

pkg_contents
    = _? val:(id_definition
            / field_definition
            / array_definition
            / comment
            / str) __? { return val }
