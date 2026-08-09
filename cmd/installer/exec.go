package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = installerOutput.stderr
	cmd.Stderr = installerOutput.stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w", commandLabel(name, args), err)
	}
	return nil
}

func output(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = installerOutput.stderr
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s failed: %w", commandLabel(name, args), err)
	}
	return strings.TrimSpace(out.String()), nil
}

// outputQuiet is output() for probes whose failure is expected and handled by
// the caller -- it keeps the command's stderr off the console.
func outputQuiet(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

func runShell(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = installerOutput.stderr
	cmd.Stderr = installerOutput.stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("shell command failed: %w", err)
	}
	return nil
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func kubectlApply(manifest []byte) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = bytes.NewReader(manifest)
	cmd.Stdout = installerOutput.stderr
	cmd.Stderr = installerOutput.stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kubectl apply failed: %w", err)
	}
	return nil
}

func commandLabel(name string, args []string) string {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return name
	}
	return name + " " + args[0]
}
