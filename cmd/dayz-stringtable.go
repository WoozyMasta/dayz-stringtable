// Package main implements the CLI entrypoint for DayZ CSV localization helper.
package main

import (
	"fmt"
	"os"

	"github.com/woozymasta/dayz-stringtable/internal/commands"
	"github.com/woozymasta/dayz-stringtable/internal/vars"

	"github.com/jessevdk/go-flags"
)

// Options provide base cli flags
type Options struct {
	Version bool `short:"v" long:"version"` // Show version and build info
}

// main sets up the parser and dispatches commands.
func main() {
	for _, arg := range os.Args[1:] {
		if arg == "-v" || arg == "--version" {
			printVersion()
		}
	}

	opts := &Options{}
	parser := flags.NewParser(opts, flags.Default)
	parser.Name = "dayz-stringtable"
	parser.ShortDescription = "DayZ CSV localization helper"

	// register subcommands
	for _, c := range []struct {
		cmd                  any
		name, desc, longDesc string
	}{
		{
			&commands.PotCmd{},
			"pot",
			"Generate POT template from CSV",
			"Read .csv and emit .pot template",
		},
		{
			&commands.PosCmd{},
			"pos",
			"Generate PO files from CSV per language",
			"Read .csv and output .po per lang",
		},
		{
			&commands.MakeCmd{},
			"make",
			"Make new CSV from original CSV and PO files",
			"Read .csv + .po files and write merged CSV",
		},
		{
			&commands.UpdateCmd{},
			"update",
			"Update existing PO with new records",
			"Read .csv and rewrite .po files in-place or to out-dir",
		},
	} {
		mustAdd(parser, c.name, c.desc, c.longDesc, c.cmd)
	}

	_, err := parser.Parse()
	if err != nil {
		if opts.Version {
			printVersion()
		}

		if ferr, ok := err.(*flags.Error); ok && ferr.Type == flags.ErrHelp {
			os.Exit(0)
		}

		os.Exit(1)
	}
}

// mustAdd registers a subcommand or exits on error.
func mustAdd(parser *flags.Parser, name, desc, longDesc string, cmd any) {
	if _, err := parser.AddCommand(name, desc, longDesc, cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding %s command: %v", name, err)
		os.Exit(1)
	}
}

func printVersion() {
	fmt.Printf(
		"file:     %s\nversion:  %s\ncommit:   %s\nbuilt:    %s\nproject:  %s\n",
		os.Args[0], vars.Version, vars.Commit, vars.BuildTime, vars.URL)
	os.Exit(0)
}
