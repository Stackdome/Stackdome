package environment

import (
	"context"
	"net/http/httptest"
	"time"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/signupprotection"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Signup protection wiring", func() {
	It("uses explicit disabled protection in self-hosted mode", func() {
		environment := newEnvironment(testSpec)

		Expect(environment.initializeSignupProtection(context.Background())).To(Succeed())
		Expect(environment.PasswordSignupProtection).NotTo(BeNil())
		Expect(environment.SignupClientIPResolver).To(BeNil())
		Expect(environment.PasswordSignupProtection.Check(
			context.Background(),
			signupprotection.PasswordSignupAttempt{},
		)).To(BeNil())
	})

	It("uses the configured Cloudflare client IP source in cloud mode", func() {
		environment := newEnvironment(testSpec)
		environment.Config.RuntimeMode = config.RuntimeModeStackdomeCloud
		environment.Config.TurnstileSecret = "test-secret"
		environment.Config.StackdomeCloud = cloudConfigForSignup(config.StackdomeCloudClientIPSourceCloudflare)

		Expect(environment.initializeSignupProtection(context.Background())).To(Succeed())
		request := httptest.NewRequest("POST", "/api/v1/user-signup", nil)
		request.Header.Set("CF-Connecting-IP", "203.0.113.10")
		clientIP, err := environment.SignupClientIPResolver.Resolve(request)

		Expect(err).NotTo(HaveOccurred())
		Expect(clientIP.String()).To(Equal("203.0.113.10"))
	})
})

func cloudConfigForSignup(clientIPSource config.StackdomeCloudClientIPSource) *config.StackdomeCloudConfig {
	return &config.StackdomeCloudConfig{
		Signup: config.StackdomeCloudSignupConfig{
			ClientIPSource: clientIPSource,
			Turnstile: config.StackdomeCloudTurnstileConfig{
				Enabled:             true,
				SiteKey:             "public-site-key",
				ExpectedHostname:    "stackdome.com",
				ExpectedAction:      "turnstile-spin-v2",
				VerificationTimeout: config.ConfigDuration(5 * time.Second),
			},
			Throttle: config.StackdomeCloudThrottleConfig{
				IP: config.StackdomeCloudIPThrottleConfig{
					MaxTrackedClients: 10,
					MaxAttempts:       2,
					Window:            config.ConfigDuration(time.Minute),
				},
				Email: config.StackdomeCloudEmailThrottleConfig{
					MaxTrackedAddresses: 10,
					MaxAttempts:         2,
					Window:              config.ConfigDuration(time.Minute),
				},
			},
		},
	}
}
