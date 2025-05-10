package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/handlers/validation"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
)

func NewUserServiceHandler(spec UserServiceHandlerSpec) *usersHandler {
	return &usersHandler{
		userService: spec.UserService,
		authzClient: spec.AuthzClient,
	}
}

type UserServiceHandlerSpec struct {
	UserService services.UserService
	AuthzClient auth.AuthorizationClient
}

type usersHandler struct {
	userService services.UserService
	authzClient auth.AuthorizationClient
}

func (a usersHandler) Get(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()
			currentUser, ctxerr := auth.GetCurrentUserFromCtx(ctx)
			if ctxerr != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			id := mux.Vars(r)["id"]

			result, err := a.userService.Get(ctx, id)
			if err != nil {
				return nil, err
			}

			allowed, accessErr := a.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.User,
				id,
				result.ID,
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize user access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to access user '%s'", currentUser.ID, id)
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
			currentUser, ctxerr := auth.GetCurrentUserFromCtx(ctx)
			if ctxerr != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			result, err := a.userService.Get(ctx, currentUser.ID)
			if err != nil {
				return nil, err
			}
			return presenters.PresentUser(result), nil
		},
	}
	handleGet(w, r, cfg)
}

func (a usersHandler) Create(w http.ResponseWriter, r *http.Request) {
	var user openapi.UserCreateRequest
	cfg := &handlerConfig{
		&user,
		validation.ValidateUserCreate(&user),
		func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()
			convertedUser := presenters.ConvertUser(&user)

			user, err := a.userService.Create(ctx, convertedUser)
			if err != nil {
				return nil, err
			}

			return presenters.PresentUser(user), nil
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
