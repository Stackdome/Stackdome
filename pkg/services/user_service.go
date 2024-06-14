package services

import (
	"context"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Get(ctx context.Context, ID string) (*models.User, *errors.ServiceError)
	Create(ctx context.Context, user *models.User) (*models.User, *errors.ServiceError)
	Login(ctx context.Context, loginRequest *openapi.LoginRequest) (*openapi.LoginResponse, *errors.ServiceError)
}

var _ UserService = &usersService{}

func NewUserService(spec UserServiceSpec) UserService {
	return &usersService{
		userStore: pgstore.NewUserStore(
			pgstore.UserStoreSpec{
				SessionFactory: spec.SessionFactory,
			},
		),
		logger:       spec.Logger,
		jwtSecretKey: spec.JwtSecretKey,
	}
}

type UserServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
	JwtSecretKey   string
}

type usersService struct {
	userStore    stores.UserStore
	logger       logger.Logger
	jwtSecretKey string
}

func (u usersService) Get(ctx context.Context, ID string) (*models.User, *errors.ServiceError) {
	return u.userStore.GetByID(ctx, ID)
}

func (u usersService) Create(ctx context.Context, user *models.User) (*models.User, *errors.ServiceError) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		u.logger.Errorf("failed to hash password, %s", err.Error())
		return nil, errors.GeneralError("failed to create user")
	}
	user.Password = string(hashedPassword)
	if len(user.Role) == 0 {
		user.Role = models.UserRole
	}
	return u.userStore.Create(ctx, user)
}

func (u usersService) Login(ctx context.Context, loginRequest *openapi.LoginRequest) (*openapi.LoginResponse, *errors.ServiceError) {
	userInDB, err := u.userStore.GetByEmail(ctx, loginRequest.GetEmail())
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userInDB.Password), []byte(loginRequest.Password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, errors.BadRequest("invalid email or password")
		}
		u.logger.Errorf("failed to login: %s", err.Error())
		return nil, errors.GeneralError("failed to login")
	}
	expirationTime := time.Now().UTC().Add(24 * time.Hour)
	claims := &auth.Claims{
		UserID: userInDB.ID,
		Role:   string(userInDB.Role),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, tokenErr := token.SignedString([]byte(u.jwtSecretKey))
	if tokenErr != nil {
		return nil, errors.GeneralError("failed to generate token: %s", tokenErr.Error())
	}

	res := openapi.NewLoginResponse()
	res.SetToken(tokenString)
	return res, nil
}
