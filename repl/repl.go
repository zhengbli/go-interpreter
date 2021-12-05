package repl

import (
	"bufio"
	"fmt"
	"inter/evaluator"
	"inter/lexer"
	"inter/object"
	"inter/parser"
	"io"
)

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintf(out, ">> ")
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()
		errs := p.Errors()
		if len(errs) > 0 {
			printErrors(out, errs)
			continue
		}

		env := object.NewEnvironment()
		evaled := evaluator.Eval(program, env)
		if evaled != nil {
			io.WriteString(out, "Result: "+evaled.Inspect()+"\n")
		}
	}
}

func printErrors(out io.Writer, errs []string) {
	for _, err := range errs {
		io.WriteString(out, "\t"+err+"\n")
	}
}
