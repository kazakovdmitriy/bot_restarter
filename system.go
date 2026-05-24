package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetUptime() (string, error) {
	out, err := exec.Command("uptime").Output()
	if err != nil {
		return "", fmt.Errorf("uptime: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func Reboot() error {
	return exec.Command("sudo", "reboot").Run()
}
