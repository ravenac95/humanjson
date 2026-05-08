// Command hujson converts HuJSON / JSONC input (JSON with Commas and Comments)
// into standard JSON written to stdout.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tailscale/hujson"
)

var version = "dev"

const usageText = `Usage: hujson [file ...]

Convert HuJSON / JSONC input (JSON with Commas and Comments) to standard JSON
on stdout.

With no file arguments, hujson reads from stdin. Pass "-" to mean stdin
explicitly. Pass multiple files to write each converted document to stdout in
order.

Options:
  -version    Print the version and exit.
  -h, -help   Show this message.
`

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() { fmt.Fprint(os.Stderr, usageText) }
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	paths := flag.Args()
	if len(paths) == 0 {
		paths = []string{"-"}
	}

	for _, p := range paths {
		if err := convert(p, os.Stdin, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "hujson: %s: %v\n", displayName(p), err)
			os.Exit(1)
		}
	}
}

func displayName(path string) string {
	if path == "-" {
		return "<stdin>"
	}
	return path
}

func convert(path string, stdin io.Reader, stdout io.Writer) error {
	var (
		data []byte
		err  error
	)
	if path == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return err
	}
	v, err := hujson.Parse(data)
	if err != nil {
		return err
	}
	v.Standardize()
	out := v.Pack()
	if _, err := stdout.Write(out); err != nil {
		return err
	}
	if len(out) == 0 || out[len(out)-1] != '\n' {
		if _, err := stdout.Write([]byte{'\n'}); err != nil {
			return err
		}
	}
	return nil
}
