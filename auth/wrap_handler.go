package auth

import (
	"context"
	"net/http"
)

var authenticated = "authenticated"
var teamNameKey = "teamName"
var teamIDKey = "teamID"
var isAdminKey = "isAdmin"
var isSystemKey = "system"

func WrapHandler(
	handler http.Handler,
	validator Validator,
	userContextReader UserContextReader,
) http.Handler {
	return authHandler{
		handler:           handler,
		validator:         validator,
		userContextReader: userContextReader,
	}
}

type authHandler struct {
	handler           http.Handler
	validator         Validator
	userContextReader UserContextReader
}

func (h authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), authenticated, h.validator.IsAuthenticated(r))
	teamName, teamID, isAdmin, found := h.userContextReader.GetTeam(r)
	if found {
		ctx = context.WithValue(ctx, teamNameKey, teamName)
		ctx = context.WithValue(ctx, teamIDKey, teamID)
		ctx = context.WithValue(ctx, isAdminKey, isAdmin)
	}

	isSystem, found := h.userContextReader.GetSystem(r)
	if found {
		ctx = context.WithValue(ctx, isSystemKey, isSystem)
	}
	h.handler.ServeHTTP(w, r.WithContext(ctx))
}
