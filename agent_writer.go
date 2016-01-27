package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
)

type AgentWriter struct {
	InterfaceName, PackageName string
	Input *ast.File
	fset *token.FileSet
}

func NewAgentWriter(interfaceName, packageName string, parsed *ast.File) AgentWriter {
	return AgentWriter{
		interfaceName,
		packageName,
		parsed,
		token.NewFileSet(),
	}
}

func WriteCodeGenerationWarning(out io.Writer) {
	fmt.Fprintln(out, "// THIS CODE WAS GENERATED USING github.com/tokenshift/gogogadget")
	fmt.Fprintln(out, "// ANY CHANGES TO THIS FILE MAY BE OVERWRITTEN")
	fmt.Fprintln(out, "")
}

func WriteAgentInterface(out io.Writer) {
	fmt.Fprintln(out, `type Agent interface {
	Start()
	Stop()
	Close()
	State() AgentState
}

type AgentSignal byte
const (
	AGENT_START AgentSignal = iota
	AGENT_STOP
	AGENT_CLOSE
)

type AgentState byte
const (
	AGENT_STARTED AgentState = iota
	AGENT_STOPPED
	AGENT_CLOSED
)`)
	fmt.Fprintln(out, "")
}

func (w AgentWriter) WritePackageName(out io.Writer) {
	fmt.Fprintf(out, "package %s\n\n", w.PackageName)
}

func (w AgentWriter) WriteAgent(out io.Writer) {
	// FooAgent type and wrapped interface.
	fmt.Fprintf(out, "type %sAgent struct {\n", w.InterfaceName)
	fmt.Fprintf(out, "\twrapped %s\n", w.InterfaceName)

	// Request and response channels for each wrapped method.
	for method := range(w.interfaceMethods()) {
		fmt.Fprintf(out, "\treq%s chan struct{%s}\n", method.Names[0], w.methodParams(method.Type.(*ast.FuncType)))
		fmt.Fprintf(out, "\tres%s chan struct{%s}\n", method.Names[0], w.methodReturns(method.Type.(*ast.FuncType)))
	}

	// Close the FooAgent type definition.
	fmt.Fprintln(out, "}\n")

	// Create method wrappers.
	for method := range w.interfaceMethods() {
		fmt.Fprintf(out,
			"func (agent %sAgent) %s(%s) (%s) {\n",
			w.InterfaceName,
			method.Names[0],
			w.methodParams(method.Type.(*ast.FuncType)),
			w.methodReturns(method.Type.(*ast.FuncType)))

		// Send the method arguments as a message to the req channel.
		fmt.Fprintf(out,
			"\tagent.req%s <- struct{%s}{\n",
			method.Names[0],
			w.methodParams(method.Type.(*ast.FuncType)))
		for i, param := range method.Type.(*ast.FuncType).Params.List {
			for _, pname := range param.Names {
				fmt.Fprintf(out, "\t\t%s: %s,\n", pname, pname)
			}

			if len(param.Names) == 0 {
				fmt.Fprintf(out, "\t\targ%d: arg%d,\n", i+1, i+1)
			}
		}
		fmt.Fprintln(out, "\t}\n")

		// Receive the return value(s) on the res channel.
		fmt.Fprintf(out, "\tres := <- agent.res%s\n", method.Names[0])
		fmt.Fprint(out, "\treturn ")
		for i, param := range method.Type.(*ast.FuncType).Results.List {
			if i > 0 {
				fmt.Fprint(out, ", ")
			}

			for _, pname := range param.Names {
				fmt.Fprintf(out, "res.%s", pname)
			}

			if len(param.Names) == 0 {
				fmt.Fprintf(out, "rest.rval%d", i+1)
			}
		}
		fmt.Fprintln(out, "")

		fmt.Fprintln(out, "}\n")
	}
}

//func (c CounterAgent) Add(val int64) int64 {
	//c.reqAdd <- struct{
		//val int64
		//a int64
		//b int64
	//}{val, val, val}
	//res := <- c.resAdd
	//return res.int64
//}

func (w AgentWriter) fieldList(fields *ast.FieldList, argPrefix string) string {
	var out bytes.Buffer

	for i, param := range(fields.List) {
		if i != 0 {
			fmt.Fprint(&out, ", ")
		}

		for j, pname := range(param.Names) {
			if j != 0 {
				fmt.Fprint(&out, ", ")
			}

			fmt.Fprint(&out, pname)
			fmt.Fprint(&out, " ")
		}

		if len(param.Names) == 0 {
			// Give the arg an arbitrary name, since anonymous structs can't
			// otherwise have multiple fields of the same type.
			fmt.Fprintf(&out, "%s%d ", argPrefix, i+1)
		}

		printer.Fprint(&out, w.fset, param.Type)
	}

	return out.String()
}

func (w AgentWriter) interfaceMethods() <-chan *ast.Field {
	out := make(chan *ast.Field)

	go func() {
		for _, decl := range(w.Input.Decls) {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
				for _, spec := range(genDecl.Specs) {
					tspec := spec.(*ast.TypeSpec)
					if tspec.Name.Name == w.InterfaceName {
						if iface, ok := tspec.Type.(*ast.InterfaceType); ok {
							for _, method := range(iface.Methods.List) {
								// TODO: Is there any case where an interface method
								// might have more than one name? The AST type supports
								// it, but I can't think of any valid construct for
								// such a method.
								out <- method
							}
						}
					}
				}
			}
		}

		close(out)
	}()

	return out
}

func (w AgentWriter) methodParams(mtype *ast.FuncType) string {
	return w.fieldList(mtype.Params, "arg")
}

func (w AgentWriter) methodReturns(mtype *ast.FuncType) string {
	return w.fieldList(mtype.Results, "rval")
}
