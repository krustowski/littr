package config

import (
	"os"
)

var (
	BackendToken string
)

func ParseEnv() {
	BackendToken = os.Getenv("API_TOKEN")
}
