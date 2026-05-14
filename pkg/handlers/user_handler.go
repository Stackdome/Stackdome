package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/handlers/validation"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

func NewUserServiceHandler(spec UserServiceHandlerSpec) *usersHandler {
	return &usersHandler{
		userService:   spec.UserService,
		signupService: spec.SignupService,
	}
}

type UserServiceHandlerSpec struct {
	UserService   services.UserService
	SignupService services.SignupService
}

type usersHandler struct {
	userService   services.UserService
	signupService services.SignupService
}

func (a usersHandler) Get(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]

			result, err := a.userService.Get(ctx, id)
			if err != nil {
				return nil, err
			}

			return presenters.PresentUser(result), nil
		},
	}
	handleGet(w, r, cfg)
}

func (a usersHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()
			user, memberships, err := a.userService.GetUserFromContext(ctx)
			if err != nil {
				return nil, err
			}
			return presenters.PresentUserWithTeams(user, memberships), nil
		},
	}
	handleGet(w, r, cfg)
}

func (a usersHandler) ListByOrgID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			params := parseListParams(r, nil)

			result, serr := a.userService.ListByOrgID(r.Context(), orgID, params)
			if serr != nil {
				return nil, serr
			}

			presented := make([]openapi.User, len(result.Items))
			for i, u := range result.Items {
				presented[i] = presenters.PresentUser(u)
			}
			return openapi.UserList{
				Items:      presented,
				Total:      ptr.To(int32(result.Total)),
				Page:       ptr.To(int32(result.Page)),
				PageSize:   ptr.To(int32(result.PageSize)),
				TotalPages: ptr.To(int32(result.TotalPages)),
			}, nil
		},
	}
	handleList(w, r, cfg)
}

func (a usersHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req openapi.UserSignupRequest
	cfg := &handlerConfig{
		&req,
		validation.ValidateUserCreate(&req),
		func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()
			convertedUser := presenters.ConvertUser(&req)

			inviteToken := ""
			if req.InviteToken != nil {
				inviteToken = *req.InviteToken
			}

			user, err := a.signupService.Signup(ctx, convertedUser, inviteToken)
			if err != nil {
				return nil, err
			}

			return user, nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (a usersHandler) Login(w http.ResponseWriter, r *http.Request) {
	var login openapi.LoginRequest
	cfg := &handlerConfig{
		&login,
		validation.ValidateUserLogin(&login),
		func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()
			response, err := a.userService.Login(ctx, &login)
			if err != nil {
				return nil, err
			}

			return response, nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}
