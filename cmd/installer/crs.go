package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Stackdome/stackdome/install"
)

const (
	apiServerResourceName = "api-server"
	dbResourceName        = "db"
)

var externallyReachable bool

func applyCRs(vals install.TemplateValues, domain string) error {
	phaseLog(4, "Applying StackDome CRs...")

	_ = runShell("kubectl create namespace " + chartNamespace + " 2>/dev/null || true")

	stepLog("Applying Volume CR for database storage...")
	volumeManifest, err := install.ReadManifest("volume-cr.yaml")
	if err != nil {
		return fmt.Errorf("reading volume CR: %w", err)
	}
	if err := kubectlApply(volumeManifest); err != nil {
		return fmt.Errorf("applying volume CR: %w", err)
	}

	stepLog("Applying Stack CR...")
	stackManifest, err := install.ReadManifest("stack-cr.yaml")
	if err != nil {
		return fmt.Errorf("reading stack CR: %w", err)
	}
	if err := kubectlApply(stackManifest); err != nil {
		return fmt.Errorf("applying stack CR: %w", err)
	}

	stepLog("Applying db StackResource CR...")
	dbManifest, err := install.RenderManifest("db-resource-cr.yaml", vals)
	if err != nil {
		return fmt.Errorf("rendering db resource CR: %w", err)
	}
	if err := kubectlApply(dbManifest); err != nil {
		return fmt.Errorf("applying db resource CR: %w", err)
	}

	stepLog("Applying api-server StackResource CR...")
	apiManifest, err := install.RenderManifest("api-server-resource-cr.yaml", vals)
	if err != nil {
		return fmt.Errorf("rendering api-server resource CR: %w", err)
	}
	if err := kubectlApply(apiManifest); err != nil {
		return fmt.Errorf("applying api-server resource CR: %w", err)
	}

	stepLog("Waiting for API server to be reachable...")
	if err := waitForAPIServer(); err != nil {
		return err
	}

	successLog("StackDome CRs applied and API server is ready")

	checkExternalAccess(domain)

	return nil
}

func checkExternalAccess(domain string) {
	stepLog("Checking external access...")
	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("http://%s/health", domain)
	resp, err := client.Get(url)
	if err != nil {
		warnLog(fmt.Sprintf("Port 80 not reachable from the internet (tried %s)", domain))
		warnLog("Check your firewall rules -- ensure ports 80 and 443 are open for inbound traffic")
		externallyReachable = false
		return
	}
	_ = resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		successLog(fmt.Sprintf("Dashboard reachable from the internet at %s", domain))
		externallyReachable = true
	} else {
		warnLog(fmt.Sprintf("Port 80 reachable but health check returned %d", resp.StatusCode))
		externallyReachable = false
	}
}

func waitForAPIServer() error {
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		healthURL := apiServerHealthURL()
		if healthURL == "" {
			stepLog("Waiting for api-server service to be created...")
			time.Sleep(5 * time.Second)
			continue
		}
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(healthURL)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timed out waiting for API server")
}

func apiServerHealthURL() string {
	svcIP, err := output("kubectl", "get", "svc", apiServerResourceName,
		"-n", chartNamespace,
		"-o", "jsonpath={.spec.clusterIP}")
	if err != nil || svcIP == "" {
		return ""
	}
	return fmt.Sprintf("http://%s:8000/health", svcIP)
}

func apiServerBaseURL() string {
	svcIP, err := output("kubectl", "get", "svc", apiServerResourceName,
		"-n", chartNamespace,
		"-o", "jsonpath={.spec.clusterIP}")
	if err != nil || svcIP == "" {
		return "http://localhost:8000"
	}
	return fmt.Sprintf("http://%s:8000", svcIP)
}
