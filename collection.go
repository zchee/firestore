// Copyright 2023 The firestore Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	json "github.com/goccy/go-json"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

// Collection represents a collection subcommand.
type Collection struct {
	*cli

	search string
}

var _ Command = (*Collection)(nil)

// Register implements Command.
func (coll *Collection) Register(cmd *cobra.Command) {
	cmd.AddCommand(coll.NewCommand())
}

// NewCommand implements Command.
func (coll *Collection) NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "collection [collection path]",
		Aliases: []string{"c"},
		Short:   "Describe the collection.",
		Long:    "Describe the collection.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkArgs(cmd.Name(), exactArgs, 1, args...); err != nil {
				return err
			}

			return coll.Run(cmd.Context(), args)
		},
	}

	f := cmd.Flags()
	f.StringVar(&coll.search, "search", "", "name of search documents")

	return cmd
}

// Run implements Command.
func (coll *Collection) Run(ctx context.Context, args []string) error {
	collPath := args[0]
	if strings.HasSuffix(collPath, "/") {
		return fmt.Errorf("invalid collection path: %s", collPath)
	}

	ref := coll.fs.Collection(collPath)
	iter := ref.Documents(ctx)

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
		return nil
	}

	if val := coll.search; val != "" {
		fixed, err := coll.searchFields(val, datas)
		if err != nil {
			return err
		}
		datas = fixed
	}

	enc := json.NewEncoder(coll.Out)
	enc.SetIndent("", "  ")
	optFuncs := []json.EncodeOptionFunc{
		json.DisableNormalizeUTF8(), // optimize
	}
	if coll.color {
		optFuncs = append(optFuncs, json.Colorize(colorScheme))
	}
	if err := enc.EncodeContext(ctx, datas, optFuncs...); err != nil {
		return fmt.Errorf("marshaling to json: %w", err)
	}

	return nil
}

type docField struct {
	val   string
	index int
}

func (coll *Collection) searchFields(val string, datas []map[string]interface{}) ([]map[string]interface{}, error) {
	ss := strings.SplitN(val, ":", 2)
	if len(ss) != 2 {
		return nil, fmt.Errorf("invalid --search flag value: %s", val)
	}

	k, v := ss[0], ss[1]

	var fields []docField
	for i, data := range datas {
		if d, ok := data[k]; ok {
			fields = append(fields, docField{
				val:   fmt.Sprint(d),
				index: i,
			})
		}
	}

	includes := make(map[int]bool)
	for _, field := range fields {
		if strings.Contains(field.val, v) {
			includes[field.index] = true
		}
	}

	tmpData := make([]map[string]interface{}, len(datas))
	copy(tmpData, datas)
	datas = []map[string]interface{}{}
	for idx := range includes {
		datas = append(datas, tmpData[idx])
	}

	return datas, nil
}
