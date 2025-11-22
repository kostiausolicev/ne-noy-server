package service

import (
	"time"

	"gopkg.in/dgrijalva/jwt-go.v3"
)

type jwtService struct {
	secret string
}

func (j jwtService) GenerateToken(claims map[string]interface{}) (string, error) {
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	tokenString, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

type JWTService interface {
	GenerateToken(claims map[string]interface{}) (string, error)
}

func NewJWTService(secret string) JWTService {
	return jwtService{secret: secret}
}
