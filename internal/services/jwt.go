package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenIsInvalid = errors.New("token is invalid")
	ErrTokenIsExpired = errors.New("token is expired")
)

type JWTService struct {
	authSecretKey string
}

func NewJWTService(authSecretKey string) *JWTService {
	return &JWTService{authSecretKey}
}

func (j *JWTService) GenerateJWT(subject string) (string, error) {
	now := time.Now()
	tokenString, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": subject,
			"iat": now.Unix(),
			"exp": now.Add(24 * time.Hour).Unix(),
		}).SignedString([]byte(j.authSecretKey))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (j *JWTService) ValidateToken(token string) (*jwt.Token, error) {
	claims := &jwt.RegisteredClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(j.authSecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenIsExpired
		}

		return nil, err
	}

	if !parsedToken.Valid {
		return nil, ErrTokenIsInvalid
	}

	return parsedToken, nil
}
