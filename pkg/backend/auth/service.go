package auth

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/tokens"
	"go.vxn.dev/littr/pkg/models"
)

/*type AuthServiceInterface interface {
	Auth(ctx context.Context, auth.AuthUser) error
	Logout(ctx context.Context) error
}*/

type AuthService struct {
	tokenRepository models.TokenRepositoryInterface
	userRepository  models.UserRepositoryInterface
}

func NewAuthService(
	tokenRepository models.TokenRepositoryInterface,
	userRepository models.UserRepositoryInterface,
) models.AuthServiceInterface {
	if tokenRepository == nil || userRepository == nil {
		return nil
	}

	return &AuthService{
		tokenRepository: tokenRepository,
		userRepository:  userRepository,
	}
}

func (s *AuthService) Auth(ctx context.Context, authUserI interface{}) (*models.User, []string, error) {
	authUser, ok := authUserI.(*AuthUser)
	if !ok {
		return nil, nil, fmt.Errorf("cannot assert type AuthUser")
	}

	// Fetch one user from cache according to the login credentials.
	dbUser, err := s.userRepository.GetByID(authUser.Nickname)
	if err != nil {
		return nil, nil, err
	}

	// Check the passhashes.
	if dbUser.Passphrase == authUser.Passphrase || dbUser.PassphraseHex == authUser.PassphraseHex {
		// Legacy: update user's hexadecimal passphrase form, as the binary form is broken and cannot be used on BE.
		if dbUser.PassphraseHex == "" && authUser.PassphraseHex != "" {
			dbUser.PassphraseHex = authUser.PassphraseHex

			// Note the user tried to login now.
			dbUser.LastLoginTime = time.Now()

			if err := s.userRepository.Save(dbUser); err != nil {
				return nil, nil, err
			}
		}
	} else {
		// Return auth fail.
		return nil, nil, fmt.Errorf(common.ERR_AUTH_FAIL)
	}

	// Check if the user has been activated yet.
	if !dbUser.Active || !dbUser.Options["active"] {
		return nil, nil, fmt.Errorf(common.ERR_USER_NOT_ACTIVATED)
	}

	//
	//  OK, user authorized, now generete tokens
	//

	tks, err := tokens.NewToken(dbUser, s.tokenRepository)
	if err != nil {
		return nil, nil, err
	}

	// User authorized.
	return dbUser, tks, nil
}

func (s *AuthService) Logout(ctx context.Context) error {
	// Fetch the server's secret.
	secret := os.Getenv("APP_PEPPER")
	if secret == "" {
		return fmt.Errorf(common.ERR_NO_SERVER_SECRET)
	}

	var refreshCookie *http.Cookie
	var ok bool

	// Get the refresh cookie to check its validity (not necessary atm).
	if refreshCookie, ok = ctx.Value("refreshCookie").(*http.Cookie); !ok {
		return fmt.Errorf(common.ERR_BLANK_REF_TOKEN)
	}

	// Decode the contents of the refresh HTTP cookie, compare the signature with the server's secret.
	refreshClaims := tokens.ParseRefreshToken(refreshCookie.Value, secret)

	// If the refresh token is expired => user should relogin.
	if refreshClaims.Valid() != nil {
		return fmt.Errorf(common.ERR_INVALID_REF_TOKEN)
	}

	// Get the refresh token's fingerprint.
	refreshSum := sha256.New()
	refreshSum.Write([]byte(refreshCookie.Value))
	refreshTokenSum := fmt.Sprintf("%x", refreshSum.Sum(nil))

	// Fetch the current token from the database = check if exists.
	token, err := s.tokenRepository.GetByID(refreshTokenSum)
	if err != nil {
		return fmt.Errorf(common.ERR_INVALID_REF_TOKEN)
	}

	// Delete such token not to be prune to hijack anymore.
	err = s.tokenRepository.Delete(token.Hash)
	if err != nil {
		return err
	}

	return nil
}
