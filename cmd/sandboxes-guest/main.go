package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"

	"github.com/rheic/sandboxes-guest/internal/skilllink"
	"github.com/rheic/sandboxes-guest/internal/version"
)

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}

func run(ctx context.Context, args []string, out, errOut io.Writer) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintf(out, "sandboxes-guest %s\n", version.Version)
		return 0
	}
	if args[0] != "link-skills" {
		_, _ = fmt.Fprintln(errOut, "unknown command")
		return 2
	}
	flags := flag.NewFlagSet("link-skills", flag.ContinueOnError)
	flags.SetOutput(errOut)
	agent := flags.String("agent", "", "agent name")
	home := flags.String("home", "", "absolute agent home")
	store := flags.String("store", "", "absolute shared store mount")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		_, _ = fmt.Fprintln(errOut, "unexpected arguments")
		return 2
	}
	if *home == "" {
		current, err := user.Current()
		if err != nil || current.HomeDir == "" {
			_, _ = fmt.Fprintln(errOut, "link skills: cannot resolve the current user's home")
			return 1
		}
		*home = current.HomeDir
	}
	result, err := skilllink.Link(ctx, *agent, *home, *store)
	if err != nil {
		_, _ = fmt.Fprintln(errOut, "link skills:", err)
		return 1
	}
	_, _ = fmt.Fprintf(out, "linked=%d existing=%d conflicts=%d\n", result.Linked, result.Existing, result.Conflicts)
	return 0
}
