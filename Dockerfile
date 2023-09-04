#
# litter-go / Dockerfile
#


#
# stage 0 -- build
#

# https://hub.docker.com/_/golang
ARG GOLANG_VERSION
FROM golang:${GOLANG_VERSION}-alpine3.18 AS litter-build

ARG APP_NAME

ENV APP_NAME ${APP_NAME}
ENV CGO_ENABLED 1

RUN apk add git gcc

WORKDIR /go/src/${APP_NAME}
COPY . .

# build the client -- wasm binary
RUN GOARCH=wasm GOOS=js go build -o web/app.wasm -tags wasm

# build the server
RUN go build ${APP_NAME}


#
# stage 1 -- release
#

FROM alpine:3.18 AS litter-release

ARG APP_FLAGS
ARG APP_PEPPER
ARG APP_VERSION
ARG DOCKER_INTERNAL_PORT
ARG DOCKER_USER

ENV APP_FLAGS ${APP_FLAGS}
ENV APP_PEPPER ${APP_PEPPER}
ENV APP_VERSION ${APP_VERSION}
ENV DOCKER_DEV_PORT ${DOCKER_INTERNAL_PORT}
ENV DOCKER_USER ${DOCKER_USER}

COPY web/ /opt/web/
COPY data/ /opt/data/
COPY --from=litter-build /go/src/litter-go/litter-go /opt/litter-go
COPY --from=litter-build /go/src/litter-go/web/app.wasm /opt/web/app.wasm

WORKDIR /opt
#USER ${DOCKER_USER}
EXPOSE ${DOCKER_INTERNAL_PORT}
ENTRYPOINT ["/opt/litter-go"]


