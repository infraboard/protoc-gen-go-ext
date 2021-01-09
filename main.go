package main

import (
	"flag"
	"io/ioutil"
	"os"

	gengo "google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/infraboard/protoc-gen-go-ext/ast"
)

func main() {
	var (
		flags flag.FlagSet
	)
	// For Debug Only
	{ // Dump
		in, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile("in.pb", in, 0666); err != nil {
			panic(err)
		}
	}
	{ // Debug
		os.Stdin, _ = os.Open("in.pb")
		os.Stdout, _ = os.Create("out.pb")
	}

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = gengo.SupportedFeatures
		var originFiles []*protogen.GeneratedFile
		for _, f := range gen.Files {
			if f.Generate {
				originFiles = append(originFiles, gengo.GenerateFile(gen, f))
			}
		}
		ast.Rewrite(gen)

		for _, f := range originFiles {
			f.Skip()
		}
		return nil
	})
}
