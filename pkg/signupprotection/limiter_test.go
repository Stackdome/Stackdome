package signupprotection

import (
	"net/netip"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Signup attempt throttles", func() {
	var (
		now  time.Time
		spec ThrottleSpec
	)

	BeforeEach(func() {
		now = time.Date(2026, time.August, 12, 12, 0, 0, 0, time.UTC)
		spec = ThrottleSpec{
			MaxTrackedKeys: 2,
			MaxAttempts:    2,
			Window:         time.Minute,
			Now:            func() time.Time { return now },
		}
	})

	It("allows the configured attempts and resets after the window", func() {
		limiter, err := newIPLimiter(spec)
		Expect(err).NotTo(HaveOccurred())
		clientIP := netip.MustParseAddr("203.0.113.10")

		Expect(limiter.Allow(clientIP)).To(BeTrue())
		Expect(limiter.Allow(clientIP)).To(BeTrue())
		Expect(limiter.Allow(clientIP)).To(BeFalse())

		now = now.Add(time.Minute)
		Expect(limiter.Allow(clientIP)).To(BeTrue())
	})

	It("rejects untracked keys while the bounded store is full", func() {
		spec.MaxTrackedKeys = 1
		limiter, err := newIPLimiter(spec)
		Expect(err).NotTo(HaveOccurred())

		Expect(limiter.Allow(netip.MustParseAddr("203.0.113.10"))).To(BeTrue())
		Expect(limiter.Allow(netip.MustParseAddr("203.0.113.11"))).To(BeFalse())

		now = now.Add(time.Minute)
		Expect(limiter.Allow(netip.MustParseAddr("203.0.113.11"))).To(BeTrue())
	})

	It("normalizes equivalent IP addresses", func() {
		spec.MaxAttempts = 1
		limiter, err := newIPLimiter(spec)
		Expect(err).NotTo(HaveOccurred())

		Expect(limiter.Allow(netip.MustParseAddr("::ffff:203.0.113.10"))).To(BeTrue())
		Expect(limiter.Allow(netip.MustParseAddr("203.0.113.10"))).To(BeFalse())
	})

	It("normalizes and hashes email addresses", func() {
		spec.MaxAttempts = 1
		limiter, err := newEmailLimiter(spec)
		Expect(err).NotTo(HaveOccurred())

		Expect(limiter.Allow(" Person@Example.COM ")).To(BeTrue())
		Expect(limiter.Allow("person@example.com")).To(BeFalse())
	})
})
