package signupprotection

import (
	"context"
	"errors"
	"fmt"
	"net/netip"

	"github.com/Stackdome/stackdome/pkg/clients/turnstile"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
)

const (
	signupLimitReason        = "too many signup attempts"
	signupVerificationReason = "signup verification failed"
)

// PasswordSignupAttempt contains the untrusted inputs needed at the public
// password-signup request boundary.
type PasswordSignupAttempt struct {
	ClientIP       netip.Addr
	Email          string
	TurnstileToken string
}

// PasswordSignupProtection decides whether password signup may continue.
type PasswordSignupProtection interface {
	Check(context.Context, PasswordSignupAttempt) *apperrors.ServiceError
}

// PasswordSignupProtectionSpec contains the cloud verifier and bounded limits.
type PasswordSignupProtectionSpec struct {
	Verifier      turnstile.Verifier
	IPThrottle    ThrottleSpec
	EmailThrottle ThrottleSpec
}

// passwordSignupProtection verifies and throttles one public password signup
// before the existing signup service is called.
type passwordSignupProtection struct {
	verifier     turnstile.Verifier
	ipLimiter    *ipLimiter
	emailLimiter *emailLimiter
}

// NewPasswordSignupProtection constructs cloud password-signup protection.
func NewPasswordSignupProtection(spec PasswordSignupProtectionSpec) (PasswordSignupProtection, error) {
	if spec.Verifier == nil {
		return nil, errors.New("signup protection verifier is required")
	}

	ipLimiter, err := newIPLimiter(spec.IPThrottle)
	if err != nil {
		return nil, fmt.Errorf("invalid IP throttle: %w", err)
	}
	emailLimiter, err := newEmailLimiter(spec.EmailThrottle)
	if err != nil {
		return nil, fmt.Errorf("invalid email throttle: %w", err)
	}

	return &passwordSignupProtection{
		verifier:     spec.Verifier,
		ipLimiter:    ipLimiter,
		emailLimiter: emailLimiter,
	}, nil
}

func (p *passwordSignupProtection) Check(
	ctx context.Context,
	attempt PasswordSignupAttempt,
) *apperrors.ServiceError {
	if !attempt.ClientIP.IsValid() {
		return apperrors.Forbidden(signupVerificationReason)
	}

	clientIP := attempt.ClientIP.Unmap()
	if !p.ipLimiter.Allow(clientIP) {
		return apperrors.TooManyRequests(signupLimitReason)
	}
	if err := p.verifier.Verify(ctx, attempt.TurnstileToken, clientIP); err != nil {
		return apperrors.Forbidden(signupVerificationReason)
	}
	if !p.emailLimiter.Allow(attempt.Email) {
		return apperrors.TooManyRequests(signupLimitReason)
	}

	return nil
}

type disabledPasswordSignupProtection struct{}

// NewDisabledPasswordSignupProtection preserves self-hosted signup without
// cloud verification or throttling.
func NewDisabledPasswordSignupProtection() PasswordSignupProtection {
	return disabledPasswordSignupProtection{}
}

func (disabledPasswordSignupProtection) Check(context.Context, PasswordSignupAttempt) *apperrors.ServiceError {
	return nil
}

var _ PasswordSignupProtection = (*passwordSignupProtection)(nil)
var _ PasswordSignupProtection = disabledPasswordSignupProtection{}
