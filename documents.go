// Copyright 2023 The firestore Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

// Docss represents a docs subcommand.
type Docs struct {
	*cli
}

var _ Command = (*Docs)(nil)

// Register implements Command.
func (ds *Docs) Register(cmd *cobra.Command) {
	cmd.AddCommand(ds.NewCommand())
}

// NewCommand implements Command.
func (ds *Docs) NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "docs",
		Aliases: []string{"ds"},
		Short:   "List the document",
		Long:    "List the document",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkArgs(cmd.Name(), exactArgs, 1, args...); err != nil {
				return err
			}

			return ds.Run(cmd.Context(), args)
		},
	}

	return cmd
}

// Run implements Command.
func (c *cli) Run(ctx context.Context, args []string) error {
	collPath := args[0]
	if collPath[len(collPath)-1] == '/' {
		return fmt.Errorf("invalid collection path: %q", collPath)
	}

	coll := c.fs.Collection(collPath)
	if coll == nil {
		return fmt.Errorf("not found %s documents", collPath)
	}

	iter := coll.DocumentRefs(ctx)
	w := bufio.NewWriter(c.Out)
	for {
		ref, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return fmt.Errorf("get next iterator result: %w", err)
		}

		w.WriteString(ref.ID + "\n")
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush stdout I/O: %w", err)
	}

	return nil
}
