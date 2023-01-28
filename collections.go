// Copyright 2023 The firestore Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

// Collections represents a collections subcommand.
type Collections struct {
	*cli
}

var _ Command = (*Collections)(nil)

// Register implements Command.
func (colls *Collections) Register(cmd *cobra.Command) {
	cmd.AddCommand(colls.NewCommand())
}

// NewCommand implements Command.
func (colls *Collections) NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "collections [collection path]",
		Aliases: []string{"cs"},
		Short:   "List the collections",
		Long:    "List the collections",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkArgs(cmd.Name(), maxArgs, 1, args...); err != nil {
				return err
			}

			return colls.Run(cmd.Context(), args)
		},
	}

	return cmd
}

// Run implements Command.
func (colls *Collections) Run(ctx context.Context, args []string) error {
	// default is root level collections
	iter := colls.fs.Collections(ctx)

	// set sub-collection if args is non-nil
	if len(args) == 1 {
		path := args[0]
		collPath, docPath, ok := strings.Cut(path, "/")
		if ok {
			iter = colls.fs.Collection(collPath).Doc(docPath).Collections(ctx)
		}
	}

	w := bufio.NewWriter(colls.Out)
	for {
		col, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return fmt.Errorf("get next iterator result: %w", err)
		}

		w.WriteString(col.ID + "\n")
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush stdout I/O: %w", err)
	}

	return nil
}
