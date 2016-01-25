package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
)

func writeCodeGenerationWarning(out io.Writer) {
	fmt.Fprintln(out, "// THIS CODE WAS GENERATED USING github.com/tokenshift/gogogadget")
	fmt.Fprintln(out, "// ANY CHANGES TO THIS FILE MAY BE OVERWRITTEN")
	fmt.Fprintln(out, "")
}

func writePackageName(out io.Writer, packageName string) {
	fmt.Fprintf(out, "package %s\n\n", packageName)
}

func writeAgentInterface(out io.Writer) {
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

func writeAgent(out io.Writer, interfaceName string, parsed *ast.File) {
	for _, decl := range(parsed.Decls) {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range(genDecl.Specs) {
				tspec := spec.(*ast.TypeSpec)
				if tspec.Name.Name == interfaceName {
					if iface, ok := tspec.Type.(*ast.InterfaceType); ok {
						fmt.Fprintf(out, "type %sAgent struct {\n", interfaceName)
						fmt.Fprintf(out, "\twrapped %s\n", interfaceName)

						for _, method := range(iface.Methods.List) {
							for _, name := range(method.Names) {
								fmt.Fprintf(out, "\t%s() (returns)\n", name)
							}
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