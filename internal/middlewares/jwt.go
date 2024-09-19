package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/models"
	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/services"
)

type userFieldType string

const userField userFieldType = "userField"

type AuthMiddlewareConfig struct {
	excludePaths []string
}

func AuthMiddleware() *AuthMiddlewareConfig {
	return &AuthMiddlewareConfig{}
}

func (a *AuthMiddlewareConfig) WithExcludedPaths(paths ...string) *AuthMiddlewareConfig {
	a.excludePaths = paths
	return a
}

func (a *AuthMiddlewareConfig) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, path := range a.excludePaths {
			if strings.HasPrefix(r.URL.Path, path) {
				next.ServeHTTP(w, r)
				return
			}
		}

		authService := GetServiceFromContext[models.AuthService](w, r, AuthServiceKey)
		jwtService := GetServiceFromContext[models.JWTService](w, r, JwtServiceKey)

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenString == "" {
			http.Error(w, "Bearer token is empty", http.StatusUnauthorized)
			return
		}

		token, err := (*jwtService).ValidateToken(tokenString)

		if err != nil {
			if errors.Is(err, services.ErrTokenIsInvalid) {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if errors.Is(err, services.ErrTokenIsExpired) {
				http.Error(w, "Token is expired", http.StatusUnauthorized)
				return
			}

			http.Error(w, fmt.Sprintf("Error occurred during validating token: %s", err.Error()), http.StatusUnauthorized)
			return
		}

		login, err := token.Claims.GetSubject()

		if err != nil {
			http.Error(w, fmt.Sprintf("Error occurred during reading sub field: %s", err.Error()), http.StatusUnauthorized)
			return
		}

		user, err := (*authService).GetUser(r.Context(), login)

		if err != nil {
			if errors.Is(err, services.ErrUserIsNotExist) {
				http.Error(w, fmt.Sprintf("UnknownUser login %s doesn't exist", login), http.StatusConflict)
				return
			}

			http.Error(w, fmt.Sprintf("Error occurred during validation user login: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userField, user)))
	})
}

func GetUserFromContext(w http.ResponseWriter, r *http.Request) *models.User {
	user, ok := r.Context().Value(userField).(*models.User)

	if !ok {
		http.Error(w, "Could not retrieve user from context", http.StatusInternalServerError)
		return nil
	}

	return user
}
