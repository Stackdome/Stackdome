package resourceaccess

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestResourceAccess(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ResourceAccess Suite")
}

var _ = Describe("casbinResourceAccessPolicyManager grouping changes", func() {
	const (
		subject  = "user-1"
		role     = "TestRole"
		domain   = "project-1"
		resource = "stacks"
		action   = "read"
	)

	var mgr ResourceAccessPolicyManager

	BeforeEach(func() {
		var err error
		mgr, err = NewInMemoryPolicyManager()
		Expect(err).ToNot(HaveOccurred())
		Expect(mgr.AddPolicy(role, domain, resource, action)).To(Succeed())
	})

	It("clears a cached deny once the grouping is added", func() {
		allowed, err := mgr.CheckPermission(subject, domain, resource, action)
		Expect(err).ToNot(HaveOccurred())
		Expect(allowed).To(BeFalse())

		Expect(mgr.AddGroupingPolicy(subject, role, domain)).To(Succeed())

		allowed, err = mgr.CheckPermission(subject, domain, resource, action)
		Expect(err).ToNot(HaveOccurred())
		Expect(allowed).To(BeTrue())
	})

	It("clears a cached allow once the grouping is removed", func() {
		Expect(mgr.AddGroupingPolicy(subject, role, domain)).To(Succeed())
		allowed, err := mgr.CheckPermission(subject, domain, resource, action)
		Expect(err).ToNot(HaveOccurred())
		Expect(allowed).To(BeTrue())

		Expect(mgr.RemoveGroupingPolicy(subject, role, domain)).To(Succeed())

		allowed, err = mgr.CheckPermission(subject, domain, resource, action)
		Expect(err).ToNot(HaveOccurred())
		Expect(allowed).To(BeFalse())
	})
})
