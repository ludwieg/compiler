# ludco
> `ludwieg/ludco` provides facilities to check and generate sources based on `.lud` files

#### Prerelease software
Ludwieg is WIP and is not ready for production environments. Use at your own risk.


## The Ludiweg format
Ludwieg is a lightweight contractual binary protocol used to exchange data
between systems. Cross-platform support is available and can be easily
implemented.

 - Ludwieg generates code based on input files, allowing easily implementation
against the framework.
 - Ludwieg is future-proof, allowing safe deprecation, and unknown fields.

A protocol file, with the `.lud` extension provides can provide several
top-level packages and children structures. This allows a fine-grained
implementation of communication protocols. An example of a Ludwieg package
carrying user authentication information is:

```
package authentication {
    id 0x01

    string      username
    string      password
}
```

The following patterns can be noted in the example above:
 - All identifiers are lowercase. `ludco` will treat violations of this rule as
errors.
 - Every package contains an identifier. It is used to identify the package
 globally among other packages of your protocol. `id`s must be unique.


### Submodules

Ludwieg allows packages to define custom substructures. This is done using the
`structure` keyword. Assuming you want to exchange information of users that
signed up for your service on a specific date, the following structure could be
used as a response to that query:

```
package users {
    id 0x02

    uint64      date
    @entry[*]   users

    struct entry {
        username    string
        email       string
    }
}
```

On this new example, we can notice a few new elements:
 - The array definition: `[*]`. Any type (except `any`) can be declared as an
array, and arrays can be declared as having a fixed size, such as `[10]`, or
`[27]`, or having a dynamic size, throught the `[*]` notation. Although
frameworks does not enforce maximum sizing, information might be used to assist
memory allocation mechanisms, improving runtime speed.
 - The custom type notation: `@entry`. `struct`s are referenced this way when
being used by fields.

### Organization
`ludco` uses an input folder to read definition files (`.lud`) and validate your
protocol. This measure is used to allow the tool to check for `id` clashes, and
compile the defined protocol in one step. The utility also takes an output
folder to store compilation results, generating code in the language provided
through the command line.

## Using the ludco utility
`ludco` can perform two actions: `show` (`s`) or `compile` (`c`).

`show` loads, parses, and validates all definition files found inside the
provided directory, and outputs a visual representation of your packages,
fields, and structures. For demonstration purposes, assume the `authentication`
package has been defined in a file named `authentication.lud`, in `~/ludwieg`.
A visual representation of the defined structures can be obtained by running the
following command:

```
$ ludco s ~/ludwieg

authentication (0x01)
└── Fields
    ├── [0] string username
    └── [1] string password

```

ludco outputs all found packages, their `id`, and their indexed fields, types,
and any other annotation.

`compile` loads, parses, validates, and generates code on the provided language.
The following languages can be used:
 - `objc` for Objective-C
 - `java` for Java
 - `go` for Golang

Provided flags depends on the choosen language:

### Objective-C

When generating Objective-C files, the following command is invoked:
```
$ ludco c InputFolder OutputFolder --lang objc --prefix Prefix
```

The first positional argument is the input folder, where definition files are
located. The second, the output folder, where `ludco` will write output files (
the folder is automatically created, if it does not exist). Then, `--lang`
defines the target language for code generation, whilst `--prefix` is used as
the prefix for all your classes.

After the process is completed, integration information is then presented.
Follow those instructions to fully integrate the generated sources to your
project.

> **Notice**: `--prefix` is not required, altough highly recommended by Apple,
> based on their code convetions:
>
> Your own classes should use three letter prefixes. These might relate to a
> combination of your company name and your app name, or even a specific
> component within your app. As an example, if your company were called
> Whispering Oak, and you were developing a game called Zebra Surprise, you
> might choose `WZS` or `WOZ` as your class prefix.

### Golang

When generating Golang files, the following command is invoked:
```
$ ludco c InputFolder OutputFolder --lang objc --package PackageName
```

The first positional argument is the input folder, where definition files are
located. The second, the output folder, where `ludco` will write output files (
the folder is automatically created, if it does not exist). Then, `--lang`
defines the target language for code generation, whilst `--package` indicates
to which Golang package generated code belongs to. If omitted, `ludco` uses
a normalized version of the output directory's name.

Differently from the Objective-C and Java generators, Golang does not require
any dependency to be installed, as a special file named `ludwieg_base.go` will
automatically be included among the generated sources. This file contains the
basic mechanisms for encoding/decoding package information.

### Java
When generating Java files, the following command is invoked:
```
$ ludco c InputFolder OutputFolder --lang java --package Package
```

The first positional argument is the input folder, where definition files are
located. The second, the output folder, where `ludco` will write output files (
the folder is automatically created, if it does not exist). Then, `--lang`
defines the target language for code generation, whilst `--package` indicates
to which Java package generated code belongs to. If omitted, `ludco` uses
a normalized version of the output directory's name.

After the process is completed, integration information is then presented.
Follow those instructions to fully integrate the generated sources to your
project.

> **Notice**: `ludco` will not create folder structure based on package names,
> such as `com.example.project` even when `--package` is provided.

## License

```
MIT License

Copyright (c) 2017 Victor Gama de Oliveira

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
