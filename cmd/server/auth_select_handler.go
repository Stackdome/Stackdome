package server

import (
	"net/http"
	"regexp"
)

type SelectFunc = func(r *http.Request) bool

type HandlerSelection struct {
	handler    http.Handler
	selectFunc SelectFunc
}

type AuthSelectorHandler struct {
	selections     []*HandlerSelection
	defaultHandler http.Handler
	publicPaths    []*regexp.Regexp
	mainHandler    http.Handler
}

type AuthSelectorHandlerSpec struct {
	PublicPaths        []string
	MainHandler        http.Handler
	DefaultAuthHandler http.Handler
}

func NewAuthSelectHandler(spec AuthSelectorHandlerSpec) *AuthSelectorHandler {
	publicExprs := make([]*regexp.Regexp, len(spec.PublicPaths))
	for i, expr := range spec.PublicPaths {
		res := regexp.MustCompile(expr)
		publicExprs[i] = res
	}
	return &AuthSelectorHandler{
		mainHandler:    spec.MainHandler,
		defaultHandler: spec.DefaultAuthHandler,
		publicPaths:    publicExprs,
		selections:     make([]*HandlerSelection, 0),
	}
}

func (h *AuthSelectorHandler) Add(handler http.Handler, selectFunc SelectFunc) {
	h.selections = append(
		h.selections,
		&HandlerSelection{
			handler:    handler,
			selectFunc: selectFunc,
		})
}

func (h *AuthSelectorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, expr := range h.publicPaths {
		if expr.MatchString(r.URL.Path) {
			h.mainHandler.ServeHTTP(w, r)
			return
		}
	}

	for _, selection := range h.selections {
		if selection.selectFunc(r) {
			selection.handler.ServeHTTP(w, r)
			return
		}
	}
	h.defaultHandler.ServeHTTP(w, r)
}
