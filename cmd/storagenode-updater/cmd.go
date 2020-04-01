// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"storj.io/common/errs2"
	"storj.io/common/fpath"
	"storj.io/common/identity"
	"storj.io/common/storj"
	"storj.io/common/sync2"
	"storj.io/private/cfgstruct"
	"storj.io/private/process"
	_ "storj.io/storj/private/version" // This attaches version information during release builds.
	"storj.io/storj/private/version/checker"
)

const (
	updaterServiceName = "storagenode-updater"
	minCheckInterval   = time.Minute
)

var (
	cancel context.CancelFunc
	// TODO: replace with config value of random bytes in storagenode config.
	nodeID storj.NodeID

	rootCmd = &cobra.Command{
		Use:   "storagenode-updater",
		Short: "Version updater for storage node",
	}
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run the storagenode-updater for storage node",
		Args:  cobra.OnlyValidArgs,
		RunE:  cmdRun,
	}

	runCfg struct {
		// TODO: check interval default has changed from 6 hours to 15 min.
		checker.Config
		Identity identity.Config

		BinaryLocation string `help:"the storage node executable binary location" default:"storagenode.exe"`
		ServiceName    string `help:"storage node OS service name" default:"storagenode"`
		// deprecated
		Log string `help:"deprecated, use --log.output" default:""`
	}

	confDir     string
	identityDir string
)

func init() {
	// TODO: this will probably generate warnings for mismatched config fields.
	defaultConfDir := fpath.ApplicationDir("storj", "storagenode")
	defaultIdentityDir := fpath.ApplicationDir("storj", "identity", "storagenode")
	cfgstruct.SetupFlag(zap.L(), rootCmd, &confDir, "config-dir", defaultConfDir, "main directory for storagenode configuration")
	cfgstruct.SetupFlag(zap.L(), rootCmd, &identityDir, "identity-dir", defaultIdentityDir, "main directory for storagenode identity credentials")
	defaults := cfgstruct.DefaultsFlag(rootCmd)

	rootCmd.AddCommand(runCmd)

	process.Bind(runCmd, &runCfg, defaults, cfgstruct.ConfDir(confDir), cfgstruct.IdentityDir(identityDir))
}

func cmdRun(cmd *cobra.Command, args []string) (err error) {
	err = openLog()
	if err != nil {
		zap.S().Errorf("Error creating new logger: %v", err)
	}

	if !fileExists(runCfg.BinaryLocation) {
		zap.S().Fatal("Unable to find storage node executable binary")
	}

	ident, err := runCfg.Identity.Load()
	if err != nil {
		zap.S().Fatalf("Error loading identity: %v", err)
	}
	nodeID = ident.ID
	if nodeID.IsZero() {
		zap.S().Fatal("Empty node ID")
	}

	var ctx context.Context
	ctx, cancel = process.Ctx(cmd)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c

		signal.Stop(c)
		cancel()
	}()

	loopFunc := func(ctx context.Context) (err error) {
		all, err := checker.New(runCfg.ClientConfig).All(ctx)
		if err != nil {
			zap.S().Errorf("Error retrieving version info: %v", err)
			return nil
		}

		if err := update(ctx, runCfg.ServiceName, runCfg.BinaryLocation, all.Processes.Storagenode); err != nil {
			// don't finish loop in case of error just wait for another execution
			zap.S().Errorf("Error updating %s: %v", runCfg.ServiceName, err)
		}

		updaterBinName := os.Args[0]
		if err := update(ctx, updaterServiceName, updaterBinName, all.Processes.StoragenodeUpdater); err != nil {
			// don't finish loop in case of error just wait for another execution
			zap.S().Errorf("Error updating %s: %v", updaterServiceName, err)
		}
		return nil
	}

	switch {
	case runCfg.CheckInterval <= 0:
		err = loopFunc(ctx)
	case runCfg.CheckInterval < minCheckInterval:
		zap.S().Errorf("Check interval below minimum: %s, setting to %s", runCfg.CheckInterval, minCheckInterval)
		runCfg.CheckInterval = minCheckInterval
		fallthrough
	default:
		loop := sync2.NewCycle(runCfg.CheckInterval)
		err = loop.Run(ctx, loopFunc)
	}
	if err != nil && !errs2.IsCanceled(err) {
		log.Fatal(err)
	}
	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return info.Mode().IsRegular()
}

func openLog() error {
	if runCfg.Log != "" {
		logPath := runCfg.Log
		if runtime.GOOS == "windows" && !strings.HasPrefix(logPath, "winfile:///") {
			logPath = "winfile:///" + logPath
		}
		logger, err := process.NewLoggerWithOutputPaths(logPath)
		if err != nil {
			return err
		}
		zap.ReplaceGlobals(logger)
	}
	return nil
}
