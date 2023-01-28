// Copyright 2023 The firestore Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/spf13/cobra"
)

const appName = "firestore"

var ioStreams = IOStreams{
	In:     os.Stdin,
	Out:    os.Stdout,
	ErrOut: os.Stderr,
}

// aos represents a root command options.
type cli struct {
	IOStreams

	fs      *firestore.Client
	project string
	color   bool
	debug   bool
}

// NewCommand creates the aos root command.
func NewCommand(ctx context.Context, args []string) (*cobra.Command, error) {
	cli := &cli{
		IOStreams: ioStreams,
	}

	cmd := &cobra.Command{
		Use:     appName,
		Short:   "Google Cloud Firestore command-line tool.",
		Long:    "Google Cloud Firestore command-line tool.",
		Version: "v0.0.1",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return cli.NewClient(cmd.Context())
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return cli.CloseClient()
		},
		SilenceErrors: true,
	}

	cli.SetFlags(cmd)
	cli.SetCommands(cmd)

	return cmd, nil
}

// SetFlags sets global flags.
func (c *cli) SetFlags(cmd *cobra.Command) {
	f := cmd.PersistentFlags()
	f.BoolVar(&c.debug, "debug", false, "Use debug output")
	f.StringVar(&c.project, "project", "", "Google Cloud Firestore project ID")
	f.BoolVar(&c.color, "color", true, "Colorize output")
}

// SetCommands sets each subcommand.
func (c *cli) SetCommands(cmd *cobra.Command) {
	coll := &Collection{cli: c}
	coll.Register(cmd)

	colls := &Collections{cli: c}
	colls.Register(cmd)

	doc := &Doc{cli: c}
	doc.Register(cmd)

	docs := &Docs{cli: c}
	docs.Register(cmd)

	// hack of PreRunE for initializes and closes firestore client
	for _, sc := range cmd.Commands() {
		sc.PreRunE = cmd.PreRunE
		sc.PostRunE = cmd.PostRunE
	}
}

type checkType uint8

const (
	exactArgs checkType = iota + 1
	minArgs
	maxArgs
)

func checkArgs(cmdName string, typ checkType, expected int, args ...string) error {
	switch typ {
	case exactArgs:
		if len(args) != expected {
			return fmt.Errorf("%s: %q requires exactly %d argument(s), args: <%s>\n", appName, cmdName, expected, strings.Join(args, " "))
		}

	case minArgs:
		if len(args) < expected {
			return fmt.Errorf("%s: %q requires a minimum of %d argument(s), args: <%s>\n", appName, cmdName, expected, strings.Join(args, " "))
		}

	case maxArgs:
		if len(args) > expected {
			return fmt.Errorf("%s: %q requires a maximum of %d argument(s), args: <%s>\n", appName, cmdName, expected, strings.Join(args, " "))
		}

	default:
		panic(fmt.Errorf("unknown checkType(%d) check type", typ))
	}

	return nil
}
