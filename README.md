![littr logo](/web/android-chrome-192x192.png)

# littr nanoblogger

[![Go Reference](https://pkg.go.dev/badge/go.vxn.dev/littr.svg)](https://pkg.go.dev/go.vxn.dev/littr)
[![Go Report Card](https://goreportcard.com/badge/go.vxn.dev/littr)](https://goreportcard.com/report/go.vxn.dev/littr)

A simple nanoblogging platform for a raw mind _flow_.

[Read more](https://krusty.space/projects/littr/) (a bit more verbose documentation post).


## features

+ embedded **in-memory cache** system
+ **data persistence** on container's restart/shutdown (on `SIGINT`) in Docker volume(s)
+ swift client side (a **WebAssebmly** binary)
+ flow **posts filtering** using the FlowList --- simply choose who to follow
+ shade function to block other accounts from following you and reading your posts
+ **webpush notifications** management --- choose which notifications (reply/mention) is your device willing to accept
+ live in-app event (**server-sent events, SSE**) notifications --- get alerted when a new post/poll is added to your flow --- to be toggable soon
+ **private account** --- others have to file a follow request to such account (and have to approved by the acc's owner)
+ safe photo sharing --- **EXIF metadata removed** while an image file is uploading
+ user verification and passphrase reset via e-mail
+ **dark/light mode** switch
+ **keyboard shortcuts** to navigate throughout the site, or to toggle some features (e.g. hide replies using `X`)
+ Gravatar integration


## REST API service

+ an universal interface above the whole backend for the indirect data retrieval and manipulation
+ layered architecture (controller + service + repository layers)
+ JSON response
+ the main and secure data source for the frontend Go/JS/WASM client
+ reachable via the `/api/v1` route prefix
+ API documentation is stored in the `api/` root repo directory as Swagger JSON-formatted file: [Swagger live docs](https://www.littr.eu/docs/)


## how it should work

+ users must **register** (`/register`), or existed users must **login** (`/login`)
+ users can navigate to the **flow** (`/flow`) and read other's mind _flows_
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
+ Swagger documentation

`cmd`
+ various (dev) tooling, and the build entrypoints for both BE and FE

`cmd/littr/http_server.go`
+ REST API server initialization procedure (BE)

`cmd/littr/wasm_client.go`
+ initialization procedure for the client binary (FE), cross-compiled into a WASM binary

`configs`
+ configuration files
+ nginx proxy full configuration

`deployments`
+ docker deployment files

`.env` + `.env.example`
+ environmental contants and vars for the app to run smoothly (in Docker via Makefile)

`pkg`
+ shared repo packages 

`pkg/backend`
+ REST API backend service source
+ service is used by WASM client for the app's running data fetch

`pkg/frontend`
+ frontend pages/views source
+ go-app framework usecase

`pkg/models`
+ various shared model declarations

`test/data`
+ sample data used to flush existing container's data by `make flush`

`web`
+ static web files, logos, web manifest


----

## development

The `littr-backend` container can be run locally on any dev machine (using Docker engine or using the required tag-locked Go runtime; the latter option is not documented though).

```
make dev

http://localhost:8080/flow
```

### profiling

To debug memory allocations, or possible blocking goroutine(s), one may utilize the `pprof` package. (One may need the `graphviz` system package installed).

```
make run_pprof
```

The default debug `pprof`'s index: [http://localhost:8080/debug/pprof/](http://localhost:8080/debug/pprof/)

### investigating the runtime data

Enter the container's shell:

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

### roadmap to v0.46
+ add new modal to show logged user and links to flow and to log out
+ ~~add the user's flow link to logout modal and user info modal~~
+ ~~enhance the UI: consistency (volors, buttons, modals, textboxes)~~, fix texts in toasts
+ live switch implementation (to disable SSE)

### roadmap to v0.45
+ add controler+service+repository experiment for ~~polls~~, ~~users~~, posts (wip)
+ ~~redesign the flow subviews (profile pages)~~
+ ~~reimplement JS fetch() function to omit `http.Client` from FE~~
+ ~~save new post/reply drafts to localStorage~~
+ ~~write integration/e2e tests draft~~ (wip)

### roadmap to v0.44
+ ~~fix the invalid overally metrics states after the backend reload~~
+ ~~improve the API logger (add the request time at least)~~
+ ~~refix the avatar URL migration procedure~~ (wip)
+ ~~reimplement the SSE client in Go~~ (wip)
+ ~~single poll referenced by ID (singlePoll subpage/view]~~
+ ~~user activation via mail~~ (wip)

### roadmap to v0.43
+ ~~design and implement new Token model~~ (wip)
+ ~~enable cache metrics gauges~~
- ~~fix the streamer: event type broadcasting~~
+ ~~refactor mailing on BE~~
+ ~~union options under the umbrella user.Options map~~ (wip)
+ ~~update swagger docs, ensure proper params, inputs, responses etc are valid~~ (wip)

### roadmap to v0.42
+ ~~deeply refactor BE+FE again: declarative syntax for BE logger~~ 
+ ~~implement 5s toast's appearence (autodismiss)~~ (wip)
+ ~~unified clean API response (wip)~~

### roadmap to v0.41
+ ~~pagination for polls~~
+ ~~pagination for users~~
+ ~~refactor FE (wip) --- polls, users, settings, register, reset, tos, welcome, login, flow, post~~
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
+ implement inf. scroll to stats (too complex)???
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
