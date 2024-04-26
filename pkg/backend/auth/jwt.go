package auth
// https://pascalallen.medium.com/jwt-authentication-with-go-242215a9b4f8

import (
	"go.savla.devlittr/pkg/backend/users"

	"github.com/golang-jwt/jwt"
)

type UserClaims struct {
	Nickname string      `json:"nickname"`
	User     users.User `json:"user"`
	jwt.StandardClaims
}

func NewAccessToken(claims UserClaims, secret string) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return accessToken.SignedString([]byte(secret))
}

func NewRefreshToken(claims jwt.StandardClaims, secret string) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return refreshToken.SignedString([]byte(secret))
}

func ParseAccessToken(accessToken string, secret string) *UserClaims {
	parsedAccessToken, _ := jwt.ParseWithClaims(accessToken, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	return parsedAccessToken.Claims.(*UserClaims)
}

func ParseRefreshToken(refreshToken string, secret string) *jwt.StandardClaims {
	parsedRefreshToken, _ := jwt.ParseWithClaims(refreshToken, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	return parsedRefreshToken.Claims.(*jwt.StandardClaims)
}
