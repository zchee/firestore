// Copyright 2023 The firestore Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

func (c *cli) Collections() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "collections [collection path]",
		Aliases: []string{"cs"},
		Short:   "List the collections",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkArgs(cmd.Name(), maxArgs, 1, args...); err != nil {
				return err
			}

			// default is root level collections
			iter := c.fs.Collections(cmd.Context())
			// set sub-collection if exists args
			if len(args) == 1 {
				path := args[0]
				collPath, docPath, ok := strings.Cut(path, "/")
				if ok {
					iter = c.fs.Collection(collPath).Doc(docPath).Collections(cmd.Context())
				}
			}

			w := bufio.NewWriter(c.Out)
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
		},
	}

	return cmd
}
