package pgstore_test

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("APITokenStore", func() {
	var (
		store stores.APITokenStore
		ctx   context.Context
	)

	const userID = "user-1"

	newToken := func(id, name string) *models.APIToken {
		return &models.APIToken{
			ID:          id,
			Name:        name,
			UserID:      userID,
			TokenHash:   "hash-" + id,
			TokenPrefix: "sd_" + id,
			Scopes:      models.StringSlice{"read"},
			OrgID:       "org-1",
			CreatedAt:   time.Now().UTC(),
		}
	}

	BeforeEach(func() {
		sf := newSQLiteSessionFactory(`
			CREATE TABLE IF NOT EXISTS api_tokens (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				user_id TEXT NOT NULL,
				token_hash TEXT NOT NULL,
				token_prefix TEXT NOT NULL,
				scopes JSONB NOT NULL,
				resource_ids JSONB,
				org_id TEXT NOT NULL,
				expires_at DATETIME,
				last_used_at DATETIME,
				created_at DATETIME NOT NULL,
				revoked_at DATETIME
			)
		`)
		store = pgstore.NewAPITokenStore(pgstore.APITokenStoreSpec{SessionFactory: sf})
		ctx = context.Background()
	})

	It("omits revoked tokens from the list", func() {
		_, err := store.Create(ctx, newToken("token-live", "live"))
		Expect(err).To(BeNil())
		_, err = store.Create(ctx, newToken("token-revoked", "revoked"))
		Expect(err).To(BeNil())

		Expect(store.Revoke(ctx, "token-revoked")).To(BeNil())

		tokens, err := store.ListByUserID(ctx, userID)
		Expect(err).To(BeNil())
		Expect(tokens).To(HaveLen(1))
		Expect(tokens[0].ID).To(Equal("token-live"))
	})
})
