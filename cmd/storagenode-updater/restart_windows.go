// Copyright (C) 2020 Storj Labs, Inc.
// See LICENSE for copying information.

// +build windows,!unittest

package main

import (
	"context"
	"os"
	"time"

	"github.com/zeebo/errs"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func restartService(ctx context.Context, service, binaryLocation, newVersionPath, backupPath string) error {
	manager, err := mgr.Connect()
	if err != nil {
		return errs.Combine(errs.Wrap(err), os.Remove(newVersionPath))
	}
	defer func() {
		err = errs.Combine(errs.Wrap(manager.Disconnect()))
	}()

	srvc, err := manager.OpenService(service)
	if err != nil {
		return errs.Combine(errs.Wrap(err), os.Remove(newVersionPath))
	}
	defer func() {
		err = errs.Combine(errs.Wrap(srvc.Close()))
	}()

	status, err := srvc.Query()
	if err != nil {
		return errs.Combine(errs.Wrap(err), os.Remove(newVersionPath))
	}

	// stop service if it's not stopped
	if status.State != svc.Stopped && status.State != svc.StopPending {
		if err = serviceControl(srvc, ctx, svc.Stop, svc.Stopped, time.Now().Add(time.Second*10)); err != nil {
			return errs.Combine(errs.Wrap(err), os.Remove(newVersionPath))
		}
		// if it is stopping wait for it to complete
	} else if status.State == svc.StopPending {
		if err = serviceWaitForState(srvc, ctx, svc.Stopped, time.Now().Add(time.Second*10)); err != nil {
			return errs.Combine(errs.Wrap(err), os.Remove(newVersionPath))
		}
	}

	// error during substitution
	recovered, err := substituteWithRecovery(binaryLocation, backupPath, newVersionPath)
	if err != nil {
		if recovered {
			return errs.Combine(errs.Wrap(err), srvc.Start(), os.Remove(newVersionPath))
		}

		// if not recovered just end
		return errs.Combine(errs.Wrap(err), os.Remove(newVersionPath))
	}

	// successfully substituted binaries
	err = retry(2,
		func() (bool, error) {
			if err := ctx.Err(); err != nil {
				return true, err
			}
			return false, srvc.Start()
		},
	)
	// if fail to start the service, try again with backup
	if err != nil {
		if rerr := os.Rename(backupPath, binaryLocation); rerr != nil {
			return errs.Combine(err, rerr)
		}

		return errs.Combine(err, srvc.Start())
	}

	return nil
}

func serviceControl(service *mgr.Service, ctx context.Context, cmd svc.Cmd, state svc.State, timeout time.Time) error {
	status, err := service.Control(cmd)
	if err != nil {
		return err
	}

	for status.State != state {
		if err := ctx.Err(); err != nil {
			return err
		}
		if timeout.Before(time.Now()) {
			return errs.New("timeout")
		}

		status, err = service.Query()
		if err != nil {
			return err
		}
	}

	return nil
}

func serviceWaitForState(service *mgr.Service, ctx context.Context, state svc.State, timeout time.Time) error {
	status, err := service.Query()
	if err != nil {
		return err
	}

	for status.State != state {
		if err := ctx.Err(); err != nil {
			return err
		}
		if timeout.Before(time.Now()) {
			return errs.New("timeout")
		}

		status, err = service.Query()
		if err != nil {
			return err
		}
	}

	return nil
}

func retry(count int, cb func() (bool, error)) error {
	var err error
	var abort bool

	for i := 0; i < count; i++ {
		if abort, err = cb(); err == nil {
			return nil
		}
		if abort {
			return err
		}
	}

	return err
}

func substituteWithRecovery(target, backup, new string) (bool, error) {
	if err := os.Rename(target, backup); err != nil {
		return false, errs.Wrap(err)
	}

	if err := os.Rename(new, target); err != nil {
		if rerr := os.Rename(backup, target); rerr != nil {
			return false, errs.Combine(err, rerr)
		}

		return true, errs.Wrap(err)
	}

	return false, nil
}
