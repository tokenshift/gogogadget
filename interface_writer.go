package main

import (
	"fmt"
	"go/ast"
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
}