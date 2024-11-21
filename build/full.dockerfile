# syntax=docker/dockerfile:1.10
#
# littr / Dockerfile
#
 

#
# stage 0 -- build
#

# https://hub.docker.com/_/golang
ARG DOCKER_BUILD_IMAGE="golang:1.23-alpine"
ARG DOCKER_BUILD_IMAGE_RELEASE "alpine:3.20"
FROM ${DOCKER_BUILD_IMAGE} AS littr-build

ARG APP_NAME APP_PEPPER APP_VERSION API_TOKEN DOCKER_USER VAPID_PUB_KEY

ENV APP_NAME ${APP_NAME}
ENV APP_PEPPER ${APP_PEPPER}
ENV APP_VERSION ${APP_VERSION}
ENV API_TOKEN ${API_TOKEN}
ENV CGO_ENABLED 0
ENV VAPID_PUB_KEY ${VAPID_PUB_KEY}

RUN --mount=type=cache,target=/var/cache/apk \
	apk add git gcc build-base libc-dev gzip tzdata

WORKDIR /go/src/${APP_NAME}
COPY go.mod .

ARG COMMON_BUILD_LDFLAGS GOMODCACHE GOCACHE GOARCH
RUN --mount=type=cache,target="$GOMODCACHE" go mod download

COPY cmd/littr /usr/local/go/src/cmd/littr
COPY . /go/src/${APP_NAME}/

# build the client (wasm binary)
RUN --mount=type=cache,target="$GOMODCACHE" \
	--mount=type=cache,target="$GOCACHE" \
	GOARCH=wasm GOOS=js go build \
		-o web/app.wasm \
		-tags wasm \
		-ldflags "${COMMON_BUILD_LDFLAGS} -X 'go.vxn.dev/littr/pkg/frontend/common.AppVersion=$APP_VERSION' -X 'go.vxn.dev/littr/pkg/frontend/common.AppPepper=$APP_PEPPER' -X 'go.vxn.dev/littr/pkg/frontend/common.VapidPublicKey=$VAPID_PUB_KEY'"\
		cmd/littr/

# build the server (go binary)
RUN --mount=type=cache,target="$GOMODCACHE" \
	--mount=type=cache,target="$GOCACHE" \
	CGO_ENABLED=1 GOOS=linux GOARCH=$GOARCH go build \
		-o littr \
		-tags server \
		-ldflags "${COMMON_BUILD_LDFLAGS}" \
		cmd/littr/

RUN gzip -9 web/app.wasm

RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "/opt" \
  --shell "/sbin/sh" \
  --no-create-home \
  --uid 1000 \
  ${DOCKER_USER}


#
# stage 1 -- release
#

FROM ${DOCKER_BUILD_IMAGE_RELEASE} AS littr-release

ARG APP_FLAGS APP_VERSION DOCKER_INTERNAL_PORT DOCKER_USER TZ

ENV APP_FLAGS ${APP_FLAGS}
ENV APP_VERSION ${APP_VERSION}
ENV APP_PORT ${DOCKER_INTERNAL_PORT}
ENV APP_USER ${DOCKER_USER}

COPY web/ /opt/web/
COPY api/swagger.json /opt/web/

RUN echo "${TZ}" > /etc/timezone
COPY --from=littr-build /usr/share/zoneinfo/${TZ} /etc/localtime

COPY --from=littr-build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=littr-build /etc/passwd /etc/passwd
COPY --from=littr-build /etc/group /etc/group

COPY --chown=1000:1000 --chmod=700 test/data/ /opt/data/
COPY --chown=1000:1000 --chmod=700 test/data/.gitkeep /opt/pix/
COPY --chown=1000:1000 --chmod=700 pkg/backend/mail/templates/ /opt/templates/

COPY --from=littr-build /go/src/littr/littr /opt/littr
COPY --from=littr-build /go/src/littr/web/app.wasm.gz /opt/web/app.wasm.gz

# workaround for pix
RUN cd /opt/web && ln -s ../pix .
RUN ln -s /opt/littr /usr/local/bin
RUN chown -R ${DOCKER_USER}:${DOCKER_USER} /opt/

WORKDIR /opt
USER ${DOCKER_USER}
EXPOSE ${DOCKER_INTERNAL_PORT}
ENTRYPOINT ["littr"]
