package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"

	m "github.com/hoisie/mustache"
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

// Code generation warning added to the top of every output file.
const CODE_GENERATION_WARNING = `// THIS CODE WAS GENERATED USING github.com/tokenshift/gogogadget
// ANY CHANGES TO THIS FILE MAY BE OVERWRITTEN`

// Agent type, containing message and signal channels and the wrapped implementation.
var tmplAgentType, _ = m.ParseString(`type {{InterfaceName}}Agent struct {
	wrapped {{InterfaceName}}
	signal chan AgentSignal
	state AgentStat{{#Methods}}


	req{{Name}} chan {{RequestType}}
	res{{Name}} chan {{ResponseType}}{{/Methods}}
}`)

type tmplAgentTypeParams struct{
	InterfaceName string
	Methods []struct{Name, RequestType, ResponseType string}
}

func WriteCodeGenerationWarning(out io.Writer) {
	fmt.Fprintln(out, CODE_GENERATION_WARNING, "\n")
}

func (w AgentWriter) WritePackageName(out io.Writer) {
	fmt.Fprintf(out, "package %s\n\n", w.PackageName)
}

func (w AgentWriter) WriteAgentType(out io.Writer) {
	params := tmplAgentTypeParams{
		InterfaceName: w.InterfaceName,
	}

	for method := range(w.interfaceMethods()) {
		params.Methods = append(params.Methods, struct{Name, RequestType, ResponseType string}{
			method.Names[0].Name,
			w.methodParams(method.Type.(*ast.FuncType)),
			w.methodReturns(method.Type.(*ast.FuncType)),
		})
	}

	fmt.Fprintln(out, tmplAgentType.Render(params), "\n")

	//// FooAgent type and wrapped interface.
	//fmt.Fprintf(out, "type %sAgent struct {\n", w.InterfaceName)
	//fmt.Fprintf(out, "\twrapped %s\n", w.InterfaceName)

	//// Request and response channels for each wrapped method.
	//for method := range(w.interfaceMethods()) {
		//fmt.Fprintf(out, "\treq%s chan struct{%s}\n", method.Names[0], w.methodParams(method.Type.(*ast.FuncType)))
		//fmt.Fprintf(out, "\tres%s chan struct{%s}\n", method.Names[0], w.methodReturns(method.Type.(*ast.FuncType)))
	//}

	//// Agent signal channel and run state.
	//fmt.Fprintln(out, "\tsignal chan AgentSignal")
	//fmt.Fprintln(out, "\tstate AgentState")

	//// Close the FooAgent type definition.
	//fmt.Fprintln(out, "}\n")
}

func (w AgentWriter) WriteConstructor(out io.Writer, cname string) {
	constructor := w.findConstructor(cname)
	fmt.Fprintf(out,
		"func %sAgent(%s) %sAgent {\n",
		cname,
		w.methodParams(constructor),
		w.InterfaceName)

	// Call the wrapped constructor.
	fmt.Fprintf(out, "\twrapped := %s(", cname)
	for i, param := range constructor.Params.List {
		if i > 0 {
			fmt.Fprint(out, ", ")
		}

		for _, pname := range param.Names {
			fmt.Fprint(out, pname)
		}

		if len(param.Names) == 0 {
			fmt.Fprintf(out, "arg%d", i+1, i+1)
		}
	}
	fmt.Fprint(out, ")\n")

	// Initialize the FooAgent with the wrapped object and req/res channels.
	fmt.Fprintf(out, "\tagent := %sAgent{\n", w.InterfaceName)
	fmt.Fprintln(out, "\t\twrapped,")
	for method := range(w.interfaceMethods()) {
		fmt.Fprintf(out,
			"\t\tmake(chan struct{%s}),\n",
			w.methodParams(method.Type.(*ast.FuncType)))
		fmt.Fprintf(out,
			"\t\tmake(chan struct{%s}),\n",
			w.methodReturns(method.Type.(*ast.FuncType)))
	}
	fmt.Fprintln(out, "\t\tmake(chan AgentSignal),")
	fmt.Fprintln(out, "\t\tAGENT_STARTED,")
	fmt.Fprintln(out, "\t}\n")

	// Start the FooAgent runLoop.
	fmt.Fprintln(out, "\tgo agent.runLoop()\n")

	// Return the FooAgent.
	fmt.Fprintln(out, "\treturn agent")

	fmt.Fprintln(out, "}\n")
}

func (w AgentWriter) WriteAgentControl(out io.Writer) {
	// Generate agent run loop.
	fmt.Fprintf(out, "func (agent *%sAgent) runLoop() {", w.InterfaceName)
	fmt.Fprintln(out, `
	for {
		select {
		case signal := <-agent.signal:
			switch signal {
			case AGENT_START:
				agent.state = AGENT_STARTED
			case AGENT_STOP:
				agent.state = AGENT_STOPPED
			case AGENT_CLOSE:
				agent.state = AGENT_CLOSED
				agent.close()
				return
			}`)

	for method := range(w.interfaceMethods()) {
		fmt.Fprintf(out, "\t\tcase msg := <-agent.req%s:\n", method.Names[0])
		for i, param := range method.Type.(*ast.FuncType).Results.List {
			if i > 0 {
				fmt.Fprint(out, ", ")
			}

			for _, pname := range param.Names {
				fmt.Fprint(out, pname)
			}

			if len(param.Names) == 0 {
				fmt.Fprintf(out, "\t\t\trval%d", i+1)
			}
		}
		fmt.Fprintf(out, " := agent.wrapped.%s(", method.Names[0])
		for i, param := range method.Type.(*ast.FuncType).Params.List {
			if i > 0 {
				fmt.Fprint(out, ", ")
			}

			for _, pname := range param.Names {
				fmt.Fprintf(out, "msg.%s", pname)
			}

			if len(param.Names) == 0 {
				fmt.Fprintf(out, "\t\t\tmsg.arg%d", i+1)
			}
		}
		fmt.Fprintln(out, ")")

		fmt.Fprintf(out,
			"\t\t\tagent.res%s<- struct{%s}{",
			method.Names[0],
			w.methodReturns(method.Type.(*ast.FuncType)))
		for i, param := range method.Type.(*ast.FuncType).Results.List {
			if i > 0 {
				fmt.Fprint(out, ", ")
			}

			for _, pname := range param.Names {
				fmt.Fprint(out, pname)
			}

			if len(param.Names) == 0 {
				fmt.Fprintf(out, "rval%d", i+1)
			}
		}

		fmt.Fprint(out, "}\n")
	}

	fmt.Fprintln(out, "\t\t}\n\t}\n}\n")
}

//func (c CounterAgent) runLoop() {
	//for {
		//select {
		//case signal := <-c.signal:
			//switch signal {
			//case AGENT_START:
				//c.state = AGENT_STARTED
			//case AGENT_STOP:
				//c.state = AGENT_STOPPED
			//case AGENT_CLOSE:
				//c.state = AGENT_CLOSED
				//c.close()
				//return
			//}
		//case msg := <-c.reqAdd:
			//c.resAdd<- struct{int64}{c.wrapped.Add(msg.int64)}
		//case msg := <-c.reqSub:
			//c.resSub<- struct{int64}{c.wrapped.Sub(msg.int64)}
		//case _ = <-c.reqTotal:
			//c.resTotal<- struct{int64}{c.wrapped.Total()}
		//}
	//}
//}

//func (c CounterAgent) close() {
	//close(c.reqAdd)
	//close(c.resAdd)
	//close(c.reqSub)
	//close(c.resSub)
	//close(c.reqTotal)
	//close(c.resTotal)
	//close(c.signal)
//}

//func (c CounterAgent) Start() {
	//c.signal <- AGENT_START
//}

//func (c CounterAgent) Stop() {
	//c.signal <- AGENT_STOP
//}

//func (c CounterAgent) Close() {
	//c.signal <- AGENT_CLOSE
//}

//func (c CounterAgent) State() AgentState {
	//return c.state
//}

func (w AgentWriter) WriteAgentMethods(out io.Writer) {
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

func (w AgentWriter) findConstructor(cname string) *ast.FuncType {
	for _, decl := range(w.Input.Decls) {
		if fdecl, ok := decl.(*ast.FuncDecl); ok && fdecl.Name.Name == cname {
			return fdecl.Type
		}
	}

	panic(fmt.Errorf("Failed to locate constructor %s", cname))
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
	return w.fieldList(mtype.Results, "r")
}

func (w AgentWriter) requestType(mtype *ast.FuncType) string {
	return fmt.Sprint("struct{", w.methodParams(mtype), "}")
}

func (w AgentWriter) responseType(mtype *ast.FuncType) string {
	return fmt.Sprint("struct{", w.methodReturns(mtype), "}")
}
