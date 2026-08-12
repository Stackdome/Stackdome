package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/signupprotection"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type recordingSignupService struct {
	calls int
}

func (s *recordingSignupService) Signup(
	context.Context,
	*models.User,
	string,
) (*openapi.UserSignupResponse, *apperrors.ServiceError) {
	s.calls++
	return &openapi.UserSignupResponse{}, nil
}

type recordingSignupProtection struct {
	calls   int
	attempt signupprotection.PasswordSignupAttempt
	err     *apperrors.ServiceError
}

func (p *recordingSignupProtection) Check(
	_ context.Context,
	attempt signupprotection.PasswordSignupAttempt,
) *apperrors.ServiceError {
	p.calls++
	p.attempt = attempt
	return p.err
}

type recordingClientIPResolver struct {
	calls    int
	clientIP netip.Addr
	err      error
}

func (r *recordingClientIPResolver) Resolve(*http.Request) (netip.Addr, error) {
	r.calls++
	return r.clientIP, r.err
}

var _ = Describe("Password signup protection handler", func() {
	var (
		signupService *recordingSignupService
		protection    *recordingSignupProtection
		resolver      *recordingClientIPResolver
	)

	BeforeEach(func() {
		signupService = &recordingSignupService{}
		protection = &recordingSignupProtection{}
		resolver = &recordingClientIPResolver{clientIP: netip.MustParseAddr("203.0.113.10")}
	})

	It("protects public password signup before calling the signup service", func() {
		response := submitPasswordSignup(newProtectedUserHandler(signupService, protection, resolver), `{
			"name":"Person",
			"email":"person@example.com",
			"password":"password123",
			"organisation":{"name":"Example"},
			"turnstile_token":"browser-token"
		}`)

		Expect(response.Code).To(Equal(http.StatusCreated))
		Expect(protection.calls).To(Equal(1))
		Expect(protection.attempt).To(Equal(signupprotection.PasswordSignupAttempt{
			ClientIP:       netip.MustParseAddr("203.0.113.10"),
			Email:          "person@example.com",
			TurnstileToken: "browser-token",
		}))
		Expect(signupService.calls).To(Equal(1))
	})

	It("does not call the signup service when protection rejects the request", func() {
		protection.err = apperrors.Forbidden("signup verification failed")

		response := submitPasswordSignup(newProtectedUserHandler(signupService, protection, resolver), `{
			"name":"Person",
			"email":"person@example.com",
			"password":"password123",
			"organisation":{"name":"Example"},
			"turnstile_token":"invalid-token"
		}`)

		Expect(response.Code).To(Equal(http.StatusForbidden))
		Expect(signupService.calls).To(BeZero())
	})

	It("fails closed when the configured client IP is unavailable", func() {
		resolver.err = signupprotection.ErrClientIPUnavailable

		response := submitPasswordSignup(newProtectedUserHandler(signupService, protection, resolver), `{
			"name":"Person",
			"email":"person@example.com",
			"password":"password123",
			"organisation":{"name":"Example"},
			"turnstile_token":"browser-token"
		}`)

		Expect(response.Code).To(Equal(http.StatusForbidden))
		Expect(protection.calls).To(BeZero())
		Expect(signupService.calls).To(BeZero())
	})

	It("rejects invite signup when it is disabled", func() {
		handler := newProtectedUserHandler(signupService, protection, resolver)
		response := submitPasswordSignup(handler, `{
			"name":"Invited Person",
			"email":"invited@example.com",
			"password":"password123",
			"invite_token":"invite-token"
		}`)

		Expect(response.Code).To(Equal(http.StatusForbidden))
		Expect(resolver.calls).To(BeZero())
		Expect(protection.calls).To(BeZero())
		Expect(signupService.calls).To(BeZero())
	})

	It("leaves invite signup unchanged when it is enabled", func() {
		handler := NewUserServiceHandler(UserServiceHandlerSpec{
			SignupService:            signupService,
			SignupProtectionEnabled:  true,
			InviteSignupEnabled:      true,
			PasswordSignupProtection: protection,
			SignupClientIPResolver:   resolver,
		})
		response := submitPasswordSignup(handler, `{
			"name":"Invited Person",
			"email":"invited@example.com",
			"password":"password123",
			"invite_token":"invite-token"
		}`)

		Expect(response.Code).To(Equal(http.StatusCreated))
		Expect(resolver.calls).To(BeZero())
		Expect(protection.calls).To(BeZero())
		Expect(signupService.calls).To(Equal(1))
	})

	It("leaves password signup unchanged when protection is disabled", func() {
		handler := NewUserServiceHandler(UserServiceHandlerSpec{
			SignupService:            signupService,
			SignupProtectionEnabled:  false,
			PasswordSignupProtection: protection,
			SignupClientIPResolver:   resolver,
		})
		response := submitPasswordSignup(handler, `{
			"name":"Person",
			"email":"person@example.com",
			"password":"password123",
			"organisation":{"name":"Example"}
		}`)

		Expect(response.Code).To(Equal(http.StatusCreated))
		Expect(resolver.calls).To(BeZero())
		Expect(protection.calls).To(BeZero())
		Expect(signupService.calls).To(Equal(1))
	})

	DescribeTable("requiring protected signup dependencies",
		func(spec UserServiceHandlerSpec, expectedPanic string) {
			Expect(func() { NewUserServiceHandler(spec) }).To(PanicWith(expectedPanic))
		},
		Entry(
			"without signup protection",
			UserServiceHandlerSpec{SignupProtectionEnabled: true},
			"password signup protection is required when signup protection is enabled",
		),
		Entry(
			"without a client IP resolver",
			UserServiceHandlerSpec{
				SignupProtectionEnabled:  true,
				PasswordSignupProtection: &recordingSignupProtection{},
			},
			"signup client IP resolver is required when signup protection is enabled",
		),
	)
})

func newProtectedUserHandler(
	signupService *recordingSignupService,
	protection *recordingSignupProtection,
	resolver *recordingClientIPResolver,
) *usersHandler {
	return NewUserServiceHandler(UserServiceHandlerSpec{
		SignupService:            signupService,
		SignupProtectionEnabled:  true,
		InviteSignupEnabled:      false,
		PasswordSignupProtection: protection,
		SignupClientIPResolver:   resolver,
	})
}

func submitPasswordSignup(handler *usersHandler, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/user-signup", strings.NewReader(body))
	response := httptest.NewRecorder()
	handler.Signup(response, request)
	return response
}
