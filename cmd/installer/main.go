package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ashishmax31/stackdome-api-server/install"
)

const defaultAPIServerImage = "quay.io/stackdome/stackdome:latest"

func main() {
	email := flag.String("email", "", "Admin user email (required)")
	domain := flag.String("domain", "", "Dashboard domain (default: stackdome.<PUBLIC_IP>.nip.io)")
	image := flag.String("image", defaultAPIServerImage, "API server container image")
	flag.Parse()

	if *email == "" {
		fmt.Println("Usage: stackdome-install --email you@company.com [--domain stackdome.example.com]")
		fmt.Println()
		fmt.Println("Error: --email is required")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Println("  StackDome VPS Installer")
	fmt.Println("================================================================")
	fmt.Println()

	preflight, err := runPreflight(*email, *domain)
	if err != nil {
		errLog(fmt.Sprintf("Preflight failed: %v", err))
		os.Exit(1)
	}

	if err := installK3s(); err != nil {
		errLog(fmt.Sprintf("k3s installation failed: %v", err))
		os.Exit(1)
	}

	if err := installStackdomeAgent(); err != nil {
		errLog(fmt.Sprintf("stackdome-agent installation failed: %v", err))
		os.Exit(1)
	}

	vals := install.TemplateValues{
		AdminEmail:     *email,
		Domain:         preflight.Domain,
		APIServerImage: *image,
	}

	secrets, err := loadOrCreateSecrets(&vals)
	if err != nil {
		errLog(fmt.Sprintf("Secret generation failed: %v", err))
		os.Exit(1)
	}

	if err := applyCRs(vals, preflight.Domain); err != nil {
		errLog(fmt.Sprintf("CR application failed: %v", err))
		os.Exit(1)
	}

	result, err := runAPIBootstrap(vals, secrets)
	if err != nil {
		errLog(fmt.Sprintf("API bootstrap failed: %v", err))
		os.Exit(1)
	}

	if err := configureTLS(vals); err != nil {
		errLog(fmt.Sprintf("TLS configuration failed: %v", err))
		os.Exit(1)
	}

	printSummary(preflight.Domain, *email, secrets.AdminPassword, externallyReachable, result)
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
