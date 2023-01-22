// Copyright 2023 The firestore Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := NewCommand(ctx, os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := c.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
