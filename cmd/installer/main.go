package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Stackdome/stackdome/install"
	"github.com/Stackdome/stackdome/pkg/models"
)

const defaultAPIServerImage = "quay.io/stackdome/stackdome:latest"

var errEmailRequired = errors.New("--email is required")

type installOptions struct {
	email           string
	domain          string
	image           string
	chartVersion    string
	credentialsFile string
	github          *githubFlags
	platform        *platformFlags
}

func main() {
	os.Exit(runMain(os.Args[1:]))
}

func runMain(args []string) int {
	return runMainWithIO(args, os.Stdout, os.Stderr, stderrIsTerminal())
}

func runMainWithIO(args []string, stdout, stderr io.Writer, terminal bool) int {
	mode := string(outputHuman)
	if argumentsRequestJSON(args) {
		mode = string(outputJSON)
	}
	_ = configureOutput(mode, argumentsRequestNoColor(args), terminal, stdout, stderr)

	command := "install"
	commandArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		command, commandArgs = args[0], args[1:]
	}

	var err error
	switch command {
	case "install":
		err = runInstall(commandArgs)
	case "upgrade":
		err = runUpgrade(commandArgs)
	default:
		usage()
		err = installationError("arguments", "unknown installer command", fmt.Errorf("unknown command %q", command))
	}
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		emitFailure(err)
		return 1
	}
	return 0
}

func usage() {
	installerOutput.diagnosticf("Usage:\n")
	installerOutput.diagnosticf("  stackdome-install install --email you@company.com [--domain stackdome.example.com]\n")
	installerOutput.diagnosticf("                            [--image IMAGE] [--chart-version VERSION]\n")
	installerOutput.diagnosticf("                            [--output human|json] [--no-color]\n")
	installerOutput.diagnosticf("                            [--credentials-file PATH] [GitHub flags] [platform flags]\n")
	installerOutput.diagnosticf("  stackdome-install upgrade [--image IMAGE] [--chart-version VERSION]\n")
	installerOutput.diagnosticf("                              [--output human|json] [--no-color]\n")
	installerOutput.diagnosticf("                              [GitHub flags] [platform flags]\n\n")
	installerOutput.diagnosticf("GitHub flags (all optional, stored in the bootstrap secret and reused by\n")
	installerOutput.diagnosticf("later upgrades unless given again):\n")
	installerOutput.diagnosticf("  --github-client-id ID --github-client-secret SECRET   'Sign in with GitHub'\n")
	installerOutput.diagnosticf("  --github-app-id ID --github-app-slug SLUG             platform GitHub App\n")
	installerOutput.diagnosticf("  --github-app-key-file PATH --github-app-webhook-secret SECRET\n\n")
	installerOutput.diagnosticf("Platform flags (optional; configure wildcard hostnames and TLS):\n")
	installerOutput.diagnosticf("  --platform-base-domain DOMAIN\n")
	installerOutput.diagnosticf("  --platform-cloudflare-token-file PATH\n")
	installerOutput.diagnosticf("  --platform-acme-environment production|staging\n")
	installerOutput.diagnosticf("Only the Cloudflare token may be changed after initial configuration.\n\n")
	installerOutput.diagnosticf("upgrade reuses the email, domain and secrets of the existing install,\n")
	installerOutput.diagnosticf("and keeps the deployed image unless --image is given.\n")
}

func parseInstallOptions(args []string) (*installOptions, error) {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(installerOutput.stderr)
	email := fs.String("email", "", "Admin user email (required)")
	domain := fs.String("domain", "", "Dashboard domain (default: stackdome.<PUBLIC_IP>.nip.io)")
	image := fs.String("image", defaultAPIServerImage, "API server container image")
	chartVersion := fs.String("chart-version", defaultChartVersion, "stackdome-agent Helm chart version")
	credentialsFile := fs.String("credentials-file", "", "Secure file for generated admin credentials")
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
	if *email == "" {
		return nil, errEmailRequired
	}
	return &installOptions{
		email:           *email,
		domain:          *domain,
		image:           *image,
		chartVersion:    *chartVersion,
		credentialsFile: credentialsPath(*credentialsFile),
		github:          github,
		platform:        platform,
	}, nil
}

func runInstall(args []string) error {
	opts, err := parseInstallOptions(args)
	if err != nil {
		if errors.Is(err, errEmailRequired) {
			usage()
			return installationError("arguments", err.Error(), err)
		}
		return installationError("arguments", "invalid installer arguments", err)
	}
	totalPhases = 6
	banner("StackDome VPS Installer")

	preflight, err := runPreflight(opts.email, opts.domain)
	if err != nil {
		return installationError("preflight", err.Error(), err)
	}

	if err := installK3s(); err != nil {
		return installationError("k3s", "k3s installation failed", err)
	}

	if err := installStackdomeAgent(opts.chartVersion); err != nil {
		return installationError("agent", "stackdome-agent installation failed", err)
	}
	if err := applyAPIServerRBAC(); err != nil {
		return installationError("agent", "API server RBAC installation failed", err)
	}

	vals := install.TemplateValues{
		AdminEmail:     opts.email,
		Domain:         preflight.Domain,
		APIServerImage: opts.image,
		DBWorkloadType: string(models.WorkloadTypeStatefulService),
		TLSEnabled:     isTLSDomain(preflight.Domain),
	}
	storedPlatform := install.PlatformConfig{}
	if existingSecrets, readErr := readExistingSecrets(); readErr == nil {
		storedPlatform = existingSecrets.Platform
	}
	vals.Platform, err = opts.platform.resolvePlatformConfig(storedPlatform)
	if err != nil {
		return installationError("configuration", "reading platform configuration failed", err)
	}
	if err := ensurePlatformClusterCredentials(&vals.Platform); err != nil {
		return installationError("configuration", "reading platform cluster credentials failed", err)
	}
	if err := opts.github.applyTo(&vals); err != nil {
		return installationError("configuration", "reading GitHub credentials failed", err)
	}

	secrets, err := loadOrCreateSecrets(&vals)
	if err != nil {
		return installationError("configuration", "secret generation failed", err)
	}

	if err := configureTLS(vals); err != nil {
		return installationError("configuration", "TLS configuration failed", err)
	}

	if err := applyCRs(vals, preflight.Domain); err != nil {
		return installationError("configuration", "custom resource application failed", err)
	}

	result, err := runAPIBootstrap(vals, secrets)
	if err != nil {
		return installationError("bootstrap", "API bootstrap failed", err)
	}

	return finishInstall(preflight.Domain, opts.email, secrets.AdminPassword, opts.credentialsFile, externallyReachable, result)
}

func banner(title string) {
	installerOutput.diagnosticf("\n================================================================\n")
	installerOutput.diagnosticf("  %s\n", title)
	installerOutput.diagnosticf("================================================================\n\n")
}

func finishInstall(domain, email, password, credentialsFile string, reachable bool, result *bootstrapResult) error {
	url := "https://" + domain
	if credentialsFile != "" {
		credentials := bootstrapCredentials{URL: url, AdminEmail: email, AdminPassword: password}
		if err := writeCredentials(credentialsFile, credentials); err != nil {
			return installationError("configuration", "writing credentials file failed", err)
		}
	}

	successLog("Installation complete!")
	if installerOutput.isJSON() {
		response := installSuccessResult{
			Status:          "installed",
			URL:             url,
			AdminEmail:      email,
			CredentialsFile: credentialsFile,
			Organization:    namedResult{Name: "Default", ID: result.OrgID},
			Cluster:         namedResult{Name: "local", ID: result.ClusterID},
		}
		if err := emitJSON(response); err != nil {
			return installationError("output", "writing JSON result failed", err)
		}
		return nil
	}

	printSummary(domain, email, password, credentialsFile, reachable)
	return nil
}

func printSummary(domain, email, password, credentialsFile string, externallyReachable bool) {
	installerOutput.finalf("\n================================================================\n")
	installerOutput.finalf("  StackDome installed successfully!\n")
	installerOutput.finalf("================================================================\n\n")
	installerOutput.finalf("  URL:            https://%s\n", domain)
	installerOutput.finalf("  Email:          %s\n", email)
	installerOutput.finalf("  Password:       %s\n", password)
	if credentialsFile != "" {
		installerOutput.finalf("  Credentials:    %s\n", credentialsFile)
	}
	installerOutput.finalf("\n  What's ready:\n")
	installerOutput.finalf("    - Organization \"Default\" with domain %s\n", domain)
	installerOutput.finalf("    - Cluster \"local\" registered and connected\n")
	installerOutput.finalf("    - In-cluster image registry (50Gi)\n")
	installerOutput.finalf("    - PostgreSQL database (10Gi persistent volume)\n\n")
	if !externallyReachable {
		installerOutput.finalf("  WARNING: Dashboard is NOT reachable from the internet.\n")
		installerOutput.finalf("  Ensure ports 80 and 443 are open in your firewall/security group.\n\n")
	}
	installerOutput.finalf("  WARNING: Human-mode stdout contains the generated password.\n")
	installerOutput.finalf("  Do not redirect it to an insecure file.\n\n")
	installerOutput.finalf("  Save these credentials -- they won't be shown again.\n\n")
	installerOutput.finalf("  export KUBECONFIG=/etc/rancher/k3s/k3s.yaml\n\n")
	installerOutput.finalf("================================================================\n")
}
