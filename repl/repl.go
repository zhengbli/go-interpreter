package repl

import (
	"bufio"
	"fmt"
	"inter/lexer"
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

		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}
}

func printErrors(out io.Writer, errs []string) {
	for _, err := range errs {
		io.WriteString(out, "\t"+err+"\n")
	}
}
