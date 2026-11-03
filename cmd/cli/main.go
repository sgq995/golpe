// Command golpe is the Golpe CLI. It turns .golpe files into plain Go.
//
// Build it with: go build -o golpe ./cmd/cli
//
// See docs/knowledge/specs/003-cli.md.
package main

import (
	"fmt"
	"io"
	"os"
)

const version = "0.0.0-dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, "golpe", version)
		return 0
	case "help", "-h", "--help":
		usage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "golpe: unknown command %q\n\n", args[0])
		usage(stderr)
		return 2
	}
}

func usage(w io.Writer) {
	fmt.Fprint(w, `Usage: golpe <command> [arguments]

Commands:
  version   print the golpe version
  help      show this help

Planned (see docs/knowledge/specs/003-cli.md):
  generate  compile .golpe files into _golpe.go files
  fmt       format .golpe files
  lsp       run the language server
`)
}
