package main

import (
	"fmt"
	"time"
)

const (
	chartRepo      = "oci://quay.io/stackdome/charts/stackdome-agent"
	chartVersion   = "0.6.7-alpha"
	chartRelease   = "stackdome-agent"
	chartNamespace = "stackdome-control-plane"
)

func installHelm() error {
	if commandExists("helm") {
		stepLog("Helm already installed -- skipping")
		return nil
	}

	stepLog("Installing Helm...")
	if err := runShell("curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"); err != nil {
		return fmt.Errorf("helm install failed: %w", err)
	}

	return nil
}

func installStackdomeAgent() error {
	phaseLog(3, "Installing stackdome-agent chart...")

	if err := installHelm(); err != nil {
		return err
	}

	stepLog(fmt.Sprintf("Installing stackdome-agent chart v%s...", chartVersion))
	if err := run("helm", "upgrade", "--install", chartRelease, chartRepo,
		"--version", chartVersion,
		"--namespace", chartNamespace,
		"--create-namespace",
		"--wait",
		"--timeout", "10m",
	); err != nil {
		return fmt.Errorf("stackdome-agent chart install failed: %w", err)
	}

	stepLog("Waiting for operator pods to be ready...")
	if err := waitForOperatorPods(); err != nil {
		return err
	}

	successLog("stackdome-agent is ready")
	return nil
}

func waitForOperatorPods() error {
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		err := run("kubectl", "wait", "--for=condition=Ready", "pod",
			"--all", "-n", chartNamespace, "--timeout=10s")
		if err == nil {
			return nil
		}
		time.Sleep(10 * time.Second)
	}
	return fmt.Errorf("timed out waiting for operator pods in %s", chartNamespace)
}
