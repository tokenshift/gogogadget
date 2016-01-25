package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

func main() {
	var err error
	var interfaceName, packageName, inputFile string
	var constructors []string
	var createAgentInterface bool

	if os.Args[1] != "agent" {
		fatalError(fmt.Errorf("Unrecognized command: %s", os.Args[1]))
	}

	if len(os.Args) < 3 {
		fatalError(fmt.Errorf("Please provide an interface name."))
	}

	interfaceName = os.Args[2]

	for i := 3; i < len(os.Args); i++ {
		arg := os.Args[i]

		if strings.HasPrefix(arg, "-") {
			switch arg[1:] {
			case "i":
				if inputFile != "" {
					fatalError(fmt.Errorf("Only one input file can be specified."))
				}
				inputFile = os.Args[i+1]
				i++
			case "c":
				constructors = append(constructors, os.Args[i+1])
				i++
			case "p":
				if packageName != "" {
					fatalError(fmt.Errorf("Only one package name can be specified."))
				}
				packageName = os.Args[i+1]
				i++
			case "I":
				createAgentInterface = true
			default:
				fatalError(fmt.Errorf("Unrecognized option: %s", arg))
			}
		} else {
			fatalError(fmt.Errorf("Unexpected parameter: %s", os.Args[i]))
		}
	}

	var parsed *ast.File
	if inputFile != "" {
		parsed, err = parseFile(inputFile)
		fatalError(err)
	}
	
	writeCodeGenerationWarning(os.Stdout)

	writePackageName(os.Stdout, packageName)

	if createAgentInterface {
		writeAgentInterface(os.Stdout)
	}

	if parsed != nil {
		writeAgent(os.Stdout, interfaceName, parsed)
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