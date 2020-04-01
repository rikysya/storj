// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

// Implements support for running the storagenode-updater as a Windows Service.
//
// The Windows Service can be created with sc.exe, e.g.
//
// sc.exe create storagenode-updater binpath= "C:\Users\MyUser\storagenode-updater.exe run ..."

// +build windows

package main

import (
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/windows/svc"

	"storj.io/private/process"
)

func main() {
	isInteractive, err := svc.IsAnInteractiveSession()
	if err != nil {
		zap.S().Fatalf("Failed to determine if session is interactive: %v", err)
	}

	if isInteractive {
		process.Exec(rootCmd)
		return
	}

	err = svc.Run("storagenode-updater", &service{})
	if err != nil {
		zap.S().Fatalf("Service failed: %v", err)
	}
}

type service struct{}

func (m *service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	changes <- svc.Status{State: svc.StartPending}

	var group errgroup.Group
	group.Go(func() error {
		process.Exec(rootCmd)
		return nil
	})

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	for c := range r {
		switch c.Cmd {
		case svc.Interrogate:
			zap.S().Info("Interrogate request received.")
		case svc.Stop, svc.Shutdown:
			zap.S().Info("Stop/Shutdown request received.")
			changes <- svc.Status{State: svc.StopPending}
			// Cancel the command's root context to cleanup resources
			_, cancel := process.Ctx(runCmd)
			cancel()
			_ = group.Wait() // process.Exec does not return an error
			// After returning the Windows Service is stopped and the process terminates
			return
		default:
			zap.S().Infof("Unexpected control request: %d\n", c)
		}
	}

	return
}
