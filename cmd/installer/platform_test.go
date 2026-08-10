package main

import (
	"flag"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/install"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func parsePlatformFlags(args ...string) (*platformFlags, error) {
	fs := flag.NewFlagSet("platform-test", flag.ContinueOnError)
	flags := registerPlatformFlags(fs)
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return flags, nil
}

var _ = Describe("Platform installer flags", func() {
	It("preserves stored TLS when the flag is omitted", func() {
		stored := install.PlatformConfig{
			BaseDomain:         "apps.example.com",
			TLSEnabled:         true,
			CloudflareAPIToken: "cloudflare-token",
			ACMEEnvironment:    config.ACMEEnvironmentStaging,
		}
		flags, err := parsePlatformFlags()
		Expect(err).NotTo(HaveOccurred())

		resolved, err := flags.resolvePlatformConfig(stored)
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved).To(Equal(stored))
	})

	It("allows shared compute routing without TLS", func() {
		flags, err := parsePlatformFlags("--platform-base-domain", "apps.example.com")
		Expect(err).NotTo(HaveOccurred())

		resolved, err := flags.resolvePlatformConfig(install.PlatformConfig{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Enabled()).To(BeTrue())
		Expect(resolved.TLSEnabled).To(BeFalse())
		Expect(resolved.CloudflareAPIToken).To(BeEmpty())
		Expect(resolved.ACMEEnvironment).To(BeEmpty())
	})

	It("requires a Cloudflare token when TLS is enabled", func() {
		flags, err := parsePlatformFlags(
			"--platform-base-domain", "apps.example.com",
			"--platform-tls",
		)
		Expect(err).NotTo(HaveOccurred())

		_, err = flags.resolvePlatformConfig(install.PlatformConfig{})
		Expect(err).To(MatchError(ContainSubstring("--platform-cloudflare-token-file is required")))
	})

	DescribeTable("rejects TLS-only flags when TLS is disabled",
		func(args ...string) {
			flags, err := parsePlatformFlags(args...)
			Expect(err).NotTo(HaveOccurred())

			_, err = flags.resolvePlatformConfig(install.PlatformConfig{})
			Expect(err).To(MatchError(ContainSubstring("require --platform-tls=true")))
		},
		Entry("Cloudflare token", "--platform-base-domain", "apps.example.com", "--platform-cloudflare-token-file", "token.txt"),
		Entry("ACME environment", "--platform-base-domain", "apps.example.com", "--platform-acme-environment", "staging"),
	)
})
