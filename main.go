package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

type Ctx struct {
	Css  string
	Js   string
	Cast string
}

func unsafeReadAll(r io.Reader) string {
	res, _ := ioutil.ReadAll(r)
	return string(res)
}

func unsafeAsset(path string) string {
	data, ok := RES.String(path)
	if !ok {
		panic("failed to open loaded resource")
	}

	return string(data)
}

func b64EncodeString(name string) string {
	var outBuf strings.Builder
	encoder := base64.NewEncoder(base64.StdEncoding, &outBuf)
	encoder.Write([]byte(name))
	encoder.Close()
	return outBuf.String()
}

func main() {
	var inputFile io.Reader
	var outputFile io.Writer
	var err error

	inputFilePath := flag.String("in", "", "Input record json file")
	outputFilePath := flag.String("out", "", "Ouput html file")

	flag.Parse()
	if len(flag.Args()) != 0 {
		flag.Usage()
		os.Exit(1)
	}

	if *inputFilePath == "" {
		inputFile = os.Stdin
	} else {
		inputFile, err = os.Open(*inputFilePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not process input file:", err)
			os.Exit(1)
		}

		defer inputFile.(*os.File).Close()
	}

	if *outputFilePath == "" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.OpenFile(*outputFilePath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not process output file:", err)
			os.Exit(1)
		}

		defer outputFile.(*os.File).Close()
	}

	tmpl, err := template.New("asciinema").Parse(unsafeAsset("/data/template.html"))
	if err != nil {
		panic(err)
	}
	ctx := Ctx{
		Css:  unsafeAsset("/data/asciinema-player.css"),
		Js:   unsafeAsset("/data/asciinema-player.js"),
		Cast: b64EncodeString(unsafeReadAll(inputFile)),
	}
	err = tmpl.Execute(outputFile, ctx)
	if err != nil {
		panic(err)
	}
}
