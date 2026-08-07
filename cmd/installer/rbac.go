package main

import (
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/install"
)

const apiServerServiceAccountSecret = "stackdome-api-server-account-secret"

func applyAPIServerRBAC() error {
	stepLog("Applying API server RBAC...")
	manifest, err := install.ReadManifest("rbac.yaml")
	if err != nil {
		return fmt.Errorf("reading RBAC manifest: %w", err)
	}
	if err := kubectlApply(manifest); err != nil {
		return fmt.Errorf("applying RBAC: %w", err)
	}
	if err := waitForAPIServerCredentials(); err != nil {
		return err
	}
	stepLog("API server RBAC applied")
	return nil
}

func waitForAPIServerCredentials() error {
	deadline := time.Now().Add(time.Minute)
	for time.Now().Before(deadline) {
		token, tokenErr := outputQuiet(
			"kubectl", "get", "secret", apiServerServiceAccountSecret,
			"-n", chartNamespace,
			"-o", "jsonpath={.data.token}",
		)
		caData, caErr := outputQuiet(
			"kubectl", "get", "secret", apiServerServiceAccountSecret,
			"-n", chartNamespace,
			"-o", "jsonpath={.data.ca\\.crt}",
		)
		if tokenErr == nil && caErr == nil && token != "" && caData != "" {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timed out waiting for API server service-account credentials")
}
