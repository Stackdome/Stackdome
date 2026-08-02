package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Stackdome/stackdome/install"
	"github.com/Stackdome/stackdome/pkg/models"
)

const defaultAPIServerImage = "quay.io/stackdome/stackdome:latest"

func main() {
	args := os.Args[1:]
	command := "install"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		command, args = args[0], args[1:]
	}

	switch command {
	case "install":
		runInstall(args)
	case "upgrade":
		runUpgrade(args)
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  stackdome-install install --email you@company.com [--domain stackdome.example.com]")
	fmt.Println("                            [--image IMAGE] [--chart-version VERSION]")
	fmt.Println("  stackdome-install upgrade [--image IMAGE] [--chart-version VERSION]")
	fmt.Println()
	fmt.Println("upgrade reuses the email, domain and secrets of the existing install,")
	fmt.Println("and keeps the deployed image unless --image is given.")
}

func runInstall(args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	email := fs.String("email", "", "Admin user email (required)")
	domain := fs.String("domain", "", "Dashboard domain (default: stackdome.<PUBLIC_IP>.nip.io)")
	image := fs.String("image", defaultAPIServerImage, "API server container image")
	chartVersion := fs.String("chart-version", defaultChartVersion, "stackdome-agent Helm chart version")
	_ = fs.Parse(args)

	if *email == "" {
		usage()
		fmt.Println()
		fmt.Println("Error: --email is required")
		os.Exit(1)
	}

	banner("StackDome VPS Installer")

	preflight, err := runPreflight(*email, *domain)
	if err != nil {
		exitErr("Preflight failed", err)
	}

	if err := installK3s(); err != nil {
		exitErr("k3s installation failed", err)
	}

	if err := installStackdomeAgent(*chartVersion); err != nil {
		exitErr("stackdome-agent installation failed", err)
	}

	vals := install.TemplateValues{
		AdminEmail:     *email,
		Domain:         preflight.Domain,
		APIServerImage: *image,
		DBWorkloadType: string(models.WorkloadTypeStatefulService),
		TLSEnabled:     isTLSDomain(preflight.Domain),
	}

	secrets, err := loadOrCreateSecrets(&vals)
	if err != nil {
		exitErr("Secret generation failed", err)
	}

	if err := configureTLS(vals); err != nil {
		exitErr("TLS configuration failed", err)
	}

	if err := applyCRs(vals, preflight.Domain); err != nil {
		exitErr("CR application failed", err)
	}

	result, err := runAPIBootstrap(vals, secrets)
	if err != nil {
		exitErr("API bootstrap failed", err)
	}

	printSummary(preflight.Domain, *email, secrets.AdminPassword, externallyReachable, result)
}

func banner(title string) {
	fmt.Println()
	fmt.Println("================================================================")
	fmt.Printf("  %s\n", title)
	fmt.Println("================================================================")
	fmt.Println()
}

func exitErr(msg string, err error) {
	errLog(fmt.Sprintf("%s: %v", msg, err))
	os.Exit(1)
}

func printSummary(domain, email, password string, externallyReachable bool, result *bootstrapResult) {
	successLog("Installation complete!")
	fmt.Println()
	fmt.Println("================================================================")
	fmt.Println("  StackDome installed successfully!")
	fmt.Println("================================================================")
	fmt.Println()
	fmt.Printf("  URL:            https://%s\n", domain)
	fmt.Printf("  Email:          %s\n", email)
	fmt.Printf("  Password:       %s\n", password)
	fmt.Println()
	fmt.Println("  What's ready:")
	fmt.Printf("    - Organization \"Default\" with domain %s\n", domain)
	fmt.Println("    - Cluster \"local\" registered and connected")
	fmt.Println("    - In-cluster image registry (50Gi)")
	fmt.Println("    - PostgreSQL database (10Gi persistent volume)")
	fmt.Println()
	if !externallyReachable {
		fmt.Println("  WARNING: Dashboard is NOT reachable from the internet.")
		fmt.Println("  Ensure ports 80 and 443 are open in your firewall/security group.")
		fmt.Println()
	}
	fmt.Println("  Save these credentials -- they won't be shown again.")
	fmt.Println()
	fmt.Println("  export KUBECONFIG=/etc/rancher/k3s/k3s.yaml")
	fmt.Println()
	fmt.Println("================================================================")
}
