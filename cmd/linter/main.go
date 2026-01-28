package main

import (
	"github.com/MarkelovSergey/url-shorter/cmd/linter/exitcheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(exitcheck.Analyzer)
}
