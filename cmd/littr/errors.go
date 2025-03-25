package main

import "errors"

var (
	errMissingSecretOrToken = errors.New("server secret string or data dump token must not be blank")
	errServerShutdownFailed = errors.New("HTTP server shutdown failed, trying force shutdown... ")
)
