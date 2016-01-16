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
	var interfaceName string
	var constructors, filenames []string
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
				filenames = append(filenames, os.Args[i+1])
				i++
			case "c":
				constructors = append(constructors, os.Args[i+1])
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

	parsed, err := parseFiles(filenames)
	fatalError(err)

	_, err = createAgent(interfaceName, parsed)
	fatalError(err)

	if createAgentInterface {
		fmt.Println("TODO: Create the agent base interface.")
	}
}

func fatalError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseFiles(filenames []string) ([]*ast.File, error) {
	var parsed []*ast.File

	fset := token.NewFileSet()
	for _, filename := range(filenames) {
		p, err := parser.ParseFile(fset, filename, nil, 0)
		if err != nil {
			return nil, err
		}

		parsed = append(parsed, p)
	}

	return parsed, nil
}