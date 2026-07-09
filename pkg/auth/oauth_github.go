package auth

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/golang-jwt/jwt"
	"github.com/google/go-github/v88/github"
	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"
)

const oauthStateExpiry = 10 * time.Minute

type oAuthUserService interface {
	InternalGetByEmail(ctx context.Context, email string) (*models.User, *errors.ServiceError)
	InternalCreateOAuthUser(ctx context.Context, email, name, githubID, avatarURL string) (*models.User, error)
	InternalCreateInvitedOAuthUser(ctx context.Context, email, name, githubID, avatarURL string, invite *models.OrgInvite) (*models.User, error)
}

type oAuthInviteService interface {
	ValidateAndConsume(ctx context.Context, rawToken, email string) (*models.OrgInvite, *errors.ServiceError)
}

type oAuthEncryptionService interface {
	EncryptData(in []byte) (string, error)
	DecryptData(in string) ([]byte, error)
}

type GitHubOAuthHandlerSpec struct {
	ClientID          string
	ClientSecret      string
	RedirectURI       string
	OAuthUserService  oAuthUserService
	OAuthStateStore   stores.OAuthStateStore
	RefreshTokenStore stores.RefreshTokenStore
	JWTSecret         []byte
	JWTClaimsBuilder  JWTClaimsBuilder
	OrgInviteService  oAuthInviteService
	EncryptionService oAuthEncryptionService
	Logger            logger.Logger
}

type gitHubOAuthHandler struct {
	clientID          string
	clientSecret      string
	redirectURI       string
	oAuthUserService  oAuthUserService
	oauthStateStore   stores.OAuthStateStore
	refreshTokenStore stores.RefreshTokenStore
	jwtSecret         []byte
	jwtClaimsBuilder  JWTClaimsBuilder
	orgInviteService  oAuthInviteService
	encryptionService oAuthEncryptionService
	logger            logger.Logger
}

func NewGitHubOAuthHandler(spec GitHubOAuthHandlerSpec) *gitHubOAuthHandler {
	return &gitHubOAuthHandler{
		clientID:          spec.ClientID,
		clientSecret:      spec.ClientSecret,
		redirectURI:       spec.RedirectURI,
		oAuthUserService:  spec.OAuthUserService,
		oauthStateStore:   spec.OAuthStateStore,
		refreshTokenStore: spec.RefreshTokenStore,
		jwtSecret:         spec.JWTSecret,
		jwtClaimsBuilder:  spec.JWTClaimsBuilder,
		orgInviteService:  spec.OrgInviteService,
		encryptionService: spec.EncryptionService,
		logger:            spec.Logger,
	}
}

func (h *gitHubOAuthHandler) oauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     h.clientID,
		ClientSecret: h.clientSecret,
		RedirectURL:  h.redirectURI,
		Scopes:       []string{"user:email"},
		Endpoint:     githubOAuth.Endpoint,
	}
}

func (h *gitHubOAuthHandler) HandleInitiate(w http.ResponseWriter, r *http.Request) {
	stateBytes := make([]byte, 16)
	if _, err := cryptorand.Read(stateBytes); err != nil {
		handleError(w, errors.ErrorGeneral, "failed to generate state")
		return
	}
	state := hex.EncodeToString(stateBytes)

	if err := h.oauthStateStore.DeleteExpired(r.Context(), oauthStateExpiry); err != nil {
		h.logger.Errorf("failed to cleanup expired oauth states: %s", err.Error())
	}

	inviteToken := r.URL.Query().Get("invite_token")
	encryptedInviteToken := ""
	if inviteToken != "" {
		encrypted, encErr := h.encryptionService.EncryptData([]byte(inviteToken))
		if encErr != nil {
			handleError(w, errors.ErrorGeneral, "failed to process invite token")
			return
		}
		encryptedInviteToken = encrypted
	}

	oauthState := &models.OAuthState{
		State:       state,
		Provider:    models.OAuthProviderGitHub,
		InviteToken: encryptedInviteToken,
		CreatedAt:   time.Now().UTC(),
	}
	if serr := h.oauthStateStore.Create(r.Context(), oauthState); serr != nil {
		handleError(w, errors.ErrorGeneral, "failed to store oauth state")
		return
	}

	url := h.oauthConfig().AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *gitHubOAuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if state == "" || code == "" {
		handleError(w, errors.ErrorBadRequest, "missing state or code")
		return
	}

	ctx := r.Context()

	storedState, serr := h.oauthStateStore.Consume(ctx, state, models.OAuthProviderGitHub)
	if serr != nil {
		handleError(w, errors.ErrorBadRequest, "invalid state parameter")
		return
	}
	if time.Since(storedState.CreatedAt) > oauthStateExpiry {
		handleError(w, errors.ErrorBadRequest, "state expired")
		return
	}

	token, err := h.oauthConfig().Exchange(ctx, code)
	if err != nil {
		handleError(w, errors.ErrorGeneral, "failed to exchange code")
		return
	}

	ghClient, err := github.NewClient(github.WithHTTPClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))))
	if err != nil {
		handleError(w, errors.ErrorGeneral, "failed to create github client")
		return
	}

	ghUser, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		handleError(w, errors.ErrorGeneral, "failed to fetch github user")
		return
	}

	email := ghUser.GetEmail()
	if email == "" {
		primaryEmail, err := h.fetchPrimaryEmail(ctx, ghClient)
		if err != nil {
			handleError(w, errors.ErrorGeneral, "failed to fetch github email")
			return
		}
		email = primaryEmail
	}

	if email == "" {
		handleError(w, errors.ErrorBadRequest, "github account has no verified email")
		return
	}

	name := ghUser.GetName()
	if name == "" {
		name = ghUser.GetLogin()
	}
	githubID := fmt.Sprintf("%d", ghUser.GetID())
	avatarURL := ghUser.GetAvatarURL()

	inviteToken := ""
	if storedState.InviteToken != "" {
		decrypted, decErr := h.encryptionService.DecryptData(storedState.InviteToken)
		if decErr != nil {
			handleError(w, errors.ErrorGeneral, "failed to process invite token")
			return
		}
		inviteToken = string(decrypted)
	}

	user, userErr := h.findOrCreateUser(ctx, email, name, githubID, avatarURL, inviteToken)
	if userErr != nil {
		handleError(w, userErr.Code, userErr.Reason)
		return
	}

	expirationTime := time.Now().UTC().Add(JwtTokenExpiry)
	claims := h.jwtClaimsBuilder.BuildClaims(user, expirationTime)
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, signErr := jwtToken.SignedString(h.jwtSecret)
	if signErr != nil {
		handleError(w, errors.ErrorGeneral, "failed to generate token")
		return
	}

	refreshToken, refreshErr := CreateRefreshToken(ctx, h.refreshTokenStore, user.ID)
	if refreshErr != nil {
		handleError(w, errors.ErrorGeneral, "failed to generate refresh token")
		return
	}

	resp := openapi.NewRefreshTokenResponse()
	resp.SetToken(tokenString)
	resp.SetRefreshToken(refreshToken)
	writeJSONResponse(w, http.StatusOK, resp)
}

func (h *gitHubOAuthHandler) findOrCreateUser(ctx context.Context, email, name, githubID, avatarURL, inviteToken string) (*models.User, *errors.ServiceError) {
	if inviteToken != "" {
		return h.createInvitedUser(ctx, email, name, githubID, avatarURL, inviteToken)
	}

	// Standard OAuth: login existing user or create new account
	existing, _ := h.oAuthUserService.InternalGetByEmail(ctx, email)
	if existing != nil {
		return existing, nil
	}

	created, err := h.oAuthUserService.InternalCreateOAuthUser(ctx, email, name, githubID, avatarURL)
	if err != nil {
		return nil, errors.InternalServerError("failed to create user")
	}
	return created, nil
}

func (h *gitHubOAuthHandler) createInvitedUser(ctx context.Context, email, name, githubID, avatarURL, inviteToken string) (*models.User, *errors.ServiceError) {
	existing, _ := h.oAuthUserService.InternalGetByEmail(ctx, email)
	if existing != nil {
		return nil, errors.Conflict("user with this email already exists")
	}

	invite, serr := h.orgInviteService.ValidateAndConsume(ctx, inviteToken, email)
	if serr != nil {
		return nil, serr
	}

	created, err := h.oAuthUserService.InternalCreateInvitedOAuthUser(ctx, email, name, githubID, avatarURL, invite)
	if err != nil {
		return nil, errors.InternalServerError("failed to create user")
	}
	return created, nil
}

func (h *gitHubOAuthHandler) fetchPrimaryEmail(ctx context.Context, client *github.Client) (string, error) {
	emails, _, err := client.Users.ListEmails(ctx, nil)
	if err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.GetPrimary() && e.GetVerified() {
			return e.GetEmail(), nil
		}
	}
	return "", fmt.Errorf("no verified primary email found")
}
