package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"github.com/alecthomas/kingpin"
)

var (
	inputFile     = kingpin.Arg("source", "Input source files.").Required().String()
	packageName   = kingpin.Flag("package", "The output package name.").Short('p').Required().String()
	interfaceName = kingpin.Flag("interface", "The name of the interface that will be wrapped.").Short('i').Required().String()
	constructors  = kingpin.Flag("constructor", "The name of a constructor that will be wrapped.").Short('c').Strings()
)

func main() {
	kingpin.Parse()

	parsed, err := parseFile(*inputFile)
	fatalError(err)

	writer := NewAgentWriter(*interfaceName, *packageName, parsed)

	WriteCodeGenerationWarning(os.Stdout)
	writer.WritePackageName(os.Stdout)
	WriteLibImport(os.Stdout)

	writer.WriteAgentType(os.Stdout)
	writer.WriteAgentMethods(os.Stdout)
	writer.WriteAgentControl(os.Stdout)
	//writer.WriteRunLoop(os.Stdout)

	/*for _, constructor := range *constructors {
		writer.WriteConstructor(os.Stdout, constructor)
	}*/
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
