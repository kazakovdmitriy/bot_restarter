package main

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestGetUptime(t *testing.T) {
	orig := uptimeCmd
	t.Cleanup(func() { uptimeCmd = orig })

	t.Run("success", func(t *testing.T) {
		uptimeCmd = func() (string, error) {
			return " 12:34  up 3 days,  2:15, 3 users, load averages: 1.23 0.45 0.12", nil
		}
		got, err := GetUptime()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(got, "up") {
			t.Errorf("unexpected uptime output: %s", got)
		}
	})

	t.Run("command fails", func(t *testing.T) {
		uptimeCmd = func() (string, error) {
			return "", errors.New("command not found")
		}
		_, err := GetUptime()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestReboot_Success(t *testing.T) {
	orig := rebootCmd
	t.Cleanup(func() { rebootCmd = orig })

	rebootCmd = func() error {
		return nil
	}

	if err := Reboot(); err != nil {
		t.Fatalf("Reboot() error = %v, want nil", err)
	}
}

func TestReboot_Failure(t *testing.T) {
	orig := rebootCmd
	t.Cleanup(func() { rebootCmd = orig })

	rebootCmd = func() error {
		return errors.New("permission denied")
	}

	if err := Reboot(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetUptime_LiveCommand(t *testing.T) {
	if _, err := exec.LookPath("uptime"); err != nil {
		t.Skip("uptime not available")
	}
	orig := uptimeCmd
	t.Cleanup(func() { uptimeCmd = orig })
	uptimeCmd = func() (string, error) {
		out, err := exec.Command("uptime").Output()
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	}
	got, err := GetUptime()
	if err != nil {
		t.Fatalf("GetUptime() error = %v", err)
	}
	if got == "" {
		t.Error("expected non-empty uptime")
	}
}
