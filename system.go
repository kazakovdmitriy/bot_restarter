package main

import (
	"fmt"
	"os/exec"
	"strings"
)

var (
	uptimeCmd = func() (string, error) {
		out, err := exec.Command("uptime").Output()
		if err != nil {
			return "", fmt.Errorf("uptime: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	}
	rebootCmd = func() error {
		return exec.Command("sudo", "reboot").Run()
	}
)

func GetUptime() (string, error) {
	return uptimeCmd()
}

func Reboot() error {
	return rebootCmd()
}
