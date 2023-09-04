package config

var (
	BackendToken string = ""
)

func initBackendConfig() {
	BackendToken = os.Getenv("API_TOKEN")
}
