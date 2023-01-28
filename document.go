// Copyright 2023 The firestore Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"cloud.google.com/go/firestore"
	json "github.com/goccy/go-json"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Doc represents a doc subcommand.
type Doc struct {
	*cli
}

var _ Command = (*Doc)(nil)

// Register implements Command.
func (d *Doc) Register(cmd *cobra.Command) {
	cmd.AddCommand(d.NewCommand())
}

// NewCommand implements Command.
func (d *Doc) NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "doc {document path}",
		Aliases: []string{"d"},
		Short:   "Describe the document",
		Long:    "Describe the document",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkArgs(cmd.Name(), exactArgs, 1, args...); err != nil {
				return err
			}

			return d.Run(cmd.Context(), args)
		},
	}
}

// Run implements Command.
func (d *Doc) Run(ctx context.Context, args []string) error {
	path := args[0]
	collPath, docPath, ok := CutLast(path, "/")
	if !ok {
		return fmt.Errorf("invalid arguments: %q", path)
	}

	var docs []*firestore.DocumentSnapshot

	ref := d.fs.Collection(collPath).Doc(docPath)
	snap, err := ref.Get(ctx)
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.InvalidArgument {
			return fmt.Errorf("get document snapshot: %w", err)
		}

		// fallback to gets documents
		if err := d.parseDocuments(ctx, collPath, docPath, docs); err != nil {
			return err
		}
	} else {
		docs = append(docs, snap)
	}

	datas, err := d.resolveDocs(ref.ID, docs)
	if err != nil {
		return err
	}

	if err := d.marshal(ctx, d.Out, datas); err != nil {
		return err
	}

	return nil
}

func (d *Doc) parseDocuments(ctx context.Context, collPath, docPath string, docs []*firestore.DocumentSnapshot) error {
	path, ok := CutSuffix(docPath, "/")
	if !ok {
		return fmt.Errorf("invalid arguments: %q", docPath)
	}

	iter := d.fs.Collection(collPath + "/" + path).Documents(ctx)
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

	return nil
}

func (d *Doc) resolveDocs(dotID string, docs []*firestore.DocumentSnapshot) ([]map[string]interface{}, error) {
	var datas []map[string]interface{}

	for _, doc := range docs {
		data := make(map[string]interface{})
		if err := doc.DataTo(&data); err != nil {
			return nil, fmt.Errorf("populate %s document: %w", dotID, err)
		}
		if len(data) == 0 {
			continue
		}

		data["id"] = doc.Ref.ID
		data["path"] = doc.Ref.Path
		data["createTime"] = doc.CreateTime.Format(RFC3339Milli)
		data["readTime"] = doc.ReadTime.Format(RFC3339Milli)
		data["updateTime"] = doc.UpdateTime.Format(RFC3339Milli)
		datas = append(datas, data)
	}

	if len(datas) == 0 {
		return nil, errors.New("not found document")
	}

	return datas, nil
}

func (d *Doc) marshal(ctx context.Context, w io.Writer, datas []map[string]interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	optFuncs := []json.EncodeOptionFunc{
		json.DisableNormalizeUTF8(), // optimize
	}
	if d.color {
		optFuncs = append(optFuncs, json.Colorize(colorScheme))
	}
	if err := enc.EncodeContext(ctx, datas, optFuncs...); err != nil {
		return fmt.Errorf("marshaling to json: %w", err)
	}

	return nil
}
