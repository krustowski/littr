package db

import (
	"go.vxn.dev/littr/pkg/models"
)

//
//  Repositories
//

type Repositories struct {
	PollRepository models.PollRepositoryInterface
	PostRepository models.PostRepositoryInterface
	UserRepository models.UserRepositoryInterface
}

var Storage *Repositories
