package signupprotection

import (
	"context"
	"errors"
	"net/netip"
	"time"

	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type verifierFunc func(context.Context, string, netip.Addr) error

func (verify verifierFunc) Verify(ctx context.Context, token string, clientIP netip.Addr) error {
	return verify(ctx, token, clientIP)
}

var _ = Describe("Password signup protection", func() {
	var throttle ThrottleSpec

	BeforeEach(func() {
		throttle = ThrottleSpec{
			MaxTrackedKeys: 10,
			MaxAttempts:    1,
			Window:         time.Minute,
			Now:            time.Now,
		}
	})

	It("checks IP throttling before Turnstile", func() {
		verificationCalls := 0
		protector := newProtector(verifierFunc(func(context.Context, string, netip.Addr) error {
			verificationCalls++
			return nil
		}), throttle)
		clientIP := netip.MustParseAddr("203.0.113.10")

		Expect(protector.Check(context.Background(), PasswordSignupAttempt{
			ClientIP: clientIP, Email: "first@example.com", TurnstileToken: "valid",
		})).To(BeNil())
		serviceErr := protector.Check(context.Background(), PasswordSignupAttempt{
			ClientIP: clientIP, Email: "second@example.com", TurnstileToken: "valid",
		})

		Expect(serviceErr.Code).To(Equal(apperrors.ErrorTooManyRequests))
		Expect(verificationCalls).To(Equal(1))
	})

	It("does not consume the email limit when Turnstile rejects the request", func() {
		protector := newProtector(verifierFunc(func(_ context.Context, token string, _ netip.Addr) error {
			if token == "invalid" {
				return errors.New("rejected")
			}
			return nil
		}), throttle)

		serviceErr := protector.Check(context.Background(), PasswordSignupAttempt{
			ClientIP:       netip.MustParseAddr("203.0.113.10"),
			Email:          "person@example.com",
			TurnstileToken: "invalid",
		})
		Expect(serviceErr.Code).To(Equal(apperrors.ErrorForbidden))

		Expect(protector.Check(context.Background(), PasswordSignupAttempt{
			ClientIP:       netip.MustParseAddr("203.0.113.11"),
			Email:          "person@example.com",
			TurnstileToken: "valid",
		})).To(BeNil())
	})

	It("applies the email limit after successful verification", func() {
		protector := newProtector(verifierFunc(func(context.Context, string, netip.Addr) error {
			return nil
		}), throttle)

		Expect(protector.Check(context.Background(), PasswordSignupAttempt{
			ClientIP:       netip.MustParseAddr("203.0.113.10"),
			Email:          " Person@Example.com ",
			TurnstileToken: "valid",
		})).To(BeNil())
		serviceErr := protector.Check(context.Background(), PasswordSignupAttempt{
			ClientIP:       netip.MustParseAddr("203.0.113.11"),
			Email:          "person@example.com",
			TurnstileToken: "valid",
		})

		Expect(serviceErr.Code).To(Equal(apperrors.ErrorTooManyRequests))
	})

	It("allows signup when protection is explicitly disabled", func() {
		protection := NewDisabledPasswordSignupProtection()

		Expect(protection.Check(context.Background(), PasswordSignupAttempt{})).To(BeNil())
	})
})

func newProtector(verifier verifierFunc, throttle ThrottleSpec) PasswordSignupProtection {
	protector, err := NewPasswordSignupProtection(PasswordSignupProtectionSpec{
		Verifier:      verifier,
		IPThrottle:    throttle,
		EmailThrottle: throttle,
	})
	Expect(err).NotTo(HaveOccurred())
	return protector
}
