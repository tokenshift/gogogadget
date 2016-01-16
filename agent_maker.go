package main

import (
	"go/ast"
	"go/printer"
	"go/token"
	"os"
)

func createAgent(interfaceName string, files []*ast.File) (*ast.File, error) {
	for _, file := range(files) {
		for _, decl := range(file.Decls) {
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

	return nil, nil
}