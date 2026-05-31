package main

import (
	"fmt"
	"os"
	"time"
)

const (
	k3sKubeconfigPath = "/etc/rancher/k3s/k3s.yaml"
)

func installK3s() error {
	phaseLog(2, "Installing k3s...")

	if commandExists("k3s") {
		stepLog("k3s already installed -- skipping")
	} else {
		stepLog("Downloading and installing k3s...")
		if err := runShell("curl -sfL https://get.k3s.io | sh -s - --disable=traefik"); err != nil {
			return fmt.Errorf("k3s install failed: %w", err)
		}
	}

	stepLog("Setting up kubeconfig...")
	if err := setupKubeconfig(); err != nil {
		return fmt.Errorf("kubeconfig setup failed: %w", err)
	}

	stepLog("Waiting for k3s node to be ready...")
	if err := waitForK3s(); err != nil {
		return fmt.Errorf("k3s not ready: %w", err)
	}

	successLog("k3s is ready")
	return nil
}

func setupKubeconfig() error {
	os.Setenv("KUBECONFIG", k3sKubeconfigPath)
	os.Chmod(k3sKubeconfigPath, 0600)
	return nil
}

func waitForK3s() error {
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		err := run("kubectl", "wait", "--for=condition=Ready",
			"node", "--all", "--timeout=10s")
		if err == nil {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timed out after 2 minutes waiting for k3s nodes")
}
