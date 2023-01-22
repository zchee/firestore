// Copyright 2023 The firestore Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"errors"
	"fmt"

	json "github.com/goccy/go-json"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

func (c *cli) Collection() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "collection [collection path]",
		Aliases: []string{"c"},
		Short:   "Describe the collection",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkArgs(cmd.Name(), exactArgs, 1, args...); err != nil {
				return err
			}

			collPath := args[0]
			ref := c.fs.Collection(collPath)
			iter := ref.Documents(cmd.Context())

			var datas []map[string]interface{}
			for {
				docsnap, err := iter.Next()
				if err != nil {
					if errors.Is(err, iterator.Done) {
						break
					}
					return fmt.Errorf("get next iterator result: %w", err)
				}

				data := make(map[string]interface{})
				if err := docsnap.DataTo(&data); err != nil {
					return fmt.Errorf("populate %s collection: %w", docsnap.Ref.ID, err)
				}
				if len(data) == 0 {
					continue
				}

				data["id"] = docsnap.Ref.ID
				data["path"] = docsnap.Ref.Path
				data["createTime"] = docsnap.CreateTime.Format(RFC3339Milli)
				data["readTime"] = docsnap.ReadTime.Format(RFC3339Milli)
				data["updateTime"] = docsnap.UpdateTime.Format(RFC3339Milli)
				datas = append(datas, data)
			}

			if len(datas) == 0 {
				return fmt.Errorf("not found %s collection", collPath)
			}

			enc := json.NewEncoder(c.Out)
			enc.SetIndent("", "  ")
			optFuncs := []json.EncodeOptionFunc{
				json.DisableNormalizeUTF8(), // optimize
			}
			if c.color {
				optFuncs = append(optFuncs, json.Colorize(colorScheme))
			}
			if err := enc.EncodeContext(cmd.Context(), datas, optFuncs...); err != nil {
				return fmt.Errorf("marshaling to json: %w", err)
			}

			return nil
		},
	}

	return cmd
}
