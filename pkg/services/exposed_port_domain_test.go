package services

import (
	"strings"
	"testing"

	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Platform domain ID", func() {
	It("is stable and distinct for each resource port", func() {
		first := platformDomainID("resource-a", 8080)

		Expect(first).To(Equal("aa982eec"))
		Expect(platformDomainID("resource-a", 8080)).To(Equal(first))
		Expect(platformDomainID("resource-b", 8080)).NotTo(Equal(first))
		Expect(platformDomainID("resource-a", 8081)).NotTo(Equal(first))
	})
})

var _ = Describe("Platform port FQDN", func() {
	It("uses the resource name without an explicit prefix", func() {
		fqdn, generatedID, err := FQDNForPortWithPlatformDomain(
			"resource-a", "My App", "platform.example.com", models.Port{Number: 8080},
		)

		Expect(err).NotTo(HaveOccurred())
		Expect(generatedID).To(Equal("aa982eec"))
		Expect(fqdn).To(Equal("my-app-aa982eec.platform.example.com"))
	})

	DescribeTable("uses an explicit or preview prefix",
		func(prefix, want string) {
			fqdn, generatedID, err := FQDNForPortWithPlatformDomain(
				"resource-a", "My App", "platform.example.com", models.Port{Number: 8080, SubdomainPrefix: prefix},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(generatedID).To(Equal("aa982eec"))
			Expect(fqdn).To(Equal(want))
		},
		Entry("explicit prefix", "api", "api-aa982eec.platform.example.com"),
		Entry("preview prefix", "pr-42-api", "pr-42-api-aa982eec.platform.example.com"),
	)

	It("truncates the readable head and retains the full ID", func() {
		fqdn, generatedID, err := FQDNForPortWithPlatformDomain(
			"resource-a", strings.Repeat("a", 60), "platform.example.com", models.Port{Number: 8080},
		)

		Expect(err).NotTo(HaveOccurred())
		firstLabel, baseDomain, found := strings.Cut(fqdn, ".")
		Expect(found).To(BeTrue())
		Expect(firstLabel).To(HaveLen(63))
		Expect(baseDomain).To(Equal("platform.example.com"))
		Expect(firstLabel).To(Equal(strings.Repeat("a", 54) + "-aa982eec"))
		Expect(generatedID).To(Equal("aa982eec"))
	})

	It("removes a hyphen at the truncation boundary", func() {
		readableHead := strings.Repeat("a", 53) + "-longer"

		fqdn, generatedID, err := FQDNForPortWithPlatformDomain(
			"resource-a", readableHead, "platform.example.com", models.Port{Number: 8080},
		)

		Expect(err).NotTo(HaveOccurred())
		firstLabel, _, found := strings.Cut(fqdn, ".")
		Expect(found).To(BeTrue())
		Expect(firstLabel).To(Equal(strings.Repeat("a", 53) + "-aa982eec"))
		Expect(generatedID).To(Equal("aa982eec"))
	})

	It("rejects an empty slug", func() {
		_, _, err := FQDNForPortWithPlatformDomain(
			"resource-a", "___", "platform.example.com", models.Port{Number: 8080},
		)

		Expect(err).To(HaveOccurred())
	})
})

func TestEncodeStackResourceSubdomainPrefix_stableForStackIDAndName(t *testing.T) {
	stackID := "stack-abc"
	name := "web"
	port := 8080

	first := EncodeStackResourceSubdomainPrefix(stackID, name, port)
	second := EncodeStackResourceSubdomainPrefix(stackID, name, port)

	assert.Equal(t, first, second)
	assert.Len(t, first, 16)
	assert.NotEqual(t,
		EncodeStackResourceSubdomainPrefix(stackID, name, 9090),
		EncodeStackResourceSubdomainPrefix(stackID, "api", port),
	)
}

var _ = Describe("Custom port FQDN", func() {
	It("preserves the legacy generated hostname", func() {
		fqdn, generatedPrefix := FQDNForPortWithCustomDomain(
			"stack-1", "web", "example.com", models.Port{Number: 8080},
		)

		Expect(fqdn).To(Equal("ek2dzkf3b56u664x.web.example.com"))
		Expect(generatedPrefix).To(Equal("ek2dzkf3b56u664x"))
	})

	It("uses the explicit prefix without generating one", func() {
		fqdn, generatedPrefix := FQDNForPortWithCustomDomain(
			"stack-1", "web", "example.com", models.Port{Number: 443, SubdomainPrefix: "api"},
		)

		Expect(fqdn).To(Equal("api.example.com"))
		Expect(generatedPrefix).To(BeEmpty())
	})
})
