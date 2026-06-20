package handlers

import "net/http"

type ConfigHandlerSpec struct {
	GitHubOAuthEnabled bool
}

type configHandler struct {
	githubOAuthEnabled bool
}

func NewConfigHandler(spec ConfigHandlerSpec) *configHandler {
	return &configHandler{githubOAuthEnabled: spec.GitHubOAuthEnabled}
}

// appConfigResponse mirrors the AppConfigResponse schema in the OpenAPI spec.
type appConfigResponse struct {
	GithubOauth bool `json:"github_oauth"`
}

func (h *configHandler) Get(w http.ResponseWriter, r *http.Request) {
	writeJSONResponse(w, http.StatusOK, appConfigResponse{GithubOauth: h.githubOAuthEnabled})
}
