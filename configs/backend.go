package configs

import (
	"os"
)

var (
	BackendToken string
)

func ParseEnv() {
	BackendToken = os.Getenv("API_TOKEN")
}

/*
 *  BE data migrations
 */

var UserDeletionList []string = []string{
	"fred",
	"fred2",
	"admin",
	"alternative",
	"Lmao",
	"lma0",
}

var UsersToUnshade []string = []string{}
