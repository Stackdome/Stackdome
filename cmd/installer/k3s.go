package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	k3sKubeconfigPath = "/etc/rancher/k3s/k3s.yaml"
	k3sConfigPath     = "/etc/rancher/k3s/config.yaml"
	maxPodsKey        = "max-pods"
	k3sMaxPods        = 250
)

func installK3s() error {
	phaseLog(2, "Installing k3s...")

	wroteConfig, err := writeK3sConfig()
	if err != nil {
		return fmt.Errorf("k3s config: %w", err)
	}

	if commandExists("k3s") {
		stepLog("k3s already installed -- skipping")
		if wroteConfig {
			stepLog("Restarting k3s to pick up the new config...")
			if err := run("systemctl", "restart", "k3s"); err != nil {
				return fmt.Errorf("restarting k3s: %w", err)
			}
		}
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

// kubelet defaults to 110 pods per node, which a single-box install reaches
// well before it runs out of CPU or memory. The node CIDR is a /24, so 250 is
// the practical ceiling. Reports whether it wrote the file.
func writeK3sConfig() (bool, error) {
	desired := fmt.Sprintf("kubelet-arg:\n  - \"max-pods=%d\"\n", k3sMaxPods)

	existing, err := os.ReadFile(k3sConfigPath)
	switch {
	case err == nil && strings.Contains(string(existing), maxPodsKey):
		stepLog("k3s config already sets max-pods -- leaving it alone")
		return false, nil
	case err == nil:
		warnLog(fmt.Sprintf("%s exists without %s -- not overwriting it", k3sConfigPath, maxPodsKey))
		return false, nil
	case !os.IsNotExist(err):
		return false, fmt.Errorf("reading %s: %w", k3sConfigPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(k3sConfigPath), 0755); err != nil {
		return false, fmt.Errorf("creating %s: %w", filepath.Dir(k3sConfigPath), err)
	}
	if err := os.WriteFile(k3sConfigPath, []byte(desired), 0644); err != nil {
		return false, fmt.Errorf("writing %s: %w", k3sConfigPath, err)
	}
	stepLog(fmt.Sprintf("Wrote %s with max-pods=%d", k3sConfigPath, k3sMaxPods))
	return true, nil
}

func setupKubeconfig() error {
	if err := os.Setenv("KUBECONFIG", k3sKubeconfigPath); err != nil {
		return fmt.Errorf("setting KUBECONFIG: %w", err)
	}
	if err := os.Chmod(k3sKubeconfigPath, 0600); err != nil {
		return fmt.Errorf("chmod kubeconfig: %w", err)
	}
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
