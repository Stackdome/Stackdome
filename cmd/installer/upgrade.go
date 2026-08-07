package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Stackdome/stackdome/install"
)

func runUpgrade(args []string) {
	fs := flag.NewFlagSet("upgrade", flag.ExitOnError)
	image := fs.String("image", "", "API server container image (default: keep the currently deployed one)")
	chartVersion := fs.String("chart-version", defaultChartVersion, "stackdome-agent Helm chart version")
	github := registerGitHubFlags(fs)
	platform := registerPlatformFlags(fs)
	_ = fs.Parse(args)

	totalPhases = 4
	banner("StackDome Upgrade")

	phaseLog(1, "Checking existing install...")
	if err := requireRoot(); err != nil {
		exitErr("Preflight failed", err)
	}
	if err := requireExistingInstall(); err != nil {
		errLog(fmt.Sprintf("Nothing to upgrade: %v", err))
		fmt.Println()
		fmt.Println("Run the installer first:")
		fmt.Println("  sudo ./stackdome-install install --email you@company.com")
		os.Exit(1)
	}

	domain, err := existingDomain()
	if err != nil {
		exitErr("Reading existing configuration failed", err)
	}
	stepLog(fmt.Sprintf("Domain: %s", domain))

	if *image == "" {
		current, err := existingImage()
		if err != nil {
			exitErr("Reading existing configuration failed", err)
		}
		*image = current
		stepLog(fmt.Sprintf("Image: %s (unchanged)", current))
	} else {
		stepLog(fmt.Sprintf("Image: %s", *image))
	}

	phaseLog(2, "Loading existing secrets...")
	secrets, err := readExistingSecrets()
	if err != nil {
		exitErr("Reading bootstrap secrets failed", err)
	}
	successLog("Secrets loaded")

	// An existing db keeps the workload type it was installed with -- switching a
	// live Postgres between Deployment and StatefulSet is not this command's job.
	dbWorkloadType, err := existingDBWorkloadType()
	if err != nil {
		exitErr("Reading existing configuration failed", err)
	}
	stepLog(fmt.Sprintf("Database workload type: %s", dbWorkloadType))

	vals := install.TemplateValues{
		Domain:         domain,
		APIServerImage: *image,
		DBWorkloadType: dbWorkloadType,
		TLSEnabled:     isTLSDomain(domain),
		DBPassword:     secrets.DBPassword,
		JWTSecret:      secrets.JWTSecret,
		EncryptionKey:  secrets.EncryptionKey,
		AdminPassword:  secrets.AdminPassword,
	}
	vals.Platform, err = platform.resolvePlatformConfig(secrets.Platform)
	if err != nil {
		exitErr("Reading platform configuration failed", err)
	}
	if err := github.applyTo(&vals); err != nil {
		exitErr("Reading GitHub credentials failed", err)
	}

	if err := installStackdomeAgent(*chartVersion); err != nil {
		exitErr("stackdome-agent upgrade failed", err)
	}
	if err := applyAPIServerRBAC(); err != nil {
		exitErr("API server RBAC upgrade failed", err)
	}
	if err := ensurePlatformClusterCredentials(&vals.Platform); err != nil {
		exitErr("Reading platform cluster credentials failed", err)
	}
	if err := mergeBootstrapConfig(&vals, secrets); err != nil {
		exitErr("Storing bootstrap configuration failed", err)
	}

	if err := applyCRs(vals, domain); err != nil {
		exitErr("CR application failed", err)
	}

	successLog("Upgrade complete!")
	fmt.Println()
	fmt.Printf("  URL:            https://%s\n", domain)
	fmt.Printf("  API server:     %s\n", *image)
	fmt.Printf("  Agent chart:    v%s\n", *chartVersion)
	fmt.Println()
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
