package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

func NewVolumeHandler(spec VolumeHandlerSpec) *volumeHandler {
	return &volumeHandler{
		volumeService:  spec.VolumeService,
		projectService: spec.ProjectService,
	}
}

type VolumeHandlerSpec struct {
	VolumeService  services.VolumeService
	ProjectService services.ProjectService
}

type volumeHandler struct {
	volumeService  services.VolumeService
	projectService services.ProjectService
}

func (h *volumeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			if id == "current" {
				projectID, serr := resolveProjectID(r, h.projectService)
				if serr != nil {
					return nil, serr
				}
				currentUser, err := auth.GetCurrentUserFromCtx(ctx)
				if err != nil {
					return nil, errors.Unauthorized("failed to fetch current user")
				}
				objs, serr := h.volumeService.ListByUserID(ctx, projectID, currentUser.ID)
				if serr != nil {
					return nil, serr
				}
				listResp := openapi.VolumeList{
					Items: presenters.PresentVolumeList(objs, true),
					Total: ptr.To(int32(len(objs))),
				}
				return listResp, nil
			}
			obj, err := h.volumeService.Get(ctx, id)
			if err != nil {
				return nil, err
			}
			return presenters.PresentVolume(obj, true), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *volumeHandler) ListByStackID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			stackID := mux.Vars(r)["id"]
			objs, err := h.volumeService.ListVolumesUsedByStack(ctx, stackID)
			if err != nil {
				return nil, err
			}
			listResp := openapi.VolumeList{
				Items: presenters.PresentVolumeList(objs, true),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *volumeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			serr := h.volumeService.Delete(ctx, id)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
