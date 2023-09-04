package config

// those vars are mainly used by frontend, their values are compiled in during docker build from the environment vars based on the build machine environment
var (
	APIToken string
	Pepper   string
	Version  string
)
