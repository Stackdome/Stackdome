package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

type ProjectHandlerSpec struct {
	ProjectService services.ProjectService
}

type projectHandler struct {
	projectService services.ProjectService
}

func NewProjectHandler(spec ProjectHandlerSpec) *projectHandler {
	return &projectHandler{
		projectService: spec.ProjectService,
	}
}

func (h *projectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.ProjectCreateRequest
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			project := &models.Project{Name: req.GetName()}
			created, serr := h.projectService.CreateProject(r.Context(), orgID, project)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentProject(created), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *projectHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			projects, serr := h.projectService.ListProjects(r.Context(), orgID)
			if serr != nil {
				return nil, serr
			}
			return openapi.ProjectList{
				Items: presenters.PresentProjectList(projects),
				Total: ptr.To(int32(len(projects))),
			}, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *projectHandler) ListCurrentUserProjects(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			identity := auth.GetIdentityFromCtx(r.Context())
			if identity == nil {
				return nil, errors.Unauthorized("user identity not found in context")
			}
			memberships, serr := h.projectService.ListUserProjects(r.Context(), identity.UserID)
			if serr != nil {
				return nil, serr
			}
			return openapi.ProjectMembershipList{
				Items: presenters.PresentProjectMembershipList(memberships),
				Total: ptr.To(int32(len(memberships))),
			}, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *projectHandler) GetByName(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			projectName := mux.Vars(r)["project_name"]
			project, serr := h.projectService.GetProjectByOrgAndName(r.Context(), orgID, projectName)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentProject(project), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *projectHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req openapi.ProjectUpdateRequest
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			projectName := mux.Vars(r)["project_name"]
			project, serr := h.projectService.GetProjectByOrgAndName(r.Context(), orgID, projectName)
			if serr != nil {
				return nil, serr
			}
			updated, serr := h.projectService.UpdateProject(r.Context(), project.ID, &models.Project{Name: req.GetName()})
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentProject(updated), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *projectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			projectName := mux.Vars(r)["project_name"]
			project, serr := h.projectService.GetProjectByOrgAndName(r.Context(), orgID, projectName)
			if serr != nil {
				return nil, serr
			}
			return nil, h.projectService.DeleteProject(r.Context(), project.ID)
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}

func (h *projectHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	var req openapi.AddProjectMemberRequest
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			projectName := mux.Vars(r)["project_name"]
			project, serr := h.projectService.GetProjectByOrgAndName(r.Context(), orgID, projectName)
			if serr != nil {
				return nil, serr
			}
			membership, serr := h.projectService.AddMember(r.Context(), project.ID, req.GetUserId(), presenters.ConvertProjectRole(req.GetRole()))
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentProjectMembership(membership), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *projectHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			projectName := mux.Vars(r)["project_name"]
			project, serr := h.projectService.GetProjectByOrgAndName(r.Context(), orgID, projectName)
			if serr != nil {
				return nil, serr
			}
			memberships, serr := h.projectService.ListMembers(r.Context(), project.ID)
			if serr != nil {
				return nil, serr
			}
			return openapi.ProjectMembershipList{
				Items: presenters.PresentProjectMembershipList(memberships),
				Total: ptr.To(int32(len(memberships))),
			}, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *projectHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	var req openapi.UpdateProjectMemberRoleRequest
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			membershipID := mux.Vars(r)["id"]
			membership, serr := h.projectService.UpdateMemberRole(r.Context(), membershipID, presenters.ConvertProjectRole(req.GetRole()))
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentProjectMembership(membership), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *projectHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			membershipID := mux.Vars(r)["id"]
			return nil, h.projectService.RemoveMember(r.Context(), membershipID)
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}

func (h *projectHandler) ListProjectRoles(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			return openapi.ProjectRoleList{
				Roles: []openapi.ProjectRole{openapi.DEVELOPER, openapi.VIEWER},
			}, nil
		},
	}
	handleGet(w, r, cfg)
}
