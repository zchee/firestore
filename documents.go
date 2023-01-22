// Copyright 2023 The firestore Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bufio"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

func (c *cli) Docs() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "docs",
		Aliases: []string{"ds"},
		Short:   "List the document",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkArgs(cmd.Name(), exactArgs, 1, args...); err != nil {
				return err
			}

			collPath := args[0]
			if collPath[len(collPath)-1] == '/' {
				return fmt.Errorf("invalid collection path: %q", collPath)
			}
			coll := c.fs.Collection(collPath)
			if coll == nil {
				return fmt.Errorf("not found %s documents", collPath)
			}

			iter := coll.DocumentRefs(cmd.Context())
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
		},
	}

	return cmd
}
