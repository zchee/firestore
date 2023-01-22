// Copyright 2023 The firestore Authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	firestoreapiv1 "cloud.google.com/go/firestore/apiv1"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	grpc_oauth "google.golang.org/grpc/credentials/oauth"
)

// RFC3339Milli is the time format layout in RFC3339 format with millisecond precision.
//
// This is the time format of the response from Firestore.
const RFC3339Milli = "2006-01-02T15:04:05.999999Z"

// NewClient finds the Google default application credentials,
// and creates a new Firestore client that sets it to the grpc.DialOption.
func (c *cli) NewClient(ctx context.Context) error {
	if c.project == "" {
		return errors.New("requires project flag")
	}

	perCreds, err := grpc_oauth.NewApplicationDefault(ctx, firestoreapiv1.DefaultAuthScopes()...)
	if err != nil {
		return fmt.Errorf("find application default credentials: %w", err)
	}

	fs, err := firestore.NewClient(ctx, c.project,
		option.WithGRPCDialOption(grpc.WithPerRPCCredentials(perCreds)),
	)
	if err != nil {
		return fmt.Errorf("create new firestore client: %w", err)
	}
	c.fs = fs

	return nil
}

// CloseClient closes firestore client.
func (c *cli) CloseClient() error {
	if c.fs != nil {
		if err := c.fs.Close(); err != nil {
			return fmt.Errorf("close firestore client: %w", err)
		}
	}
	return nil
}
