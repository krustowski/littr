![littr logo](/web/android-chrome-192x192.png)

# littr (formerly litter-go)

[![Go Reference](https://pkg.go.dev/badge/go.vxn.dev/littr.svg)](https://pkg.go.dev/go.vxn.dev/littr)
[![Go Report Card](https://goreportcard.com/badge/go.vxn.dev/littr)](https://goreportcard.com/report/go.vxn.dev/littr)

[![littr CI/CD test and build pipeline](https://github.com/krustowski/littr/actions/workflows/test-and-build.yml/badge.svg?branch=master)](https://github.com/krustowski/littr/actions/workflows/test-and-build.yml)
[![littr CI/CD deployment pipeline](https://github.com/krustowski/littr/actions/workflows/deployment.yml/badge.svg?branch=master)](https://github.com/krustowski/littr/actions/workflows/deployment.yml)

a simple nanoblogging platform for a raw mind _flow_

[read more](https://krusty.space/projects/littr/) (a bit more verbose documentation post)


## features

+ in-memory runtime cache(s)
+ data persistence on container restart (on `SIGINT`) in Docker volumes
+ flow posts filtering using the FlowList --- simply choose who to follow
+ shade function to block other accounts from following you and reading your posts
+ webpush notification management --- choose which notifications (reply/mention) is one's device willing to accept
+ private acccount --- others have to file a follow request to such account (and have to approved by the acc's owner)
+ swift client side (a WebAssebmly binary)
+ safe photo sharing --- EXIF metadata are removed while image file is uploading
+ passphrase reset via e-mail
+ dark/light mode switch
+ live in-app event (SSE) notifications --- get alerted when a new post/poll is added to your flow


## REST API service

+ API documentation is stored in the `api/` root directory as Swagger JSON-formatted file
+ the service is reachable via (`/api/v1`) route
+ [swagger live docs](https://www.littr.eu/docs/)


## how it should work

+ users must register (`/register`) or existed users must login (`/login`)
+ users can navigate to the flow (`/flow`) and read other's mind _flows_
+ users can modify their FlowList (list of followed accounts, `/users`)
+ users can change their passphrase, the _about_ (bio) description and many more (`/settings`)
+ polls/posts can be written and sent (`/post`) by any logged-in user
+ users can logout (`/logout`)


## how to run (docker, to-be-reviewed)

```shell
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

## repo vademecum

`api`
+ swagger documentation

`cmd`
+ main app entrypoints for build

`cmd/littr/http_server.go`
+ init app file for the app's backend side with REST API service

`cmd/littr/wasm_client.go`
+ lightened version of HTTP server, includes basic app router, lacks REST API service

`configs`
+ configuration files
+ nginx proxy full configuration

`deployments`
+ docker deployment files

`.env` + `.env.example`
+ environmental contants and vars for the app to run smoothly (in Docker via Makefile)

`pkg`
+ app libs 

`pkg/backend`
+ REST API backend service
+ service is used by WASM client for the app's running data fetch

`pkg/frontend`
+ frontend pages/views
+ go-app framework usecase

`pkg/models`
+ various shared model declarations

`test/data`
+ sample data used to flush existing container's data by `make flush`

`web`
+ static web files, logos, web manifest


----

## development

`littr-backend` container can be run locally on any dev machine (using Docker engine or using the required tag-locked Go runtime)

```
make dev

http://localhost:8093/flow
```

### investigating the runtime data

```
make sh

ls -laR data/
```

```
data/:
total 44
drwxr-xr-x    1 littr    littr           96 Nov 14 22:21 .
drwxr-sr-x    1 littr    littr           24 Dec 17 20:58 ..
-rw-r--r--    1 littr    littr         2562 Dec 17 22:02 polls.json
-rw-r--r--    1 littr    littr        24309 Dec 17 22:02 posts.json
-rw-r-----    1 littr    littr         1154 Dec 17 22:02 subscriptions.json
-rw-r--r--    1 littr    littr        11923 Dec 17 22:02 users.json
```

In the listing above we can see freshly dumped data from littr runtime. We can fetch it outside the container using `make fetch_run_data`.

```
make fetch_running_dump

 Copying dumped data from the container...

Successfully copied 13.8kB to /home/user/littr/run_data/
Successfully copied 4.61kB to /home/user/littr/run_data/
Successfully copied 26.1kB to /home/user/littr/run_data/
Successfully copied 3.07kB to /home/user/littr/run_data/
```

### nice-to-have(s)

+ ~~account deletion (`settings` page)~~
+ ~~add timestamps on server, render them on client side~~
+ ~~automatic links for known TLDs~~
+ ~~autosubmit on password manager autofill (when username and password get filled quite quickly)~~
+ ~~break lines on \n in posts~~
+ ~~check for double posting same content~~
+ custom colour theme per user
+ dismiss any modal by clicking anywhere on the screen
+ ~~fix update indicator checker (runs only on once after reload)~~ (autoupdate)
+ implement customizable navbar items order
+ ~~show offline mode notification~~
+ ~~show user's details on the top of /flow/<username> page~~
+ ~~swagger docs~~
+ ~~test if dump dir writable (on init)~~ (chown included in Dockerfile)

### roadmap to v1.0.0
+ fix the code smells as scanned by `sonar-scanner`
+ resolve tikets in github and redmine (priv)
+ write integration and e2e tests
+ [...]

### roadmap to v0.42
+ user activation via mail

### roadmap to v0.41
+ ~~pagination for polls~~
+ refactor FE (wip) --- polls, users, settings, register, reset, tos, welcome, login, flow, post
+ single poll referenced by ID (singlePoll subpage/view)
+ ~~universal paginator on BE~~

### roadmap to v0.40
+ ~~convert GIFs to WebPs~~
+ ~~e-mail duplicity check for registration~~
+ ~~introduce the hideReplies feature~~

### roadmap to v0.39
+ ~~fix avatar image uploading, resizing and cropping~~
+ ~~implement Esc key to close toasts and modals + Ctrl-Enter for other inputs~~
+ ~~welcome page~~

### roadmap to v0.38
+ ~~add minor UI fixes (dialog buttons colouring)~~
+ ~~expand passphrase reset procedure --- add confirmation e-mail~~

### roadmap to v0.37
+ ~~improve SSE parsing on FE~~
+ ~~show server is restarting notice~~

### roadmap to v0.36
+ ~~implement localTime mode switch to see post in one's timezone (by default on)~~
+ ~~improve user's flow (profiles) --- viewable in limits of following/shading/private acc~~

### roadmap to v0.35
+ ~~implement searching for flow using hashtags~~ (wip)
+ ~~fix various bugs and typos (login, register pages; texts)~~ (wip)

### roadmap to v0.34
+ ~~fix the flow-users glitch (new users' posts not seen in flow on other devices)~~
+ ~~implement private account logic~~ (wip)
+ ~~allow picture-only posting~~

### roadmap to v0.33
+ ~~BE deep code refactoring~~ (wip)
+ ~~write and fix swagger docs~~

### roadmap to v0.32
+ ~~implement mentions with notification to such user~~

### roadmap to v0.31
+ ~~implement subscribtion list item deletion~~
+ improve the UI/UX (review issues related to UI on Github) (wip)

### roadmap to v0.30
+ ~~implement Ctrl+Enter to submit posts like YouTube~~
+ ~~unify the UI elements' border radius and their styling~~

### roadmap to v0.29
+ ~~add timestamp to a subscription~~
+ ~~list subscribed devides on settings page~~

### roadmap to v0.28
+ ~~enhance the notification service~~

### roadmap to v0.27
+ implement mailing (verification mails) on backend
+ ~~implement forgotten password recovery~~ (wip)

### roadmap to v0.26
+ ~~implement combined picture-with-text (PwT) posting~~ (wip)

### roadmap to v0.25
+ ~~implement simple loading/broadcasting of new posts~~ (wip)
+ ~~fix single-post and user flow subpages on flow~~ (wip)
### roadmap to v0.24
+ ~~add modal for post deletion confirmation~~
+ ~~implement JWT for auth (wip)~~
+ ~~preprocess and paginate posts on backend (wip)~~

### roadmap to v0.23
+ ~~reintroduce `/api/stats` route to simplify the stats page~~
+ ~~use a router (Gorilla Mux, Go-Chi)~~

### roadmap to v0.22
+ ~~fix flow reorganize glitch (single page, after post deletion etc) (wip)~~ (fixed in v0.30.29)
+ ~~send frontend's tagged version to backend (improve user debugging)~~

### roadmap to v0.21
+ ~~implement simple notification service (wip)~~
+ ~~implement subscriptions (SubscriptionCache)~~
+ ~~show user's profile on top of the flow single page view~~

### roadmap to v0.20
+ ~~fix thumbnail multiple-loading on scroll~~
+ ~~implement image uploading directly from browser~~

### roadmap to v0.19
+ ~~add avatar preview to settings page~~
+ ~~fix local JS and CSS deps (hotfix via remote CDN)~~
+ ~~fix NaN% when loading~~
+ ~~implement custom JSON log wrapper~~ (~~goroutine?~~)
+ ~~implement `<details>` tag for posts longer than `MaxPostLength)~~
+ ~~implement shadow function (acc blocking)~~
+ ~~implement some user's stats on users page~~
+ ~~show selected user's posts only --- button/link on users page~~

### roadmap to v0.18
+ ~~add links to a single flow post~~
+ ~~show replies (+history) in the single post view~~

### roadmap to v0.17
+ ~~implement user deletion (settings page)~~

### roadmap to v0.16
+ ~~add switch for light/dark mode toggle~~
+ ~~profile icons on the flow page~~ (with link to user info modal???)
+ ~~fix register date (migration)~~

### roadmap to v0.15
+ implement inf. scroll to stats (complex)???
+ ~~replies to any post in flow~~
+ ~~add simple ToS (terms of service)~~

### roadmap to v0.14
+ ~~ensure infinite scroll works properly with flow filtering~~
+ ~~disable update button, autoreload on update instead~~
+ ~~include infinite scroll to `polls` and `users`~~

### roadmap to v0.13
+ ~~implement infinite scroll (for `flow` only this time)~~
+ ~~implement pagination for flow (on backend prolly)~~

### roadmap to v0.12
+ ~~implement user's profile (real name, home page, followers) as modal~~
+ ~~mark selected fields with data labels (`data-author` etc)~~
+ ~~notice user's activity (`User.LastActiveTime`)~~

### roadmap to v0.11
+ ~~upgrade toasts to snackbars (beercss v3.3.5)~~
+ ~~disable buttons on action to prevent multiple action fires~~
+ ~~implement system stats, add flowers (follower) count~~
+ ~~add littrJS external lib~~

### roadmap to v0.10
+ ~~implement search bar (`stats` and `users` pages only)~~

### roadmap to v0.9
+ ~~implement polls (create poll, voting)~~
+ ~~healthcheck as a periodic cache dumper~~

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
