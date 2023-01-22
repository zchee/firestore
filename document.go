// Copyright 2023 The firestore Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	json "github.com/goccy/go-json"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *cli) Doc() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "doc {document path}",
		Aliases: []string{"d"},
		Short:   "Describe the document",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkArgs(cmd.Name(), exactArgs, 1, args...); err != nil {
				return err
			}

			path := args[0]
			collPath, docPath, ok := CutLast(path, "/")
			if !ok {
				return fmt.Errorf("invalid arguments: %q", path)
			}

			var docs []*firestore.DocumentSnapshot

			ref := c.fs.Collection(collPath).Doc(docPath)
			doc, err := ref.Get(cmd.Context())
			if err == nil {
				docs = append(docs, doc)
			} else {
				st := status.Convert(err)
				if st.Code() == codes.InvalidArgument {
					return fmt.Errorf("get document snapshot: %w", err)
				}

				docPath, ok = CutSuffix(docPath, "/")
				if !ok {
					return fmt.Errorf("invalid arguments: %q", path)
				}
				iter := c.fs.Collection(collPath + "/" + docPath).Documents(cmd.Context())
				for {
					docsnap, err := iter.Next()
					if err != nil {
						if errors.Is(err, iterator.Done) {
							break
						}
						return fmt.Errorf("get next iterator result: %w", err)
					}

					docs = append(docs, docsnap)
				}
			}

			var datas []map[string]interface{}
			for _, d := range docs {
				data := make(map[string]interface{})
				if err := d.DataTo(&data); err != nil {
					return fmt.Errorf("populate %s document: %w", doc.Ref.ID, err)
				}
				if len(data) == 0 {
					continue
				}

				data["id"] = d.Ref.ID
				data["path"] = d.Ref.Path
				data["createTime"] = d.CreateTime.Format(RFC3339Milli)
				data["readTime"] = d.ReadTime.Format(RFC3339Milli)
				data["updateTime"] = d.UpdateTime.Format(RFC3339Milli)
				datas = append(datas, data)
			}

			if len(datas) == 0 {
				return fmt.Errorf("not found %s document", path)
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
