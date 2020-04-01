// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

// +build unittest !windows

package main

import (
	"context"
	"os"

	"github.com/zeebo/errs"
)

func restartService(ctx context.Context, service, binaryLocation, newVersionPath, backupPath string) error {
	if err := os.Rename(binaryLocation, backupPath); err != nil {
		return errs.Wrap(err)
	}

	if err := os.Rename(newVersionPath, binaryLocation); err != nil {
		return errs.Combine(err, os.Rename(backupPath, binaryLocation), os.Remove(newVersionPath))
	}

	return nil
}
