package main

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"log"
	"os"
	"testing"
)

func TestRouters(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "testdata.go", nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	analyser := NewAnalyser(fset)
	analyser.Run(f)
	routers := analyser.Routers

	if got, want := len(routers), 4; got != want {
		t.Fatalf("got %v, want %v", got, want)
	}

	for _, r := range analyser.Routers {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", " ")
		if err := enc.Encode(r); err != nil {
			t.Fatal(err)
		}
	}
}
