#!/bin/bash
# Gocat concatenates golang sources of a same package into a single file.
#
# Due to lack of time, and perhaps means (and patience to read and deal with
# Go's ast, parser, printer, and token packages), this is a dirty hack that
# concatenates all Go base sources, strips all imports, and write everything
# back again, under the package name provided through the first positional
# argument passed to this script. It then attempts to `go fmt` it, rolling
# everything back in case it fails.
#
# Parameter order is:
# 1. Input folder
# 2. Output destination (will create a file named "ludwieg_base.go", replacing
#    any existing files)
# 3. Package name
#
# No checks are performed, omitting any param will cause undefined behaviour.
# Depends on uuidgen, sed, gawk and perl
#
# Released under MIT License, copyright (c) 2017 - Victor Gama <hey@vito.io>

function abspath {
    if [[ -d "$1" ]]; then
        pushd "$1" >/dev/null
        pwd
        popd >/dev/null
    elif [[ -e $1 ]]; then
        pushd "$(dirname "$1")" >/dev/null
        echo "$(pwd)/$(basename "$1")"
        popd >/dev/null
    else
        echo "$1" does not exist! >&2
        return 127
    fi
}

function checkDep {
    printf "Checking dependency: \"$1\"... "
    command -v $1 2>&1 1>/dev/null
    if [[ $? != 0 ]]; then
        printf "FAILED\n"
        echo "gocat depends on $1, which is not available on this system"
        echo "please use your package manager to install it."
        exit 1
    else
        printf "OK\n"
    fi
}

checkDep "uuidgen"
checkDep "sed"
checkDep "gawk"
checkDep "perl"

inputFolder="$(abspath $1)/*.go"
outputPath="$(abspath $2)/ludwieg_base.go"
pkgName=${3:-"impl"}

name=$(uuidgen | sed 's/-//g')
tmp="/tmp/$name.go"

echo "Processings source files..."
simple=$(cat $inputFolder | gawk '/^import ("[^"]+")$/ { print $2 }')
compound=$(cat $inputFolder | gawk '/^import\s\($/,/^\)$/{ sub(/import \(|\)/, "", $0); print $0;}')
imports=$(cat <(echo $simple) <(echo $compound) | gawk '/(".*")/{print $1}' | sort | uniq)
contents=$(cat $inputFolder | sed 's/^package .*$//g' | sed 's/^import ".*"$//g' | perl -p0e 's/import\s\(.*?\)//sg')

echo -e "package $pkgName\nimport (\n$imports\n)\n$contents" > $tmp

echo "Formatting sources..."
go fmt $tmp 2>&1 1>/dev/null
if [ $? != 0 ]; then
    echo "fmt failed. aborting..."
    rm $tmp
    exit 1
fi

echo "Normalising imports..."
goimports -w $tmp 2>&1 1>/dev/null
if [ $? != 0 ]; then
    echo "goimports failed. aborting..."
    rm $tmp
    exit 1
fi


if [ -f "$outputPath" ]; then
    rm "$outputPath"
fi

mv "$tmp" "$outputPath"

echo "Done"