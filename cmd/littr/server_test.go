package main

import (
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	s := newServer()

	go s.Run()

	time.Sleep(5 * time.Second)
}
