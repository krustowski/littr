# litter-go
litter again, now in Go as PWA --- microblogging service without notifications and pesky messaging, just a raw mind _flow_

## repo vademecum

`backend/`
+ files related to REST API backend service, this API server is used by WASM client for fetching of app's data

`data/`
+ sample data files used to flush existing container data by `make flush`

`pages/`
+ app pages' files sorted by their name(s)

`web/`
+ static web files, logos, manifest

`.env`
+ environmental contants/vars for the app to run smoothly

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
  + `/api/auth`
  + `/api/flow`
  + `/api/users`

## development

### roadmap to v0.3
+ ~~use local JSON storage~~
+ ~~implement `backend.authUser`~~
+ ~~functional user login/logout~~
+ functional settings page
+ functional add/remove flow user

### roadmap to v0.2
+ ~~Go backend (BE) --- server-side~~
+ ~~connect frontend (FE) to BE~~
+ ~~application logic --- functional pages' inputs, buttons, lists (flow, users)~~

