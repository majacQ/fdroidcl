// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package main

import (
	"errors"
	"fmt"
	"strconv"

	"mvdan.cc/fdroidcl/adb"
)

var cmdUninstall = &Command{
	UsageLine: "uninstall <appid...>",
	Short:     "Uninstall an app",
}

var (
	uninstallUser = cmdUninstall.Fset.String("user", "all", "Uninstall for specified user <USER_ID|current|all>")
)

func init() {
	cmdUninstall.Run = runUninstall
}

func runUninstall(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no package names given")
	}
	device, err := oneDevice()
	if err != nil {
		return err
	}
	inst, err := device.Installed()
	if err != nil {
		return err
	}
	if *uninstallUser != "all" && *uninstallUser != "current" {
		n, err := strconv.Atoi(*uninstallUser)
		if err != nil {
			return fmt.Errorf("-user has to be <USER_ID|current|all>")
		}
		if n < 0 {
			return fmt.Errorf("-user cannot have a negative number as USER_ID")
		}
		allUids := adb.AllUserIds(inst)
		if _, exists := allUids[n]; !exists {
			return fmt.Errorf("user %d does not exist", n)
		}
	}
	if *uninstallUser == "current" {
		uid, err := device.CurrentUserId()
		if err != nil {
			return err
		}
		*uninstallUser = strconv.Itoa(uid)
	}
	for _, id := range args {
		var err error
		fmt.Printf("Uninstalling %s\n", id)
		app, installed := inst[id]
		if installed {
			installedForUser := false
			if *uninstallUser == "all" {
				installedForUser = true
			} else {
				uid, err := strconv.Atoi(*uninstallUser)
				if err != nil {
					return err
				}
				for _, appUser := range app.InstalledForUsers {
					if appUser == uid {
						installedForUser = true
						break
					}
				}
			}
			if installedForUser {
				if *uninstallUser == "all" {
					err = device.Uninstall(id)
				} else {
					err = device.UninstallUser(id, *uninstallUser)
				}
			} else {
				err = errors.New("not installed for user")
			}
		} else {
			err = errors.New("not installed")
		}
		if err != nil {
			return fmt.Errorf("could not uninstall %s: %v", id, err)
		}
	}
	return nil
}
