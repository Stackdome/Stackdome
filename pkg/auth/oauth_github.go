package auth

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/golang-jwt/jwt"
	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
	githubOAuth "golang.org/x/oauth2/github"
)

const oauthStateExpiry = 10 * time.Minute

type oAuthUserService interface {
	InternalGetByEmail(ctx context.Context, email string) (*models.User, *errors.ServiceError)
	InternalCreateOAuthUser(ctx context.Context, email, name, githubID, avatarURL string) (*models.User, error)
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

	oauthState := &models.OAuthState{
		State:     state,
		Provider:  models.OAuthProviderGitHub,
		CreatedAt: time.Now().UTC(),
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

	ghClient := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(token)))

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

	user, userErr := h.oAuthUserService.InternalGetByEmail(ctx, email)
	if userErr != nil {
		name := ghUser.GetName()
		if name == "" {
			name = ghUser.GetLogin()
		}
		githubID := fmt.Sprintf("%d", ghUser.GetID())

		newUser, createErr := h.oAuthUserService.InternalCreateOAuthUser(ctx, email, name, githubID, ghUser.GetAvatarURL())
		if createErr != nil {
			handleError(w, errors.ErrorGeneral, "failed to create user")
			return
		}
		user = newUser
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
