# litter-go (littr)

[![Go Reference](https://pkg.go.dev/badge/go.savla.dev/littr.svg)](https://pkg.go.dev/go.savla.dev/littr)
[![Go Report Card](https://goreportcard.com/badge/go.savla.dev/littr)](https://goreportcard.com/report/go.savla.dev/littr)

litter again, now in Go as a PWA --- a microblogging service without notifications and pesky messaging, just a raw mind _flow_

[read more](https://krusty.space/projects/litter/) (a blog post)

## repo vademecum

`backend/`
+ files related to REST API backend service, this API server is used by WASM client for fetching of app's data

`config`
+ configuration procedures for various litter module packages

`data/`
+ sample data files used to flush existing container data by `make flush`

`frontend/`
+ app pages' files sorted by their name(s)

`models`
+ various model declarations

`web/`
+ static web files, logos, manifest

`.env`/`.env.example`
+ environmental contants/vars for the app to run smoothly (in Docker)

`http_server.go`
+ init app file for the app's backend side with REST API service

`wasm_client.go`
+ lightened version of HTTP server, includes basic app router, lacks REST API service

## how it should work
+ each user has to register (`/register`) and login (`/login`) to app for them to use it
+ user is then logged-in and navigated to the flow (`/flow`) page when one can read other's mind _flows_
+ user can change their passphrase or the _about_ description on the settings (`/settings`) page
+ any signed (by logged user) post can be written and sent by the post (`/post`) page
+ user can logout too by navigating themselves to the logout (`/logout`) page 

## REST API service
+ the service is reachable via (`/api`) endpoint --- API auth-wall to be implemented
+ there are three main endpoints: 

```http
POST /api/auth

POST   /api/flow
GET    /api/flow
PUT    /api/flow
DELETE /api/flow/:key

POST /api/users
GET  /api/users
PUT  /api/users
```

## how to run

```bash
# create environmental file copy and modify it
cp .env.example .env
vi .env

# build docker image (Docker engine is needed)
make build

# run docker-compose stack, start up the stack
make run

# Makefile helper print
make info

# flush app data --- copy empty files from 'data/' to the container
make flush kill run
```

## development

litter-server container can be run locally on any dev machine (with Docker engine, or with the required Go runtime)

```
make fmt build run

http://localhost:8093/flow
```

### nice-to-have(s)
+ implement infinite scroll for flow
+ implement sessions (SessionsCache)
+ implement custom logger goroutine

### roadmap to v0.7
+ implement polls (create poll, voting)
+ implement stats 

### roadmap to v0.6
+ healthcheck as a periodic cache dumper
+ ~~fix settings (broken pwd and about text change)~~
+ ~~fix register when given user exists already~~

### roadmap to v0.4/v0.5
+ ~~stabilize database~~ (implement in-memory caches)
+ ~~fix layout (unwanted less-data page scrolling~~
+ ~~dump data on container restart/SIGINT~~

### roadmap to v0.3
+ ~~use local JSON storage~~
+ ~~implement `backend.authUser`~~
+ ~~functional user login/logout~~
+ ~~functional settings page~~
+ ~~functional add/remove flow user~~

### roadmap to v0.2
+ ~~Go backend (BE) --- server-side~~
+ ~~connect frontend (FE) to BE~~
+ ~~application logic --- functional pages' inputs, buttons, lists (flow, users)~~

