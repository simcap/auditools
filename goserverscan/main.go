package main

import (
	"flag"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	dirFlag string
)

func main() {
	flag.StringVar(&dirFlag, "dir", "./", "Dir where to parse go files")
	flag.Parse()

	dirs := []string{}
	if err := filepath.Walk(dirFlag, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	for _, dir := range dirs {
		fset := token.NewFileSet()
		packages, err := parser.ParseDir(fset, dir, filterNonTestGOFiles, parser.AllErrors)
		if err != nil {
			log.Fatal(err)
		}

		analyser := NewAnalyser(fset)
		for _, pkg := range packages {
			for _, f := range pkg.Files {
				analyser.Run(f)
			}
		}
		Print(analyser, os.Stdout)
	}
}

func filterNonTestGOFiles(info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}

	if filepath.Ext(info.Name()) == ".go" && !strings.HasSuffix(info.Name(), "_test.go") {
		return true
	}

	return false
}
