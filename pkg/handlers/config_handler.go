package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
)

type ConfigHandlerSpec struct {
	GitHubOAuthEnabled     bool
	SignupTurnstileEnabled bool
	SignupTurnstileSiteKey string
	SignupTurnstileAction  string
}

type configHandler struct {
	githubOAuthEnabled     bool
	signupTurnstileEnabled bool
	signupTurnstileSiteKey string
	signupTurnstileAction  string
}

func NewConfigHandler(spec ConfigHandlerSpec) *configHandler {
	return &configHandler{
		githubOAuthEnabled:     spec.GitHubOAuthEnabled,
		signupTurnstileEnabled: spec.SignupTurnstileEnabled,
		signupTurnstileSiteKey: spec.SignupTurnstileSiteKey,
		signupTurnstileAction:  spec.SignupTurnstileAction,
	}
}

func (h *configHandler) Get(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (any, *errors.ServiceError) {
			return openapi.AppConfigResponse{
				GithubOauth: openapi.PtrBool(h.githubOAuthEnabled),
				Signup: &openapi.SignupConfigResponse{
					Turnstile: openapi.TurnstileConfigResponse{
						Enabled: h.signupTurnstileEnabled,
						SiteKey: h.signupTurnstileSiteKey,
						Action:  h.signupTurnstileAction,
					},
				},
			}, nil
		},
	}
	handleGet(w, r, cfg)
}
