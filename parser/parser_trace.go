package parser

import (
	"fmt"
	"strings"
)

var indent int

func increIndent() {
	indent += 1
}

func decreIndent() {
	indent -= 1
}

func getIndent() string {
	return strings.Repeat("  ", indent)
}

func trace(fnName string) string {
	increIndent()
	fmt.Printf("%sBEGIN: %s\n", getIndent(), fnName)
	return fnName
}

func untrace(msg string) {
	fmt.Printf("%sEND: %s\n", getIndent(), msg)
	decreIndent()
}
