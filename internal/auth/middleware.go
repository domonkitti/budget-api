package auth

import (
	"context"
	"net/http"
)

type contextKey string

// UserKey is the context key for the authenticated user.
const UserKey contextKey = "auth_user"

// User holds identity info extracted from a valid token.
// Populate the fields when real auth is implemented.
type User struct {
	ID    string
	Email string
	Role  string
}

// FromContext retrieves the authenticated user from a request context.
// Returns nil if auth middleware is not enabled or the user is not present.
func FromContext(ctx context.Context) *User {
	u, _ := ctx.Value(UserKey).(*User)
	return u
}

// Require is a chi-compatible middleware that enforces authentication.
// Currently a passthrough — replace the body with JWT/session validation when ready.
//
// To enable:
//  1. Validate the token from r.Header.Get("Authorization") or a cookie
//  2. Populate a User struct and inject it: ctx := context.WithValue(r.Context(), UserKey, &user)
//  3. Call next.ServeHTTP(w, r.WithContext(ctx))
//  4. In main.go, add: r.Use(auth.Require) inside r.Route("/api/v1", ...)
func Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: validate token here
		// token := r.Header.Get("Authorization")
		// user, err := validateToken(strings.TrimPrefix(token, "Bearer "))
		// if err != nil { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
		// ctx := context.WithValue(r.Context(), UserKey, &user)
		// next.ServeHTTP(w, r.WithContext(ctx))
		next.ServeHTTP(w, r)
	})
}
