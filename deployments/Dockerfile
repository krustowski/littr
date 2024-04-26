# syntax=docker/dockerfile:1.4
#
# litter-go / Dockerfile
#
 

#
# stage 0 -- build
#

# https://hub.docker.com/_/golang
ARG GOLANG_VERSION
FROM golang:${GOLANG_VERSION}-alpine3.19 AS litter-build

ARG APP_NAME APP_PEPPER APP_VERSION API_TOKEN VAPID_PUB_KEY

ENV APP_NAME ${APP_NAME}
ENV APP_PEPPER ${APP_PEPPER}
ENV APP_VERSION ${APP_VERSION}
ENV API_TOKEN ${API_TOKEN}
ENV CGO_ENABLED 0
ENV VAPID_PUB_KEY ${VAPID_PUB_KEY}

RUN --mount=type=cache,target=/var/cache/apk \
	apk add git gcc build-base libc-dev

WORKDIR /go/src/${APP_NAME}
COPY go.mod .

ARG GOMODCACHE GOCACHE
RUN --mount=type=cache,target="$GOMODCACHE" go mod download

#COPY .. . 
COPY cmd/littr /usr/local/go/src/cmd/littr
COPY . /go/src/${APP_NAME}/

# build the client (wasm binary)
RUN --mount=type=cache,target="$GOMODCACHE" \
	--mount=type=cache,target="$GOCACHE" \
	GOARCH=wasm GOOS=js go build \
		-o web/app.wasm \
		-tags wasm \
		-ldflags "-X 'go.savla.dev/littr/frontend.appVersion=$APP_VERSION' -X 'go.savla.dev/littr/frontend.vapidPublicKey=$VAPID_PUB_KEY'"\
		cmd/littr/

# build the server (go binary)
RUN --mount=type=cache,target="$GOMODCACHE" \
	--mount=type=cache,target="$GOCACHE" \
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-o littr \
		cmd/littr/


#
# stage 1 -- release
#

FROM alpine:3.19 AS litter-release

ARG APP_FLAGS APP_PEPPER APP_VERSION DOCKER_INTERNAL_PORT DOCKER_USER

ENV APP_FLAGS ${APP_FLAGS}
ENV APP_PEPPER ${APP_PEPPER}
ENV APP_VERSION ${APP_VERSION}
ENV DOCKER_DEV_PORT ${DOCKER_INTERNAL_PORT}
ENV DOCKER_USER ${DOCKER_USER}

RUN --mount=type=cache,target=/var/cache/apk \
	apk update && apk add tzdata

RUN adduser -D -h /opt -s /bin/sh ${DOCKER_USER}

COPY web/ /opt/web/
COPY --chown=1000:1000 --chmod=700 .data/ /opt/data/
COPY --chown=1000:1000 --chmod=700 .data/.gitkeep /opt/pix/

COPY --from=litter-build /go/src/litter-go/littr /opt/littr
COPY --from=litter-build /go/src/litter-go/web/app.wasm /opt/web/app.wasm

# workaround for pix
RUN cd /opt/web && ln -s ../pix .
RUN ln -s /opt/littr /usr/local/bin
RUN chown -R ${DOCKER_USER}:${DOCKER_USER} /opt/

WORKDIR /opt
USER ${DOCKER_USER}
EXPOSE ${DOCKER_INTERNAL_PORT}
ENTRYPOINT ["littr"]
