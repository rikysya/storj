// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"context"
	"os"

	"github.com/zeebo/errs"
	"go.uber.org/zap"

	"storj.io/private/version"
)

func update(ctx context.Context, serviceName, binaryLocation string, ver version.Process) error {
	suggestedVersion, err := ver.Suggested.SemVer()
	if err != nil {
		return errs.Wrap(err)
	}

	var currentVersion version.SemVer
	if serviceName == updaterServiceName {
		// TODO: find better way to check this binary version
		currentVersion = version.Build.Version
	} else {
		currentVersion, err = binaryVersion(binaryLocation)
		if err != nil {
			return errs.Wrap(err)
		}
	}

	// should update
	if currentVersion.Compare(suggestedVersion) >= 0 {
		zap.S().Infof("%s version is up to date", serviceName)
		return nil
	}
	if !version.ShouldUpdate(ver.Rollout, nodeID) {
		zap.S().Infof("New %s version available but not rolled out to this nodeID yet", serviceName)
		return nil
	}

	var backupPath string
	if serviceName == updaterServiceName {
		backupPath = prependExtension(binaryLocation, "old")
	} else {
		backupPath = prependExtension(binaryLocation, "old."+currentVersion.String())
	}

	newVersionPath := prependExtension(binaryLocation, ver.Suggested.Version)

	if err = downloadBinary(ctx, parseDownloadURL(ver.Suggested.URL), newVersionPath); err != nil {
		return errs.Wrap(err)
	}

	downloadedVersion, err := binaryVersion(newVersionPath)
	if err != nil {
		return errs.Combine(errs.Wrap(err), os.Remove(newVersionPath))
	}

	if suggestedVersion.Compare(downloadedVersion) != 0 {
		err := errs.New("invalid version downloaded: wants %s got %s",
			suggestedVersion.String(),
			downloadedVersion.String(),
		)
		return errs.Combine(err, os.Remove(newVersionPath))
	}

	zap.S().Infof("Restarting service %s", serviceName)

	if err = restartService(ctx, serviceName, binaryLocation, newVersionPath, backupPath); err != nil {
		return errs.Wrap(err)
	}

	zap.S().Infof("Service %s restarted successfully", serviceName)
	return nil
}
