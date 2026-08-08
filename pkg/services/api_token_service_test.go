package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
)

var _ = Describe("apiTokenService", func() {
	const (
		ownerID = "user-owner"
		otherID = "user-other"
		orgID   = "org-1"
		tokenID = "token-1"
	)

	var (
		ctrl    *gomock.Controller
		store   *mocks.MockAPITokenStore
		service *apiTokenService
		ownerCt context.Context
	)

	identityFor := func(userID string, role models.UserRole) *auth.Identity {
		return &auth.Identity{UserID: userID, OrgID: orgID, Role: string(role), AuthMethod: auth.AuthMethodJWT}
	}

	ctxFor := func(identity *auth.Identity) context.Context {
		return auth.SetIdentityInContext(context.Background(), identity)
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		store = mocks.NewMockAPITokenStore(ctrl)
		logger := mocks.NewMockLogger(ctrl)
		logger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		service = &apiTokenService{store: store, logger: logger}
		ownerCt = ctxFor(identityFor(ownerID, models.OrgMemberRole))
	})

	Describe("Create", func() {
		It("returns the raw token once and stores only its hash", func() {
			var stored *models.APIToken
			store.EXPECT().Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, token *models.APIToken) (*models.APIToken, *errors.ServiceError) {
					stored = token
					return token, nil
				})

			created, rawToken, serr := service.Create(ownerCt, "ci", []string{auth.ScopeFullAccess}, nil, nil)
			Expect(serr).To(BeNil())
			Expect(rawToken).To(HavePrefix(auth.ApiTokenPrefix))
			Expect(rawToken).To(HaveLen(len(auth.ApiTokenPrefix) + 2*auth.ApiTokenByteLen))

			wantHash := sha256.Sum256([]byte(rawToken))
			Expect(stored.TokenHash).To(Equal(hex.EncodeToString(wantHash[:])))
			Expect(stored.TokenHash).ToNot(ContainSubstring(rawToken))
			Expect(stored.TokenPrefix).To(Equal(rawToken[:8]))
			Expect(stored.UserID).To(Equal(ownerID))
			Expect(stored.OrgID).To(Equal(orgID))
			Expect(created).To(Equal(stored))
		})

		It("gives every token a different raw value", func() {
			store.EXPECT().Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, token *models.APIToken) (*models.APIToken, *errors.ServiceError) {
					return token, nil
				}).Times(2)

			_, first, serr := service.Create(ownerCt, "ci", nil, nil, nil)
			Expect(serr).To(BeNil())
			_, second, serr := service.Create(ownerCt, "ci", nil, nil, nil)
			Expect(serr).To(BeNil())
			Expect(first).ToNot(Equal(second))
		})

		It("rejects an unknown scope before touching the store", func() {
			_, rawToken, serr := service.Create(ownerCt, "ci", []string{"not-a-resource:read"}, nil, nil)
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorBadRequest))
			Expect(rawToken).To(BeEmpty())
		})

		It("rejects an unknown action on a known resource", func() {
			_, _, serr := service.Create(ownerCt, "ci", []string{auth.ResourceStacks + ":teleport"}, nil, nil)
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorBadRequest))
		})

		It("rejects an unauthenticated caller", func() {
			_, _, serr := service.Create(context.Background(), "ci", nil, nil, nil)
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorUnauthorized))
		})
	})

	Describe("ValidateToken", func() {
		const rawToken = auth.ApiTokenPrefix + "abc123"

		hashOf := func(raw string) string {
			sum := sha256.Sum256([]byte(raw))
			return hex.EncodeToString(sum[:])
		}

		// ValidateToken stamps last-used on a goroutine. Wait for it, or gomock
		// sees the call land after the controller has finished.
		expectLastUsedStamp := func() chan struct{} {
			done := make(chan struct{})
			store.EXPECT().UpdateLastUsed(gomock.Any(), tokenID).
				DoAndReturn(func(context.Context, string) *errors.ServiceError {
					close(done)
					return nil
				})
			return done
		}

		It("looks the token up by hash, never by the raw value", func() {
			store.EXPECT().GetByTokenHash(gomock.Any(), hashOf(rawToken)).
				Return(&models.APIToken{ID: tokenID, UserID: ownerID}, nil)
			stamped := expectLastUsedStamp()

			token, serr := service.ValidateToken(context.Background(), rawToken)
			Expect(serr).To(BeNil())
			Expect(token.ID).To(Equal(tokenID))
			Eventually(stamped).Should(BeClosed())
		})

		It("rejects a revoked token", func() {
			store.EXPECT().GetByTokenHash(gomock.Any(), hashOf(rawToken)).
				Return(&models.APIToken{ID: tokenID, RevokedAt: ptr.To(time.Now().UTC())}, nil)

			_, serr := service.ValidateToken(context.Background(), rawToken)
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorUnauthorized))
		})

		It("rejects an expired token", func() {
			store.EXPECT().GetByTokenHash(gomock.Any(), hashOf(rawToken)).
				Return(&models.APIToken{ID: tokenID, ExpiresAt: ptr.To(time.Now().UTC().Add(-time.Hour))}, nil)

			_, serr := service.ValidateToken(context.Background(), rawToken)
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorUnauthorized))
		})

		It("accepts a token whose expiry is still ahead", func() {
			store.EXPECT().GetByTokenHash(gomock.Any(), hashOf(rawToken)).
				Return(&models.APIToken{ID: tokenID, ExpiresAt: ptr.To(time.Now().UTC().Add(time.Hour))}, nil)
			stamped := expectLastUsedStamp()

			_, serr := service.ValidateToken(context.Background(), rawToken)
			Expect(serr).To(BeNil())
			Eventually(stamped).Should(BeClosed())
		})

		It("passes an unknown token's store error straight back", func() {
			store.EXPECT().GetByTokenHash(gomock.Any(), gomock.Any()).Return(nil, errors.NotFound("api token not found"))

			_, serr := service.ValidateToken(context.Background(), strings.ToUpper(rawToken))
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorNotFound))
		})
	})

	Describe("ownership", func() {
		othersToken := &models.APIToken{ID: tokenID, UserID: otherID, OrgID: orgID}

		It("lets the owner read their own token", func() {
			store.EXPECT().GetByID(gomock.Any(), tokenID).Return(&models.APIToken{ID: tokenID, UserID: ownerID}, nil)

			token, serr := service.GetByID(ownerCt, tokenID)
			Expect(serr).To(BeNil())
			Expect(token.ID).To(Equal(tokenID))
		})

		It("refuses to read another user's token", func() {
			store.EXPECT().GetByID(gomock.Any(), tokenID).Return(othersToken, nil)

			_, serr := service.GetByID(ownerCt, tokenID)
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorForbidden))
		})

		It("refuses to revoke another user's token", func() {
			store.EXPECT().GetByID(gomock.Any(), tokenID).Return(othersToken, nil)

			serr := service.Revoke(ownerCt, tokenID)
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorForbidden))
		})

		It("lets an org admin revoke another user's token", func() {
			store.EXPECT().GetByID(gomock.Any(), tokenID).Return(othersToken, nil)
			store.EXPECT().Revoke(gomock.Any(), tokenID).Return(nil)

			adminCtx := ctxFor(identityFor(ownerID, models.OrgAdminRole))
			Expect(service.Revoke(adminCtx, tokenID)).To(BeNil())
		})

		It("lists only the caller's own tokens", func() {
			store.EXPECT().ListByUserID(gomock.Any(), ownerID).Return([]*models.APIToken{{ID: tokenID}}, nil)

			tokens, serr := service.List(ownerCt)
			Expect(serr).To(BeNil())
			Expect(tokens).To(HaveLen(1))
		})
	})
})
