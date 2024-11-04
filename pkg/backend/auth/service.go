package auth

import (
	"context"
	"fmt"
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
	return fmt.Errorf("not yet implemented")
}
