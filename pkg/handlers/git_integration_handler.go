package handlers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

const (
	githubEventHeader      = "X-GitHub-Event"
	githubSignatureHeader  = "X-Hub-Signature-256"
	maxWebhookPayloadBytes = 2 << 20
)

type GitIntegrationHandlerSpec struct {
	GitIntegrationService services.GitIntegrationService
	GitHubWebhookService  services.GitHubWebhookService
	Logger                logger.Logger
}

type gitIntegrationHandler struct {
	gitIntegrationService services.GitIntegrationService
	githubWebhookService  services.GitHubWebhookService
	logger                logger.Logger
}

func NewGitIntegrationHandler(spec GitIntegrationHandlerSpec) *gitIntegrationHandler {
	return &gitIntegrationHandler{
		gitIntegrationService: spec.GitIntegrationService,
		githubWebhookService:  spec.GitHubWebhookService,
		logger:                spec.Logger,
	}
}

func (h *gitIntegrationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var integration openapi.GitIntegration
	cfg := &handlerConfig{
		MarshalInto: &integration,
		Action: func() (interface{}, *errors.ServiceError) {
			convertedObject := presenters.ConvertGitIntegration(&integration)
			convertedObject.OrganisationID = mux.Vars(r)["org_id"]

			obj, serr := h.gitIntegrationService.Create(r.Context(), convertedObject)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentGitIntegration(obj), nil
		},
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *gitIntegrationHandler) ListByOrgID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			objs, serr := h.gitIntegrationService.ListByOrgID(r.Context(), orgID)
			if serr != nil {
				return nil, serr
			}
			return openapi.GitIntegrationList{
				Items: presenters.PresentGitIntegrationList(objs),
				Total: ptr.To(int32(len(objs))),
			}, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *gitIntegrationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			obj, serr := h.gitIntegrationService.GetByID(r.Context(), id)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentGitIntegration(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *gitIntegrationHandler) Update(w http.ResponseWriter, r *http.Request) {
	var integration openapi.GitIntegration
	cfg := &handlerConfig{
		MarshalInto: &integration,
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			convertedObject := presenters.ConvertGitIntegration(&integration)

			obj, serr := h.gitIntegrationService.Update(r.Context(), id, convertedObject)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentGitIntegration(obj), nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *gitIntegrationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			if serr := h.gitIntegrationService.Delete(r.Context(), id); serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}

func (h *gitIntegrationHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var request openapi.GitIntegrationVerifyRequest
	cfg := &handlerConfig{
		MarshalInto: &request,
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			if serr := h.gitIntegrationService.Verify(r.Context(), id, request.GetRepoUrl()); serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handle(w, r, cfg, http.StatusNoContent)
}

func (h *gitIntegrationHandler) CreateGitHubAppManifest(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			flow, serr := h.gitIntegrationService.CreateGitHubAppManifest(r.Context(), orgID)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentGitHubAppManifestFlow(flow), nil
		},
	}
	handle(w, r, cfg, http.StatusCreated)
}

// GitHubManifestCallback is unauthenticated: the browser lands here from
// GitHub carrying the temporary code; the single-use state proves the flow
// was initiated by an org admin.
func (h *gitIntegrationHandler) GitHubManifestCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	redirectURL, serr := h.gitIntegrationService.HandleGitHubManifestCallback(r.Context(), code, state)
	if serr != nil {
		handleError(r.Context(), w, serr)
		return
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// GitHubAppSetup is unauthenticated: the browser lands here from GitHub after
// installing the platform app, carrying the new installation id; the
// single-use state proves the flow was initiated by an org admin.
func (h *gitIntegrationHandler) GitHubAppSetup(w http.ResponseWriter, r *http.Request) {
	installationID, err := strconv.ParseInt(r.URL.Query().Get("installation_id"), 10, 64)
	if err != nil {
		handleError(r.Context(), w, errors.MalformedRequest("installation_id must be a number"))
		return
	}

	redirectURL, serr := h.gitIntegrationService.HandleGitHubAppSetup(r.Context(), installationID, r.URL.Query().Get("state"))
	if serr != nil {
		handleError(r.Context(), w, serr)
		return
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// GitHubWebhook is unauthenticated at the HTTP layer; deliveries are
// HMAC-verified against the matched integration's webhook secret.
func (h *gitIntegrationHandler) GitHubWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookPayloadBytes))
	if err != nil {
		handleError(r.Context(), w, errors.MalformedRequest("failed to read webhook payload: %s", err.Error()))
		return
	}

	event := r.Header.Get(githubEventHeader)
	signature := r.Header.Get(githubSignatureHeader)

	if serr := h.githubWebhookService.ProcessGitHubWebhook(r.Context(), event, payload, signature); serr != nil {
		handleError(r.Context(), w, serr)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *gitIntegrationHandler) ListInstallations(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			refresh, _ := strconv.ParseBool(r.URL.Query().Get("refresh"))
			installations, serr := h.gitIntegrationService.ListInstallations(r.Context(), id, refresh)
			if serr != nil {
				return nil, serr
			}
			return openapi.GitInstallationList{
				Items: presenters.PresentGitInstallationList(installations),
				Total: ptr.To(int32(len(installations))),
			}, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *gitIntegrationHandler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			page, _ := strconv.Atoi(r.URL.Query().Get("page"))
			installationID := r.URL.Query().Get("installation_id")

			repoPage, serr := h.gitIntegrationService.ListRepositories(r.Context(), id, page, installationID)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentGitRepositoryPage(repoPage), nil
		},
	}
	handleList(w, r, cfg)
}

func (h *gitIntegrationHandler) GetRepository(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			vars := mux.Vars(r)
			repository, serr := h.gitIntegrationService.GetRepository(r.Context(), vars["id"], vars["owner"], vars["repo"])
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentGitRepository(*repository), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *gitIntegrationHandler) ListRepositoryBranches(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			vars := mux.Vars(r)
			branches, serr := h.gitIntegrationService.ListRepositoryBranches(r.Context(), vars["id"], vars["owner"], vars["repo"])
			if serr != nil {
				return nil, serr
			}
			return openapi.GitBranchList{
				Items: branches,
				Total: ptr.To(int32(len(branches))),
			}, nil
		},
	}
	handleList(w, r, cfg)
}
