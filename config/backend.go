package config

import (
	"os"
)

var (
	BackendToken string
)

func initBackendConfig() {
	BackendToken = os.Getenv("API_TOKEN")
}
