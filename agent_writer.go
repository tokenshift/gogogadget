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

func WriteCodeGenerationWarning(out io.Writer) {
	fmt.Fprintln(out, CODE_GENERATION_WARNING)
	fmt.Fprint(out, "\n")
}

func (w AgentWriter) WritePackageName(out io.Writer) {
	fmt.Fprintf(out, "package %s\n\n", w.PackageName)
}

func WriteLibImport(out io.Writer) {
	fmt.Fprintln(out, "import . \"github.com/tokenshift/gogogadget/lib\"\n")
}

// Agent type, containing message and signal channels and the wrapped implementation.
var tmplAgentType = template(`type {{InterfaceName}}Agent struct {
	wrapped {{InterfaceName}}
	signal chan AgentSignal
	state AgentState{{#Methods}}


	req{{Name}} chan {{RequestType}}
	res{{Name}} chan {{ResponseType}}{{/Methods}}
}`)

type tmplAgentTypeParams struct{
	InterfaceName string
	Methods []struct{Name, RequestType, ResponseType string}
}

func (w AgentWriter) WriteAgentType(out io.Writer) {
	params := tmplAgentTypeParams{
		InterfaceName: w.InterfaceName,
	}

	for method := range(w.interfaceMethods()) {
		params.Methods = append(params.Methods, struct{Name, RequestType, ResponseType string}{
			method.Names[0].Name,
			w.requestType(method.Type.(*ast.FuncType)),
			w.responseType(method.Type.(*ast.FuncType)),
		})
	}

	fmt.Fprintln(out, tmplAgentType.Render(params))
	fmt.Fprint(out, "\n")
}

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
			"\tagent.req%s<- struct{%s}{\n",
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
		fmt.Fprintln(out, "\t}")

		// Receive the return value(s) on the res channel.
		fmt.Fprintf(out, "\tres := <-agent.res%s\n", method.Names[0])
		fmt.Fprint(out, "\treturn ")
		for i, param := range method.Type.(*ast.FuncType).Results.List {
			if i > 0 {
				fmt.Fprint(out, ", ")
			}

			for _, pname := range param.Names {
				fmt.Fprintf(out, "res.%s", pname)
			}

			if len(param.Names) == 0 {
				fmt.Fprintf(out, "res.val%d", i+1)
			}
		}
		fmt.Fprintln(out, "")

		fmt.Fprintln(out, "}\n")
	}
}

// Agent control functions.
var tmplAgentControl = template(`func (agent {{InterfaceName}}Agent) Start() {
	agent.signal<- AGENT_START
}

func (agent {{InterfaceName}}Agent) Stop() {
	agent.signal<- AGENT_STOP
}

func (agent {{InterfaceName}}Agent) Close() {
	agent.signal<- AGENT_CLOSE
}

func (agent {{InterfaceName}}Agent) State() AgentState {
	return agent.state
}`)

func (w AgentWriter) WriteAgentControl(out io.Writer) {
	fmt.Fprintln(out, tmplAgentControl.Render(w))
	fmt.Fprint(out, "\n")
}

// TODO: Find or write a Mustache template library that supports escaping
// curly braces.
var tmplRunLoop = template(`func (agent *{{InterfaceName}}Agent) runLoop() {
	for {
		select {
		case signal := <-agent.signal:
			switch signal {
			case AGENT_START:
				agent.state = AGENT_STARTED
			case AGENT_STOP:
				agent.state = AGENT_STOPPED
			case AGENT_CLOSE:
				agent.state = AGENT_CLOSED{{#Methods}}

				close(agent.req{{MethodName}})
				close(agent.res{{MethodName}}){{/Methods}}
				close(agent.signal)
				return
			}{{#Methods}}

		case msg := <-agent.req{{MethodName}}:
			{{ResponseArgs}} := agent.wrapped.{{MethodName}}({{RequestArgs}})
			agent.res{{MethodName}}<- {{ResponseType}}{ {{ResponseArgs}} }{{/Methods}}
		}
	}
}`)

type tmplRunLoopParams struct {
	InterfaceName string
	Methods []struct {
		MethodName, RequestArgs, ResponseArgs, ResponseType string
	}
}


func (w AgentWriter) WriteRunLoop(out io.Writer) {
	params := tmplRunLoopParams{
		InterfaceName: w.InterfaceName,
	}

	for method := range(w.interfaceMethods()) {
		var requestArgs bytes.Buffer
		for i, param := range method.Type.(*ast.FuncType).Params.List {
			if i > 0 {
				fmt.Fprint(&requestArgs, ", ")
			}

			for j, pname := range param.Names {
				if j > 0 {
					fmt.Fprint(&requestArgs, ", ")
				}

				fmt.Fprintf(&requestArgs, "msg.%s", pname)
			}

			if len(param.Names) == 0 {
				fmt.Fprintf(&requestArgs, "msg.val%d", i+1)
			}
		}

		var responseArgs bytes.Buffer
		for i, param := range method.Type.(*ast.FuncType).Results.List {
			if i > 0 {
				fmt.Fprint(&responseArgs, ", ")
			}

			for j, pname := range param.Names {
				if j > 0 {
					fmt.Fprint(&responseArgs, ", ")
				}

				fmt.Fprint(&responseArgs, pname)
			}

			if len(param.Names) == 0 {
				fmt.Fprintf(&responseArgs, "val%d", i+1)
			}
		}

		mParams := struct{MethodName, RequestArgs, ResponseArgs, ResponseType string}{
			MethodName: method.Names[0].Name,
			RequestArgs: requestArgs.String(),
			ResponseArgs: responseArgs.String(),
			ResponseType: w.responseType(method.Type.(*ast.FuncType)),
		}

		params.Methods = append(params.Methods, mParams)
	}

	fmt.Fprintln(out, tmplRunLoop.Render(params))
	fmt.Fprint(out, "\n")
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
	fmt.Fprintln(out, "\t\tmake(chan AgentSignal),")
	fmt.Fprintln(out, "\t\tAGENT_STARTED,")
	for method := range(w.interfaceMethods()) {
		fmt.Fprintf(out,
			"\t\tmake(chan struct{%s}),\n",
			w.methodParams(method.Type.(*ast.FuncType)))
		fmt.Fprintf(out,
			"\t\tmake(chan struct{%s}),\n",
			w.methodReturns(method.Type.(*ast.FuncType)))
	}
	fmt.Fprintln(out, "\t}\n")

	// Start the FooAgent runLoop.
	fmt.Fprintln(out, "\tgo agent.runLoop()\n")

	// Return the FooAgent.
	fmt.Fprintln(out, "\treturn agent")

	fmt.Fprintln(out, "}\n")
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
		}

		if len(param.Names) == 0 {
			// Give the arg an arbitrary name, since anonymous structs can't
			// otherwise have multiple fields of the same type.
			fmt.Fprintf(&out, "val%d", i+1)
		}

		fmt.Fprint(&out, " ")
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
	return w.fieldList(mtype.Params)
}

func (w AgentWriter) methodReturns(mtype *ast.FuncType) string {
	return w.fieldList(mtype.Results)
}

func (w AgentWriter) requestType(mtype *ast.FuncType) string {
	return fmt.Sprint("struct{", w.methodParams(mtype), "}")
}

func (w AgentWriter) responseType(mtype *ast.FuncType) string {
	return fmt.Sprint("struct{", w.methodReturns(mtype), "}")
}

func template(source string) *m.Template {
	t, err := m.ParseString(source)
	if err != nil {
		panic(err)
	}

	return t
}
