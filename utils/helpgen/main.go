package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

const header = "// THIS FILE IS GENERATED!\n// DO NOT EDIT\npackage main\n\nconst helpMessage = `"
const footer = "\n`"

var regexPackageDocs = regexp.MustCompile(`/\*\s*([^*]*?)\s*\*/\s*package`)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	src, err := ioutil.ReadFile(filepath.Join(wd, "gov.go"))
	if err != nil {
		panic(err)
	}

	caps := regexPackageDocs.FindSubmatch(src)
	if caps == nil {
		fmt.Println("No package docs found")
		return
	}

	out, err := os.OpenFile("help.go", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 666)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	io.WriteString(out, header)
	out.Write(caps[1])
	io.WriteString(out, footer)
}
