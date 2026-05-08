// Command hujson converts HuJSON / JSONC input (JSON with Commas and Comments)
// into standard JSON written to stdout (or to a file via -o).
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/tailscale/hujson"
)

var version = "dev"

func main() {
	cmd := newRootCmd()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "hujson: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "hujson [flags] [file ...]",
		Short: "Convert HuJSON / JSONC to standard JSON",
		Long: `Convert HuJSON / JSONC input (JSON with Commas and Comments) to standard JSON.

By default, hujson writes to stdout. Use -o/--output to write to a file
instead. With no file arguments, hujson reads from stdin. Pass "-" to mean
stdin explicitly. Pass multiple files to write each converted document in
order, separated by newlines — the result is a stream of JSON values
(consumable by tools like jq), not a single valid JSON document.`,
		Version: version,
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(args, outputPath, cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "write output to `file` instead of stdout")
	return cmd
}

func run(paths []string, outputPath string, stdin io.Reader, stdout io.Writer) (rerr error) {
	if len(paths) == 0 {
		paths = []string{"-"}
	}

	out := stdout
	if outputPath != "" {
		f, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := f.Close(); cerr != nil && rerr == nil {
				rerr = cerr
			}
		}()
		out = f
	}

	for _, p := range paths {
		if err := convert(p, stdin, out); err != nil {
			return fmt.Errorf("%s: %w", displayName(p), err)
		}
	}
	return nil
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
