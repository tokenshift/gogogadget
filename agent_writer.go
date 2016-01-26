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
	for _, decl := range(w.Input.Decls) {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range(genDecl.Specs) {
				tspec := spec.(*ast.TypeSpec)
				if tspec.Name.Name == w.InterfaceName {
					if iface, ok := tspec.Type.(*ast.InterfaceType); ok {
						// FooAgent type and wrapped interface.
						fmt.Fprintf(out, "type %sAgent struct {\n", w.InterfaceName)
						fmt.Fprintf(out, "\twrapped %s\n", w.InterfaceName)

						for _, method := range(iface.Methods.List) {
							mType := method.Type.(*ast.FuncType)
							// TODO: Is there any case where an interface method
							// might have more than one name? The AST type supports
							// it, but I can't think of any valid construct for
							// such a method.

							// Request and response channels for the wrapped method.
							fmt.Fprintf(out, "\treq%s chan struct{%s}\n", method.Names[0], w.methodParams(mType))
							fmt.Fprintf(out, "\tres%s chan struct{%s}\n", method.Names[0], w.methodReturns(mType))

							//for _, name := range(method.Names) {
								//fmt.Fprintf(out, "\t%s(", name)

								//// Parameter types
								//for i, param := range(mType.Params.List) {
									//if i != 0 {
										//fmt.Fprint(out, ", ")
									//}

									//for j, pname := range(param.Names) {
										//if j == 0 {
											//fmt.Fprintf(out, "%s", pname)
										//} else {
											//fmt.Fprintf(out, ", %s", pname)
										//}
									//}

									//fmt.Fprint(out, " ")
									//printer.Fprint(out, fset, param.Type)
								//}

								//fmt.Fprintf(out, ") (")

								//// Return types
								//for i, param := range(mType.Results.List) {
									//if i != 0 {
										//fmt.Fprint(out, ", ")
									//}

									//for j, pname := range(param.Names) {
										//if j == 0 {
											//fmt.Fprintf(out, "%s", pname)
										//} else {
											//fmt.Fprintf(out, ", %s", pname)
										//}
									//}

									//if i != 0 {
										//fmt.Fprint(out, " ")
									//}
									
									//printer.Fprint(out, fset, param.Type)
								//}

								//fmt.Fprintf(out, ")\n")
						}

						// TODO: interface methods
						fmt.Fprintln(out, "}\n")
						return
					}
				}
			}
		}
	}
}

func (w AgentWriter) fieldList(fields *ast.FieldList) string {
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
			fmt.Fprintf(&out, "r%d ", i+1)
		}

		printer.Fprint(&out, w.fset, param.Type)
	}

	return out.String()
}

func (w AgentWriter) methodParams(mtype *ast.FuncType) string {
	return w.fieldList(mtype.Params)
}

func (w AgentWriter) methodReturns(mtype *ast.FuncType) string {
	return w.fieldList(mtype.Results)
}

// type CounterAgent struct {
// 	wrapped Counter

// 	reqAdd chan struct{int64}
// 	resAdd chan struct{int64}

// 	reqSub chan struct{int64}
// 	resSub chan struct{int64}

// 	reqTotal chan struct{}
// 	resTotal chan struct{int64}

// 	signal chan AgentSignal
// 	state AgentState
// }
