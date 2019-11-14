package main

import (
	"fmt"
	"io"
)

func Print(a *analyser, w io.Writer) {
	for _, r := range a.Routers {
		fmt.Fprintf(w, "Router %s\n", r.Snippet)
		for _, route := range r.Routes {
			fmt.Fprintf(w, "\tRoute %s\n", route.Snippet)
		}
	}
}
