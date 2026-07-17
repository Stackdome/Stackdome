package auth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/golang-jwt/jwt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/Stackdome/stackdome/pkg/auth"
)

var _ = Describe("jwtAuthnHandler", func() {
	const secret = "test-secret"

	var (
		ctrl    *gomock.Controller
		handler http.Handler
		nextHit bool
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		nextHit = false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextHit = true
			w.WriteHeader(http.StatusOK)
		})
		// UserGetter is never reached on the token-error paths under test, but the
		// seam is wired with a mock (expecting no calls) rather than left nil.
		handler = auth.NewJwtAuthnHandler(next, auth.JWTAuthnHandlerSpec{
			JWTSecret:  []byte(secret),
			UserGetter: auth.NewMockUserGetter(ctrl),
		})
	})

	// sign builds a token with the given expiry, signed with the test secret.
	sign := func(exp time.Time) string {
		claims := &auth.Claims{UserID: "user-1"}
		claims.ExpiresAt = exp.Unix()
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := tok.SignedString([]byte(secret))
		Expect(err).NotTo(HaveOccurred())
		return signed
	}

	do := func(authHeader string) (int, string) {
		req := httptest.NewRequest(http.MethodGet, "/stacks", nil)
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		var body struct {
			Reason string `json:"reason"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		return rec.Code, body.Reason
	}

	It("reports an expired token with a refresh-triggering reason", func() {
		status, reason := do("Bearer " + sign(time.Now().Add(-time.Hour)))
		Expect(status).To(Equal(http.StatusForbidden))
		Expect(reason).To(Equal("token expired"))
		Expect(nextHit).To(BeFalse())
	})

	It("reports a tamper/garbage token as an invalid token (no refresh)", func() {
		status, reason := do("Bearer not-a-real-jwt")
		Expect(status).To(Equal(http.StatusForbidden))
		Expect(reason).To(Equal("Invalid token"))
		Expect(nextHit).To(BeFalse())
	})

	It("reports a token signed with the wrong secret as invalid, not expired", func() {
		claims := &auth.Claims{UserID: "user-1"}
		claims.ExpiresAt = time.Now().Add(time.Hour).Unix()
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := tok.SignedString([]byte("wrong-secret"))
		Expect(err).NotTo(HaveOccurred())

		status, reason := do("Bearer " + signed)
		Expect(status).To(Equal(http.StatusForbidden))
		Expect(reason).To(Equal("Invalid token"))
		Expect(nextHit).To(BeFalse())
	})

	It("rejects a missing Authorization header", func() {
		status, reason := do("")
		Expect(status).To(Equal(http.StatusForbidden))
		Expect(reason).To(Equal("Authorization header missing"))
		Expect(nextHit).To(BeFalse())
	})
})
