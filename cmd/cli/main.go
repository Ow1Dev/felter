// Package main is the entry point for the felter CLI.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Ow1Dev/felter/cmd/cli/migrate"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args, os.Getenv, os.Stdin, os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(
	ctx context.Context,
	args []string,
	getenv func(string) string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) error {
	if len(args) < 2 {
		_, _ = fmt.Fprintln(stderr, "usage: cli <command> [args...]")
		_, _ = fmt.Fprintln(stderr, "")
		_, _ = fmt.Fprintln(stderr, "commands:")
		_, _ = fmt.Fprintln(stderr, "  migrate  Apply all pending SQL migrations")
		return fmt.Errorf("no command provided")
	}

	switch args[1] {
	case "migrate":
		return migrate.Run(ctx, args[1:], getenv, stdin, stdout, stderr)
	default:
		_, _ = fmt.Fprintf(stderr, "unknown command: %s\n", args[1])
		return fmt.Errorf("unknown command: %s", args[1])
	}
}
