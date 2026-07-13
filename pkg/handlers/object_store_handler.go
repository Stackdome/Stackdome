package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

type ObjectStoreHandlerSpec struct {
	ObjectStoreService services.ObjectStoreService
	ProjectService     services.ProjectService
}

type objectStoreHandler struct {
	objectStoreService services.ObjectStoreService
	projectService     services.ProjectService
}

func NewObjectStoreHandler(spec ObjectStoreHandlerSpec) *objectStoreHandler {
	return &objectStoreHandler{
		objectStoreService: spec.ObjectStoreService,
		projectService:     spec.ProjectService,
	}
}

func (h *objectStoreHandler) Create(w http.ResponseWriter, r *http.Request) {
	var apiObjectStore openapi.ObjectStore
	cfg := &handlerConfig{
		MarshalInto: &apiObjectStore,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			projectID, serr := resolveProjectID(r, h.projectService)
			if serr != nil {
				return nil, serr
			}

			objectStore := presenters.ConvertObjectStore(&apiObjectStore)
			objectStore.OrganisationID = orgID
			objectStore.ProjectID = projectID

			createdObjectStore, err := h.objectStoreService.Create(ctx, objectStore)
			if err != nil {
				return nil, err
			}

			return presenters.PresentObjectStore(createdObjectStore), nil
		},
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *objectStoreHandler) ListByOrgID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			objs, serr := h.objectStoreService.ListObjectStoresForCurrentUser(r.Context(), orgID)
			if serr != nil {
				return nil, serr
			}
			return openapi.ObjectStoreList{
				Items: presenters.PresentObjectStoreList(objs),
				Total: ptr.To(int32(len(objs))),
			}, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *objectStoreHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			projectID, serr := resolveProjectID(r, h.projectService)
			if serr != nil {
				return nil, serr
			}

			objectStores, err := h.objectStoreService.ListByProjectID(ctx, projectID)
			if err != nil {
				return nil, err
			}

			return openapi.ObjectStoreList{
				Items: presenters.PresentObjectStoreList(objectStores),
				Total: ptr.To(int32(len(objectStores))),
			}, nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *objectStoreHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()

			objectStore, err := h.objectStoreService.GetByID(ctx, id)
			if err != nil {
				return nil, err
			}

			return presenters.PresentObjectStore(objectStore), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *objectStoreHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var apiObjectStore openapi.ObjectStore
	cfg := &handlerConfig{
		MarshalInto: &apiObjectStore,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()

			objectStore := presenters.ConvertObjectStore(&apiObjectStore)

			updatedObjectStore, err := h.objectStoreService.Update(ctx, id, objectStore)
			if err != nil {
				return nil, err
			}

			return presenters.PresentObjectStore(updatedObjectStore), nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *objectStoreHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()

			if err := h.objectStoreService.Delete(ctx, id); err != nil {
				return nil, err
			}

			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
