// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

// +build !windows

package main

import "storj.io/private/process"

func main() {
	process.Exec(rootCmd)
}
