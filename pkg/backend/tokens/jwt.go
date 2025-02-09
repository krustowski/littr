package tokens

// https://pascalallen.medium.com/jwt-authentication-with-go-242215a9b4f8

import (
	"github.com/golang-jwt/jwt"
)

// UserClaims is a generic structure for a personal user's (access) token.
type UserClaims struct {
	Nickname string `json:"nickname"`
	jwt.StandardClaims
}

// NewAccessToken generates a new signed access token.
func NewAccessToken(claims UserClaims, secret string) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return accessToken.SignedString([]byte(secret))
}

// NewRefreshToken generates a new signed refresh token.
func NewRefreshToken(claims jwt.StandardClaims, secret string) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return refreshToken.SignedString([]byte(secret))
}

// ParseAccessToken decodes the accessCookie value to get the UserClaims payload.
func ParseAccessToken(accessToken string, secret string) *UserClaims {
	parsedAccessToken, _ := jwt.ParseWithClaims(accessToken, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	// Assert type pointer to UserClaims.
	return parsedAccessToken.Claims.(*UserClaims)
}

// ParseRefreshToken decodes the refreshCookie value to get the StandardClaims payload.
func ParseRefreshToken(refreshToken string, secret string) *jwt.StandardClaims {
	parsedRefreshToken, _ := jwt.ParseWithClaims(refreshToken, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if parsedRefreshToken == nil {
		return nil
	}

	// Assert type pointer to StandardClaims.
	return parsedRefreshToken.Claims.(*jwt.StandardClaims)
}
