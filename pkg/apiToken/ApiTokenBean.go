package apiToken

import "github.com/golang-jwt/jwt/v4"

type ApiTokenCustomClaims struct {
	Email   string `json:"email"`
	Version string `json:"version"`
	jwt.RegisteredClaims
}
