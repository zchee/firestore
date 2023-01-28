package main

import (
	"bytes"
	"context"
	"io"

	"github.com/spf13/cobra"
)

// Command represents a firestore sub-command methods.
type Command interface {
	Register(cmd *cobra.Command)
	NewCommand() *cobra.Command
	Run(ctx context.Context, args []string) error
}

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
