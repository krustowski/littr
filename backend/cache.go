package backend

import (
	"go.savla.dev/swis-nexus/pkg/nexus"
)

func databaseCall() {
	client := nexus.Client{
		BaseURL: "http://litter-swis:8050",
		Token: "xxx",
	}

	client.Read("/users")
}
