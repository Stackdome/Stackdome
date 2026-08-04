package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

const (
	crdsDirName         = "crds"
	templatesDirName    = "templates"
	chartRepo           = "oci://quay.io/stackdome/charts/stackdome-agent"
	defaultChartVersion = "0.6.11-alpha-rc5"
	chartRelease        = "stackdome-agent"
	chartNamespace      = "stackdome-control-plane"
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

func installStackdomeAgent(chartVersion string) error {
	phaseLog(3, "Installing stackdome-agent chart...")

	if err := installHelm(); err != nil {
		return err
	}

	if deployed := deployedChartVersion(); deployed == chartVersion {
		stepLog(fmt.Sprintf("stackdome-agent already at v%s -- skipping chart upgrade", deployed))
		successLog("stackdome-agent is up to date")
		return nil
	}

	if err := updateAgentCRDs(chartVersion); err != nil {
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

// deployedChartVersion returns the chart version of the live release, or ""
// when there is no release yet.
func deployedChartVersion() string {
	out, err := outputQuiet("helm", "get", "metadata", chartRelease, "-n", chartNamespace, "-o", "json")
	if err != nil {
		return ""
	}
	var meta struct {
		Version string `json:"version"`
	}
	if json.Unmarshal([]byte(out), &meta) != nil {
		return ""
	}
	return meta.Version
}

// helm never upgrades crds/ on an existing release -- apply them ourselves.
func updateAgentCRDs(chartVersion string) error {
	dir, err := os.MkdirTemp("", "stackdome-chart")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	stepLog("Updating stackdome-agent CRDs...")
	if err := run("helm", "pull", chartRepo,
		"--version", chartVersion,
		"--untar", "--untardir", dir,
	); err != nil {
		return fmt.Errorf("pulling stackdome-agent chart: %w", err)
	}

	// The stackdome CRDs live in the stackdome-agent-standalone subchart, not the
	// umbrella chart, and traefik/cert-manager/cnpg each ship their own.
	crdDirs, err := findCRDDirs(filepath.Join(dir, chartRelease))
	if err != nil {
		return fmt.Errorf("scanning chart for CRDs: %w", err)
	}
	if len(crdDirs) == 0 {
		stepLog("Chart ships no CRDs -- skipping")
		return nil
	}

	for _, crdDir := range crdDirs {
		if err := run("kubectl", "apply", "--server-side", "--force-conflicts", "-f", crdDir); err != nil {
			return fmt.Errorf("applying CRDs from %s: %w", crdDir, err)
		}
	}
	return nil
}

func findCRDDirs(chartRoot string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(chartRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		// Only a chart root's crds/ is skipped by helm on upgrade. A
		// templates/crds/ is an ordinary template the release already owns, and
		// its unrendered {{ }} would not parse here anyway.
		if entry.Name() == templatesDirName {
			return filepath.SkipDir
		}
		if entry.Name() == crdsDirName {
			dirs = append(dirs, path)
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dirs, nil
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
