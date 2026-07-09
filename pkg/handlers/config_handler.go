package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
)

type ConfigHandlerSpec struct {
	GitHubOAuthEnabled bool
}

type configHandler struct {
	githubOAuthEnabled bool
}

func NewConfigHandler(spec ConfigHandlerSpec) *configHandler {
	return &configHandler{githubOAuthEnabled: spec.GitHubOAuthEnabled}
}

func (h *configHandler) Get(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (any, *errors.ServiceError) {
			return openapi.AppConfigResponse{GithubOauth: openapi.PtrBool(h.githubOAuthEnabled)}, nil
		},
	}
	handleGet(w, r, cfg)
}
