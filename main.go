// @title		litter-go
// @version	 	0.30.29
// @description	nanoblogging platform as PWA built on go-app framework (PoC)
// @termsOfService	https://littr.n0p.cz/tos

// @contact.name	API Support
// @contact.url		https://littr.n0p.cz/docs/
// @contact.email	littr@n0p.cz

// @license.name	MIT
// @license.url		https://github.com/krustowski/litter-go/blob/master/LICENSE

// @host		littr.n0p.cz
// @BasePath		/api

// @securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url		https://swagger.io/resources/open-api/
package main

func main() {
	// https://github.com/maxence-charriere/go-app/issues/627
	initClient()
	initServer()
}
