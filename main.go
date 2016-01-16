package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

func main() {
	var interfaces, constructors, filenames []string

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		if strings.HasPrefix(arg, "-") {
			switch arg[1:] {
			case "i":
				interfaces = append(interfaces, os.Args[i+1])
			case "c":
				constructors = append(constructors, os.Args[i+1])
			default:
				panic(fmt.Errorf("Unrecognized option: %s", arg))
			}

			i++
		} else {
			filenames = append(filenames, os.Args[i])
		}
	}

	fset := token.NewFileSet()
	for _, filename := range(filenames) {
		parsed, err := parser.ParseFile(fset, filename, nil, 0)
		if err != nil {
			panic(err)
		}

		for _, decl := range(parsed.Decls) {
			if gendecl, ok := decl.(*ast.GenDecl); ok && gendecl.Tok == token.TYPE {
				for _, spec := range(gendecl.Specs) {
					if _, ok := spec.(*ast.TypeSpec).Type.(*ast.InterfaceType); ok {
						printer.Fprint(os.Stdout, token.NewFileSet(), gendecl)
						// fmt.Println("type", spec.(*ast.TypeSpec).Name, "interface", "{")
						// fmt.Println("}")
					}
				}
			}
		}
	}
}