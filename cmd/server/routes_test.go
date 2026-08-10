package server

import (
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/cmd/environment"
	"github.com/Stackdome/stackdome/config"
)

const openAPISpecPath = "../../config/openapi/stackdome_api.yaml"

// Registered routes that deliberately have no OpenAPI operation. Everything
// else must be in the spec, or the parity test below fails.
//
// Entries are "METHOD /normalised/path" — path params collapsed to {}.
var routesWithoutSpecOperation = []string{
	// SPA + health probe: not part of the REST contract.
	"GET /health",

	// SSE streams. The spec cannot describe a text/event-stream body, and the
	// generated clients would be wrong if it tried.
	"GET /api/v1/organizations/{}/projects/{}/stacks/{}/releases/{}/events/stream",

	// GitHub browser redirects and the webhook receiver. Called by GitHub, never
	// by a generated client.
	"GET /api/v1/git-integrations/github/manifest/callback",
	"GET /api/v1/git-integrations/github/setup",
	"POST /api/v1/webhooks/github",
}

// Spec operations with no registered route. Keep this list short; an entry here
// is a promise the API does not always keep.
var specOperationsWithoutRoute = []string{
	// Registered only when GitHub OAuth is configured, which the test env is not.
	"GET /api/v1/auth/github",
	"GET /api/v1/auth/github/callback",
}

var pathParam = regexp.MustCompile(`\{[^}]*\}`)

// Route and spec use different param names for the same slot ({id} vs
// {stack_id}), so compare on shape only.
func normalisePath(path string) string {
	return pathParam.ReplaceAllString(path, "{}")
}

func endpointKey(method, path string) string {
	return method + " " + normalisePath(path)
}

var _ = Describe("route/OpenAPI parity", func() {
	var registeredEndpoints, specEndpoints []string

	BeforeEach(func() {
		registeredEndpoints, specEndpoints = nil, nil
		router := apiServer{environment: environment.NewTestEnvironment(nil)}.routes()

		err := router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
			path, err := route.GetPathTemplate()
			if err != nil {
				// Catch-all SPA route has no path template.
				return nil //nolint:nilerr
			}
			methods, err := route.GetMethods()
			if err != nil {
				return nil //nolint:nilerr
			}
			for _, method := range methods {
				registeredEndpoints = append(registeredEndpoints, endpointKey(method, path))
			}
			return nil
		})
		Expect(err).NotTo(HaveOccurred())

		rawSpec, err := os.ReadFile(openAPISpecPath)
		Expect(err).NotTo(HaveOccurred())
		doc, err := openapi3.NewLoader().LoadFromData(rawSpec)
		Expect(err).NotTo(HaveOccurred())

		basePath := specBasePath(doc)
		for path, item := range doc.Paths.Map() {
			for method := range item.Operations() {
				specEndpoints = append(specEndpoints, endpointKey(method, basePath+path))
			}
		}
	})

	It("has an OpenAPI operation for every registered route", func() {
		Expect(missing(registeredEndpoints, specEndpoints, routesWithoutSpecOperation)).To(BeEmpty(),
			"routes are not in %s. Add the operation, or add the route to routesWithoutSpecOperation with a reason.", openAPISpecPath)
	})

	It("has a registered route for every OpenAPI operation", func() {
		Expect(missing(specEndpoints, registeredEndpoints, specOperationsWithoutRoute)).To(BeEmpty(),
			"spec operations have no route. Register the route, or drop the operation from %s.", openAPISpecPath)
	})
})

var _ = Describe("workspace-user route registration", func() {
	const workspaceUsersPath = "/api/v1/organizations/{org_id}/projects/{project_name}/workspace-users"

	registeredEndpoints := func(stackdomeCloudRuntime, workspaceUsersEnabled bool) []string {
		cfg := config.NewApplicationConfig()
		if stackdomeCloudRuntime {
			cfg.RuntimeMode = config.RuntimeModeStackdomeCloud
			cfg.StackdomeCloud = &config.StackdomeCloudConfig{
				Features: config.StackdomeCloudFeaturesConfig{WorkspaceUsers: workspaceUsersEnabled},
			}
		}

		router := apiServer{
			environment: environment.NewTestEnvironment(nil, environment.WithApplicationConfig(cfg)),
		}.routes()
		var endpoints []string
		Expect(router.Walk(func(route *mux.Route, _ *mux.Router, _ []*mux.Route) error {
			path, err := route.GetPathTemplate()
			if err != nil || !strings.HasPrefix(path, workspaceUsersPath) {
				return nil
			}
			methods, methodErr := route.GetMethods()
			if methodErr != nil {
				return methodErr
			}
			for _, method := range methods {
				endpoints = append(endpoints, method+" "+path)
			}
			return nil
		})).To(Succeed())
		sort.Strings(endpoints)
		return endpoints
	}

	It("keeps the self-hosted routes enabled by default", func() {
		expected := []string{
			http.MethodPost + " " + workspaceUsersPath,
			http.MethodGet + " " + workspaceUsersPath + "/current",
			http.MethodGet + " " + workspaceUsersPath + "/{id}",
			http.MethodPut + " " + workspaceUsersPath + "/{id}",
			http.MethodDelete + " " + workspaceUsersPath + "/{id}",
		}
		sort.Strings(expected)
		Expect(registeredEndpoints(false, false)).To(Equal(expected))
	})

	It("omits the routes when the Stackdome Cloud feature is disabled", func() {
		Expect(registeredEndpoints(true, false)).To(BeEmpty())
	})

	It("registers the routes when the Stackdome Cloud feature is enabled", func() {
		Expect(registeredEndpoints(true, true)).ToNot(BeEmpty())
	})
})

// The spec declares its prefix on the server URL, not on each path.
func specBasePath(doc *openapi3.T) string {
	if len(doc.Servers) == 0 {
		return ""
	}
	url := doc.Servers[0].URL
	if i := strings.Index(url, "/api/"); i >= 0 {
		return strings.TrimSuffix(url[i:], "/")
	}
	return ""
}

func missing(want, have, exceptions []string) []string {
	index := map[string]bool{}
	for _, key := range have {
		index[key] = true
	}
	for _, key := range exceptions {
		index[key] = true
	}

	var result []string
	for _, key := range want {
		if !index[key] {
			result = append(result, key)
		}
	}
	sort.Strings(result)
	return result
}
