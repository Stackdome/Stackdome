package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

const (
	validToken   = auth.ApiTokenPrefix + "deadbeef"
	tokenID      = "token-1"
	tokenOwnerID = userDev
)

func storedToken(scopes []string) *models.APIToken {
	return &models.APIToken{
		ID:     tokenID,
		UserID: tokenOwnerID,
		OrgID:  orgABC,
		Scopes: scopes,
	}
}

func tokenOwner() *models.User {
	return &models.User{ID: tokenOwnerID, Role: models.OrgMemberRole}
}

// decode returns the status and the "reason" field every auth error response carries.
func decode(rec *httptest.ResponseRecorder) (int, string) {
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return rec.Code, body.Reason
}

var _ = Describe("apiTokenAuthnHandler", func() {
	var (
		ctrl        *gomock.Controller
		tokenLookup *auth.MockTokenLookup
		userGetter  *auth.MockUserGetter
		handler     http.Handler
		gotIdentity *auth.Identity
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		tokenLookup = auth.NewMockTokenLookup(ctrl)
		userGetter = auth.NewMockUserGetter(ctrl)
		gotIdentity = nil

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotIdentity = auth.GetIdentityFromCtx(r.Context())
			w.WriteHeader(http.StatusOK)
		})
		handler = auth.NewAPITokenHandler(next, auth.ApiTokenAuthnHandlerSpec{
			TokenLookup: tokenLookup,
			UserGetter:  userGetter,
		})
	})

	do := func(authHeader string) (int, string) {
		req := httptest.NewRequest(http.MethodGet, "/stacks", nil)
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return decode(rec)
	}

	It("authenticates a valid token and resolves the identity", func() {
		tokenLookup.EXPECT().ValidateToken(gomock.Any(), validToken).
			Return(storedToken([]string{auth.ScopeFullAccess}), nil)
		userGetter.EXPECT().InternalGet(gomock.Any(), tokenOwnerID).Return(tokenOwner(), nil)

		status, _ := do("Bearer " + validToken)
		Expect(status).To(Equal(http.StatusOK))
		Expect(gotIdentity).ToNot(BeNil())
		Expect(gotIdentity.UserID).To(Equal(tokenOwnerID))
		Expect(gotIdentity.OrgID).To(Equal(orgABC))
		Expect(gotIdentity.TokenID).To(Equal(tokenID))
		Expect(gotIdentity.AuthMethod).To(Equal(auth.AuthMethodAPIToken))
		Expect(gotIdentity.TokenScopes).To(Equal([]string{auth.ScopeFullAccess}))
	})

	It("rejects a header with no scheme/value split", func() {
		status, _ := do(validToken)
		Expect(status).To(Equal(http.StatusUnauthorized))
		Expect(gotIdentity).To(BeNil())
	})

	It("rejects a missing header", func() {
		status, _ := do("")
		Expect(status).To(Equal(http.StatusUnauthorized))
		Expect(gotIdentity).To(BeNil())
	})

	DescribeTable("rejects a token the service refuses",
		func(serr *errors.ServiceError) {
			tokenLookup.EXPECT().ValidateToken(gomock.Any(), validToken).Return(nil, serr)

			status, _ := do("Bearer " + validToken)
			Expect(status).To(Equal(http.StatusUnauthorized))
			Expect(gotIdentity).To(BeNil())
		},
		Entry("unknown token", errors.NotFound("api token not found")),
		Entry("revoked token", errors.Unauthorized("token has been revoked")),
		Entry("expired token", errors.Unauthorized("token has expired")),
	)

	It("rejects a token whose owner no longer exists", func() {
		tokenLookup.EXPECT().ValidateToken(gomock.Any(), validToken).
			Return(storedToken([]string{auth.ScopeFullAccess}), nil)
		userGetter.EXPECT().InternalGet(gomock.Any(), tokenOwnerID).Return(nil, errors.NotFound("user not found"))

		status, _ := do("Bearer " + validToken)
		Expect(status).To(Equal(http.StatusUnauthorized))
		Expect(gotIdentity).To(BeNil())
	})

	It("recognises only Bearer sdm_ headers as API-token requests", func() {
		req := httptest.NewRequest(http.MethodGet, "/stacks", nil)
		Expect(auth.CanAuthenticateWithAPIToken(req)).To(BeFalse())

		req.Header.Set("Authorization", "Bearer some.jwt.value")
		Expect(auth.CanAuthenticateWithAPIToken(req)).To(BeFalse())

		req.Header.Set("Authorization", "bearer "+validToken)
		Expect(auth.CanAuthenticateWithAPIToken(req)).To(BeTrue())
	})
})

// The full request path: Authorization header → authn handler → permission
// service. Everything below the token lookup is real, so a scope that does not
// cover the action must come back as 403 and never reach the endpoint.
var _ = Describe("API token requests through the authn + authz chain", func() {
	var (
		tokenLookup *auth.MockTokenLookup
		handler     http.Handler
		endpointHit bool
	)

	newChain := func() {
		env := newTestEnv(GinkgoT(), defaultProjects())
		Expect(env.policyMgr.AddGroupingPolicy(tokenOwnerID, string(models.DeveloperRole), projectABC)).To(Succeed())
		Expect(env.policyMgr.AddGroupingPolicy(tokenOwnerID, string(models.OrgMemberRole), orgABC)).To(Succeed())

		tokenLookup = auth.NewMockTokenLookup(env.ctrl)
		userGetter := auth.NewMockUserGetter(env.ctrl)
		userGetter.EXPECT().InternalGet(gomock.Any(), tokenOwnerID).Return(tokenOwner(), nil).AnyTimes()

		endpointHit = false
		endpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serr := env.permService.Check(r.Context(), projectABC, auth.ResourceStacks, "", auth.ActionCreate)
			if serr != nil {
				w.WriteHeader(serr.HttpCode)
				_ = json.NewEncoder(w).Encode(serr.AsOpenapiError())
				return
			}
			endpointHit = true
			w.WriteHeader(http.StatusOK)
		})

		handler = auth.NewAPITokenHandler(endpoint, auth.ApiTokenAuthnHandlerSpec{
			TokenLookup: tokenLookup,
			UserGetter:  userGetter,
		})
	}

	call := func() (int, string) {
		req := httptest.NewRequest(http.MethodPost, "/api/stackdome/v1/stacks", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return decode(rec)
	}

	It("lets a token whose scope covers the action through", func() {
		newChain()
		tokenLookup.EXPECT().ValidateToken(gomock.Any(), validToken).
			Return(storedToken([]string{auth.ResourceStacks + ":" + auth.ActionCreate}), nil)

		status, _ := call()
		Expect(status).To(Equal(http.StatusOK))
		Expect(endpointHit).To(BeTrue())
	})

	It("returns 403 when the token scope does not cover the action", func() {
		newChain()
		tokenLookup.EXPECT().ValidateToken(gomock.Any(), validToken).
			Return(storedToken([]string{auth.ResourceStacks + ":" + auth.ActionRead}), nil)

		status, reason := call()
		Expect(status).To(Equal(http.StatusForbidden))
		Expect(reason).To(ContainSubstring("insufficient permissions"))
		Expect(endpointHit).To(BeFalse())
	})

	It("returns 401 for a revoked token, without reaching the endpoint", func() {
		newChain()
		tokenLookup.EXPECT().ValidateToken(gomock.Any(), validToken).
			Return(nil, errors.Unauthorized("token has been revoked"))

		status, _ := call()
		Expect(status).To(Equal(http.StatusUnauthorized))
		Expect(endpointHit).To(BeFalse())
	})
})
