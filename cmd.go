// Copyright 2023 The firestore Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/spf13/cobra"
)

const appName = "firestore"

// IOStreams represents a stdio streams.
//
// This is useful for embedding and for unit testing.
type IOStreams struct {
	// In think, os.Stdin
	In io.Reader

	// Out think, os.Stdout
	Out io.Writer

	// ErrOut think, os.Stderr
	ErrOut io.Writer
}

// NewTestIOStreams returns a valid IOStreams and in, out, errout buffers for unit tests
func NewTestIOStreams() (IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	return IOStreams{
		In:     in,
		Out:    out,
		ErrOut: errOut,
	}, in, out, errOut
}

// NewTestIOStreamsDiscard returns a valid IOStreams that just discards
func NewTestIOStreamsDiscard() IOStreams {
	in := &bytes.Buffer{}
	return IOStreams{
		In:     in,
		Out:    io.Discard,
		ErrOut: io.Discard,
	}
}

var ioStreams = IOStreams{
	In:     os.Stdin,
	Out:    os.Stdout,
	ErrOut: os.Stderr,
}

// aos represents a root command options.
type cli struct {
	IOStreams
	fs *firestore.Client

	project string
	color   bool
	debug   bool
}

// NewCommand creates the aos root command.
func NewCommand(ctx context.Context, args []string) (*cobra.Command, error) {
	c := &cli{
		IOStreams: ioStreams,
	}

	cmd := &cobra.Command{
		Use:           appName,
		Short:         "Google Cloud Firestore command-line tool.",
		Version:       "v0.0.1",
		PreRunE:       func(cmd *cobra.Command, args []string) error { return c.NewClient(cmd.Context()) },
		PostRunE:      func(cmd *cobra.Command, args []string) error { return c.CloseClient() },
		SilenceErrors: true,
	}

	// set global flags
	f := cmd.PersistentFlags()
	f.BoolVar(&c.debug, "debug", false, "Use debug output")
	f.StringVar(&c.project, "project", "", "Google Cloud Firestore project ID")
	f.BoolVar(&c.color, "color", true, "Colorize output")

	// set subcommands
	cmd.AddCommand(c.Collection())
	cmd.AddCommand(c.Collections())
	cmd.AddCommand(c.Doc())
	cmd.AddCommand(c.Docs())

	// hack of PreRunE for initializes and closes firestore client
	for _, sc := range cmd.Commands() {
		sc.PreRunE = cmd.PreRunE
		sc.PostRunE = cmd.PostRunE
	}

	return cmd, nil
}

type checkType uint8

const (
	exactArgs checkType = iota
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
	}

	return nil
}
