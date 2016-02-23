package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"

	"github.com/alecthomas/kingpin"
)

var (
	inputFile     = kingpin.Arg("source", "Input source file.").Required().String()
	outputFile    = kingpin.Flag("output", "Where the generated code will be written.").Short('o').String()
	packageName   = kingpin.Flag("package", "The output package name.").Short('p').Required().String()
	interfaceName = kingpin.Flag("interface", "The name of the interface that will be wrapped.").Short('i').Required().String()
	constructors  = kingpin.Flag("constructor", "The name of a constructor that will be wrapped.").Short('c').Strings()
)

func main() {
	var err error

	kingpin.Parse()

	parsed, err := parseFile(*inputFile)
	fatalError(err)

	writer := NewAgentWriter(*interfaceName, *packageName, parsed)

	var out io.Writer
	if *outputFile == "" {
		out = os.Stdout
	} else {
		out, err = os.Create(*outputFile)
		fatalError(err)
	}

	WriteCodeGenerationWarning(out)
	writer.WritePackageName(out)
	WriteLibImport(out)

	writer.WriteAgentType(out)
	writer.WriteAgentMethods(out)
	writer.WriteAgentControl(out)
	writer.WriteRunLoop(out)

	for _, constructor := range *constructors {
		writer.WriteConstructor(out, constructor)
	}
}

func fatalError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseFile(filename string) (*ast.File, error) {
	var parsed *ast.File

	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}
