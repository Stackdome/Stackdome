package main

import (
	"fmt"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/install"
)

func isTLSDomain(domain string) bool {
	nonTLSSuffixes := []string{".nip.io", ".sslip.io", ".local", ".localhost"}
	for _, suffix := range nonTLSSuffixes {
		if strings.HasSuffix(domain, suffix) {
			return false
		}
	}
	return domain != ""
}

func configureTLS(vals install.TemplateValues) error {
	phaseLog(6, "Configuring TLS...")

	if !vals.TLSEnabled {
		stepLog("Using wildcard/local domain -- Traefik default self-signed TLS")
		successLog("TLS configured (self-signed)")
		return nil
	}

	stepLog("Custom domain detected -- applying Let's Encrypt ClusterIssuer...")
	manifest, err := install.RenderManifest("cluster-issuer.yaml", vals)
	if err != nil {
		return fmt.Errorf("rendering ClusterIssuer: %w", err)
	}

	if err := kubectlApply(manifest); err != nil {
		return fmt.Errorf("applying ClusterIssuer: %w", err)
	}

	successLog("TLS configured (Let's Encrypt)")
	return nil
}
