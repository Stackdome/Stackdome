package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
)

type OpenAPIMiddleware struct {
	muxRouter routers.Router
}

func NewOpenAPIMiddleWare(rawSpec []byte) (*OpenAPIMiddleware, error) {
	openapi3.SchemaErrorDetailsDisabled = true
	ctx := context.Background()
	loader := &openapi3.Loader{
		Context:               ctx,
		IsExternalRefsAllowed: true}

	doc, err := loader.LoadFromData(rawSpec)
	if err != nil {
		return nil, err
	}

	err = doc.Validate(ctx)
	if err != nil {
		return nil, err
	}

	// Because openshift router is using Host as IP, we skip the server
	// validation to allow all the possible entries
	doc.Servers = []*openapi3.Server{}

	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, err
	}
	return &OpenAPIMiddleware{
		muxRouter: router,
	}, nil
}

// Middleware function, which will be called for each request
func (m *OpenAPIMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, pathParams, _ := m.muxRouter.FindRoute(r)
		if route == nil {
			// The endpoint may not be defined in the OpenAPI spec.
			// If this is the case, validation need not occur;
			// processing of the endpoint can be passed to the next handler.
			next.ServeHTTP(w, r)
			return
		}
		requestValidationInput := &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			},
		}
		err := openapi3filter.ValidateRequest(context.TODO(), requestValidationInput)
		if err != nil {
			serviceErr := errors.Validation("Failing openapi validation: %s", err)
			handleError(context.TODO(), w, serviceErr.Code, serviceErr.Reason)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleError(ctx context.Context, w http.ResponseWriter, code errors.ServiceErrorCode, reason string) {
	log := logger.NewLogger()
	err := errors.New(code, "%s", reason)
	if err.HttpCode >= 400 && err.HttpCode <= 499 {
		log.Infof(err.Error())
	} else {
		log.Errorf(err.Error())
	}

	writeJSONResponse(w, err.HttpCode, err.AsOpenapiError())
}

func writeJSONResponse(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if payload != nil {
		response, _ := json.Marshal(payload)
		_, _ = w.Write(response)
	}
}
