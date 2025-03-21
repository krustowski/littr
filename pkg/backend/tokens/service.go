package tokens

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"

	"github.com/golang-jwt/jwt"
)

/*type TokenServiceInterface interface {
	Create(ctx context.Context, token *Token) error
	Delete(ctx context.Context, tokenID string) error
	FindByID(ctx context.Context, tokenID string) (*Token, error)
}*/

type TokenService struct {
	tokenRepository models.TokenRepositoryInterface
}

func NewTokenService(tokenRepository models.TokenRepositoryInterface) models.TokenServiceInterface {
	return &TokenService{
		tokenRepository: tokenRepository,
	}
}

func (s *TokenService) Create(ctx context.Context, user *models.User) ([]string, error) {
	if user == nil {
		return nil, fmt.Errorf("given user is nil")
	}

	tokens, err := NewToken(user, s.tokenRepository)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func NewToken(user *models.User, r models.TokenRepositoryInterface) ([]string, error) {
	secret := os.Getenv("APP_PEPPER")

	if secret == "" {
		return nil, fmt.Errorf("server secret is blank")
	}

	// Compose the user's personal (access) token content.
	userClaims := UserClaims{
		Nickname: user.Nickname,
		// Access token is restricted to 15 minutes of its validity.
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
	}

	// Get new access token = sign the access token with the server's secret.
	signedAccessToken, err := NewAccessToken(userClaims, secret)
	if err != nil {
		return nil, fmt.Errorf(common.ERR_AUTH_ACC_TOKEN_FAIL)
	}

	// Compose the user's personal (refresh) token content. Refresh token is restricted (mainly) to 4 weeks of its validity.
	refreshClaims := jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(common.TOKEN_TTL).Unix(),
	}

	// Get new refresh token = sign the refresh token with the server's secret.
	signedRefreshToken, err := NewRefreshToken(refreshClaims, secret)
	if err != nil {
		return nil, fmt.Errorf(common.ERR_AUTH_REF_TOKEN_FAIL)
	}

	// Prepare the refresh token's hash for the database payload.
	refreshSum := sha256.New()
	refreshSum.Write([]byte(signedRefreshToken))
	refreshTokenSum := fmt.Sprintf("%x", refreshSum.Sum(nil))

	// Prepare the refresh token struct for the database saving.
	token := models.Token{
		Hash:      refreshTokenSum,
		CreatedAt: time.Now(),
		Nickname:  user.Nickname,
		TTL:       common.TOKEN_TTL,
	}

	// Save new refresh token's hash to the Token database.
	if err := r.Save(&token); err != nil {
		return nil, fmt.Errorf(common.ERR_TOKEN_SAVE_FAIL)
	}

	return []string{signedAccessToken, signedRefreshToken}, nil
}

func (s *TokenService) Delete(ctx context.Context, tokenID string) error {
	return errNotImplemented
}

func (s *TokenService) FindByID(ctx context.Context, tokenID string) (*models.Token, error) {
	return nil, errNotImplemented
}
