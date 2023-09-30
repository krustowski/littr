# litter-go (littr)

[![Go Reference](https://pkg.go.dev/badge/go.savla.dev/littr.svg)](https://pkg.go.dev/go.savla.dev/littr)
[![Go Report Card](https://goreportcard.com/badge/go.savla.dev/littr)](https://goreportcard.com/report/go.savla.dev/littr)

litter again, now in Go as a PWA --- a microblogging service without notifications and pesky messaging, just a raw mind _flow_

[read more](https://krusty.space/projects/litter/) (a blog post)

## repo vademecum

`backend/`
+ files related to REST API backend service, this API server is used by WASM client for fetching of app's data

`config/`
+ configuration procedures for various litter module packages

`data/`
+ sample data files used to flush existing container data by `make flush`

`frontend/`
+ app pages' files sorted by their name(s)

`models/`
+ various model declarations

`web/`
+ static web files, logos, manifest

`.env` + `.env.example`
+ environmental contants/vars for the app to run smoothly (in Docker)

`http_server.go`
+ init app file for the app's backend side with REST API service

`wasm_client.go`
+ lightened version of HTTP server, includes basic app router, lacks REST API service

## how it should work
+ each user has to register (`/register`) or existed users login (`/login`)
+ users can navigate to the flow (`/flow`) to read other's mind _flows_
+ users can change their passphrase or the _about_ description in settings (`/settings`)
+ posts can be written and sent (`/post`) by any logged in user
+ users can logout (`/logout`)

## features

+ togglable end-to-end encryption (JSON/octet-stream REST API and LocalStorage for service worker) 
+ in-memory runtime cache
+ data persistence on container restart (on `SIGINT`)
+ flow posts filtering using the FlowList --- simply choose whose posts to show
+ can run offline (using the cache)

## REST API service
+ the service is reachable via (`/api`) endpoint (~~API auth-wall to be implemented~~)
+ there are five main endpoints: 

```http
POST   /api/auth

GET    /api/dump

POST   /api/flow
GET    /api/flow
PUT    /api/flow
DELETE /api/flow/:key

POST   /api/polls
GET    /api/polls
PUT    /api/polls
DELETE /api/polls/:key

POST   /api/users
GET    /api/users
PUT    /api/users
DELETE /api/users/:key
```

## how to run

```bash
# create env file copy and modify it
cp .env.example .env
vi .env

# build docker image (Docker engine mandatory)
make build

# run docker-compose stack, start up the stack
make run

# Makefile helper print
make info

# flush app data --- copy empty files from 'data/' to the container
make flush kill run
```

## development

`litter-backend` container can be run locally on any dev machine (using Docker engine or using the required tag-locked Go runtime)

```
make fmt build run

http://localhost:8093/flow
```

### nice-to-have(s)
+ implement loading new posts
+ check for double posting same content
+ implement infinite flow scroll
+ implement sessions (SessionsCache)
+ implement custom logger goroutine
+ ~~healthcheck as a periodic cache dumper~~
+ allow emojis in usernames? NO! 😂
+ show offline mode notification
+ swagger docs
+ deep code refactoring
+ replies to posts (just one level deep iteration, no reply to a reply)
+ implement user search on `users` and `stats` pages
+ Ctrl+Enter to submit posts
+ image uploads to gspace.gscloud.cz 😎
+ autosubmit on password manager autofill
+ break lines on \n in posts
+ automatic links for known TLDs
+ user's real name
+ user's home page link

### roadmap to v0.9 *OUTDATED*
+ ~~implement polls (create poll, voting)~~

### roadmap to v0.8
+ ~~implement stats~~
+ ~~ensure system user unregisterable by the app logic (not hardcoded)~~
  - ~~define data/users.json system users (regular users, but simply taken already by the init time)~~
+ ~~implement picture posting~~ (via URL)

### readmap to v0.7
+ ~~implement functional flowList editting via UI~~
+ ~~implement functional filtering using flowList~~
+ ~~fix dynamic stargazing and post deletion~~

### roadmap to v0.6
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

