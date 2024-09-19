package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/middlewares"
	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/models"
	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/services"
)

func IsUnknownUserDataValid(data models.UnknownUser) bool {
	if data.Login == nil || data.Password == nil {
		return false
	}

	return true
}

func Register(w http.ResponseWriter, r *http.Request) {
	data := middlewares.GetParsedJSONData[models.UnknownUser](w, r)
	authService := middlewares.GetServiceFromContext[models.AuthService](w, r, middlewares.AuthServiceKey)
	jwtService := middlewares.GetServiceFromContext[models.JWTService](w, r, middlewares.JwtServiceKey)

	if ok := IsUnknownUserDataValid(data); !ok {
		http.Error(w, "Request doesn't contain login or password", http.StatusBadRequest)
		return
	}

	token, err := (*jwtService).GenerateJWT(*data.Login)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error occurred during generating jwt token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err := (*authService).Register(r.Context(), data); err != nil {
		if errors.Is(err, services.ErrUserIsAlreadyRegistered) {
			http.Error(w, "User is already registered", http.StatusConflict)
			return
		}

		http.Error(w, fmt.Sprintf("Error occurred during registration: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
}
