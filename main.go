package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/ulikunitz/xz/lzma"
)

type Ctx struct {
	Css      string
	PlayerJs string
	LzmaJs   string
	Cast     string
	Font     string
}

func unsafeReadAll(r io.Reader) []byte {
	res, _ := ioutil.ReadAll(r)
	return res
}

func unsafeAsset(path string) string {
	data, err := fs.ReadFile(assets, path)
	if err != nil {
		panic(fmt.Sprintf("failed to open loaded resource, %s", err))
	}

	return string(data)
}

func b64Encode(data []byte) []byte {
	var outBuf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &outBuf)
	encoder.Write(data)
	encoder.Close()
	return outBuf.Bytes()
}

func lzmaEncode(data []byte) []byte {
	var outBuf bytes.Buffer
	writer, _ := lzma.NewWriter(&outBuf)
	writer.Write(data)
	writer.Close()
	return outBuf.Bytes()
}

func main() {
	var inputFile io.Reader
	var outputFile io.Writer
	var err error

	inputFilePath := flag.String("in", "", "Input record json file")
	outputFilePath := flag.String("out", "", "Ouput html file")
	fontFilePath := flag.String(
		"font",
		"",
		"Font to embed (currently only TTF supported)",
	)

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
			fmt.Fprintln(os.Stderr, "Could not open input file:", err)
			os.Exit(1)
		}

		defer inputFile.(*os.File).Close()
	}

	inputfile, err := io.ReadAll(inputFile)
	if err != nil {
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not read input file:", err)
			os.Exit(1)
		}
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

	tmpl, err := template.New("asciinema").Parse(unsafeAsset("data/template.html"))
	if err != nil {
		panic(err)
	}
	ctx := Ctx{
		Css:      unsafeAsset("data/asciinema-player.css"),
		PlayerJs: string(b64Encode(lzmaEncode([]byte(unsafeAsset("data/asciinema-player.js"))))),
		LzmaJs:   unsafeAsset("data/lzma-d-min.js"),
		Cast:     string(b64Encode(lzmaEncode(inputfile))),
	}
	if *fontFilePath != "" {
		fontFile, err := ioutil.ReadFile(*fontFilePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not read font file:", err)
			os.Exit(1)
		}
		ctx.Font = string(b64Encode(fontFile))
	}
	err = tmpl.Execute(outputFile, ctx)
	if err != nil {
		panic(err)
	}
}
