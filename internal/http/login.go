package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/middlewares"
	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/models"
	"github.com/Renal37/go-musthave-diploma-tpl/tree/master/internal/services"
)

// todo add invalidation prev token
func Login(w http.ResponseWriter, r *http.Request) {
	data := middlewares.GetParsedJSONData[models.UnknownUser](w, r)
	authService := middlewares.GetServiceFromContext[models.AuthService](w, r, middlewares.AuthServiceKey)
	jwtService := middlewares.GetServiceFromContext[models.JWTService](w, r, middlewares.JwtServiceKey)

	if ok := IsUnknownUserDataValid(data); !ok {
		http.Error(w, "Request doesn't contain login or password", http.StatusBadRequest)
		return
	}

	if err := (*authService).Login(r.Context(), data); err != nil {
		if errors.Is(err, services.ErrUserIsNotExist) {
			http.Error(w, fmt.Sprintf("Login %s is not exist", *data.Login), http.StatusUnauthorized)
			return
		}

		if errors.Is(err, services.ErrPasswordIsIncorrect) {
			http.Error(w, "Password is not correct", http.StatusUnauthorized)
			return
		}

		http.Error(w, fmt.Sprintf("Error occurred during login: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	token, err := (*jwtService).GenerateJWT(*data.Login)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error occurred during generating jwt token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
}
