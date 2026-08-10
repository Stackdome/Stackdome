package main

import (
	"flag"
	"fmt"

	"github.com/Stackdome/stackdome/install"
)

type upgradeOptions struct {
	image        string
	chartVersion string
	github       *githubFlags
	platform     *platformFlags
}

func parseUpgradeOptions(args []string) (*upgradeOptions, error) {
	fs := flag.NewFlagSet("upgrade", flag.ContinueOnError)
	fs.SetOutput(installerOutput.stderr)
	image := fs.String("image", "", "API server container image (default: keep the currently deployed one)")
	chartVersion := fs.String("chart-version", defaultChartVersion, "stackdome-agent Helm chart version")
	output := registerOutputFlags(fs)
	github := registerGitHubFlags(fs)
	platform := registerPlatformFlags(fs)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if fs.NArg() != 0 {
		return nil, fmt.Errorf("unexpected positional arguments")
	}
	if err := output.apply(); err != nil {
		return nil, err
	}
	return &upgradeOptions{
		image:        *image,
		chartVersion: *chartVersion,
		github:       github,
		platform:     platform,
	}, nil
}

func runUpgrade(args []string) error {
	opts, err := parseUpgradeOptions(args)
	if err != nil {
		return installationError("arguments", "invalid installer arguments", err)
	}

	totalPhases = 4
	banner("StackDome Upgrade")

	phaseLog(1, "Checking existing install...")
	if err := requireRoot(); err != nil {
		return installationError("preflight", err.Error(), err)
	}
	if err := requireExistingInstall(); err != nil {
		installerOutput.diagnosticf("\nRun the installer first:\n")
		installerOutput.diagnosticf("  sudo ./stackdome-install install --email you@company.com\n")
		return installationError("preflight", "no existing installation found", err)
	}

	domain, err := existingDomain()
	if err != nil {
		return installationError("configuration", "reading existing configuration failed", err)
	}
	stepLog(fmt.Sprintf("Domain: %s", domain))

	if opts.image == "" {
		current, err := existingImage()
		if err != nil {
			return installationError("configuration", "reading existing configuration failed", err)
		}
		opts.image = current
		stepLog(fmt.Sprintf("Image: %s (unchanged)", current))
	} else {
		stepLog(fmt.Sprintf("Image: %s", opts.image))
	}

	phaseLog(2, "Loading existing secrets...")
	secrets, err := readExistingSecrets()
	if err != nil {
		return installationError("configuration", "reading bootstrap secrets failed", err)
	}
	successLog("Secrets loaded")

	// An existing db keeps the workload type it was installed with -- switching a
	// live Postgres between Deployment and StatefulSet is not this command's job.
	dbWorkloadType, err := existingDBWorkloadType()
	if err != nil {
		return installationError("configuration", "reading existing configuration failed", err)
	}
	stepLog(fmt.Sprintf("Database workload type: %s", dbWorkloadType))

	vals := install.TemplateValues{
		Domain:         domain,
		APIServerImage: opts.image,
		DBWorkloadType: dbWorkloadType,
		TLSEnabled:     isTLSDomain(domain),
		DBPassword:     secrets.DBPassword,
		JWTSecret:      secrets.JWTSecret,
		EncryptionKey:  secrets.EncryptionKey,
		AdminPassword:  secrets.AdminPassword,
	}
	vals.Platform, err = opts.platform.resolvePlatformConfig(secrets.Platform)
	if err != nil {
		return installationError("configuration", "reading platform configuration failed", err)
	}
	if err := opts.github.applyTo(&vals); err != nil {
		return installationError("configuration", "reading GitHub credentials failed", err)
	}

	if err := installStackdomeAgent(opts.chartVersion); err != nil {
		return installationError("agent", "stackdome-agent upgrade failed", err)
	}
	if err := applyAPIServerRBAC(); err != nil {
		return installationError("agent", "API server RBAC upgrade failed", err)
	}
	vals.SharedCompute = secrets.SharedCompute
	if err := ensureSharedComputeClusterCredentials(vals.Platform, &vals.SharedCompute); err != nil {
		return installationError("configuration", "reading shared compute cluster credentials failed", err)
	}
	if err := mergeBootstrapConfig(&vals, secrets); err != nil {
		return installationError("configuration", "storing bootstrap configuration failed", err)
	}

	if err := applyCRs(vals, domain); err != nil {
		return installationError("configuration", "custom resource application failed", err)
	}

	successLog("Upgrade complete!")
	if installerOutput.isJSON() {
		if err := emitJSON(upgradeSuccessResult{
			Status:            "upgraded",
			URL:               installationURL(domain),
			APIServerImage:    opts.image,
			AgentChartVersion: opts.chartVersion,
		}); err != nil {
			return installationError("output", "writing JSON result failed", err)
		}
		return nil
	}
	installerOutput.finalf("\n  URL:            %s\n", installationURL(domain))
	installerOutput.finalf("  API server:     %s\n", opts.image)
	installerOutput.finalf("  Agent chart:    v%s\n\n", opts.chartVersion)
	return nil
}

func requireExistingInstall() error {
	if !isK3sRunning() {
		return fmt.Errorf("k3s is not running on this host")
	}
	if err := setupKubeconfig(); err != nil {
		return fmt.Errorf("kubeconfig %s unusable: %w", k3sKubeconfigPath, err)
	}
	if !commandExists("helm") {
		return fmt.Errorf("helm is not installed")
	}
	if _, err := outputQuiet("helm", "status", chartRelease, "-n", chartNamespace); err != nil {
		return fmt.Errorf("helm release %q not found in namespace %s", chartRelease, chartNamespace)
	}
	if _, err := outputQuiet("kubectl", "get", "secret", bootstrapSecretName, "-n", bootstrapSecretNamespace); err != nil {
		return fmt.Errorf("bootstrap secret %q not found in namespace %s", bootstrapSecretName, bootstrapSecretNamespace)
	}
	if _, err := outputQuiet("kubectl", "get", "stackresource", apiServerResourceName, "-n", chartNamespace); err != nil {
		return fmt.Errorf("StackResource %q not found in namespace %s", apiServerResourceName, chartNamespace)
	}

	successLog("Existing install detected")
	return nil
}

func existingDBWorkloadType() (string, error) {
	workloadType, err := output("kubectl", "get", "stackresource", dbResourceName,
		"-n", chartNamespace,
		"-o", "jsonpath={.spec.workloadType}")
	if err != nil {
		return "", fmt.Errorf("reading db StackResource: %w", err)
	}
	if workloadType == "" {
		return "", fmt.Errorf("db StackResource has no workloadType")
	}
	return workloadType, nil
}

func existingImage() (string, error) {
	image, err := output("kubectl", "get", "stackresource", apiServerResourceName,
		"-n", chartNamespace,
		"-o", "jsonpath={.spec.imageSpec.image}")
	if err != nil {
		return "", fmt.Errorf("reading api-server StackResource: %w", err)
	}
	if image == "" {
		return "", fmt.Errorf("api-server StackResource has no image")
	}
	return image, nil
}

func existingDomain() (string, error) {
	domain, err := output("kubectl", "get", "stackresource", apiServerResourceName,
		"-n", chartNamespace,
		"-o", "jsonpath={.spec.ports[0].fqdn}")
	if err != nil {
		return "", fmt.Errorf("reading api-server StackResource: %w", err)
	}
	if domain == "" {
		return "", fmt.Errorf("api-server StackResource has no fqdn")
	}
	return domain, nil
}
