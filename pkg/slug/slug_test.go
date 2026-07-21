package slug_test

import (
	"regexp"
	"strings"

	"github.com/Stackdome/stackdome/pkg/slug"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var dns1123Label = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

var _ = Describe("slug.FromOrgName", func() {
	DescribeTable("normalises org names into DNS-1123 labels",
		func(input, expected string) {
			Expect(slug.FromOrgName(input)).To(Equal(expected))
		},
		Entry("lowercases", "Acme", "acme"),
		Entry("replaces punctuation with hyphen", "Acme, Inc.", "acme-inc"),
		Entry("collapses runs of non-alphanumeric", "a   b___c", "a-b-c"),
		Entry("trims leading and trailing hyphens", "--Acme--", "acme"),
		Entry("empty string falls back to org", "", "org"),
		Entry("symbols only fall back to org", "!!!", "org"),
		Entry("whitespace only falls back to org", "   ", "org"),
	)

	It("caps the slug at MaxSlugLength", func() {
		long := strings.Repeat("a", slug.MaxSlugLength+20)
		result := slug.FromOrgName(long)
		Expect(len(result)).To(Equal(slug.MaxSlugLength))
	})

	It("does not leave a trailing hyphen after capping", func() {
		result := slug.FromOrgName(strings.Repeat("ab-", 30))
		Expect(len(result)).To(BeNumerically("<=", slug.MaxSlugLength))
		Expect(result).ToNot(HaveSuffix("-"))
		Expect(result).To(MatchRegexp(dns1123Label.String()))
	})

	It("produces a valid DNS-1123 label", func() {
		Expect(slug.FromOrgName("My Org 2026!")).To(MatchRegexp(dns1123Label.String()))
	})

	It("leaves room for a registry suffix within 63 chars", func() {
		orgSlug := slug.FromOrgName(strings.Repeat("x", 100))
		registryName := orgSlug + "-" + strings.Repeat("a", 8)
		Expect(len(registryName)).To(BeNumerically("<=", 63))
	})
})

var _ = Describe("slug.RandomSuffix", func() {
	It("returns a 6-character lowercase hex string", func() {
		suffix := slug.RandomSuffix()
		Expect(suffix).To(HaveLen(6))
		Expect(suffix).To(MatchRegexp(`^[0-9a-f]{6}$`))
	})

	It("returns distinct values across calls", func() {
		Expect(slug.RandomSuffix()).ToNot(Equal(slug.RandomSuffix()))
	})
})
