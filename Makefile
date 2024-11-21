#
#  littr / Makefile
#

#
#  VARS
#

include .env.example
-include .env

APP_ENVIRONMENT ?= dev
APP_NAME=littr
APP_URLS_TRAEFIK ?= `${HOSTNAME}`
APP_URL_MAIN ?= ${HOSTNAME}
PROJECT_NAME=${APP_NAME}-${APP_ENVIRONMENT}
TZ ?= Europe/Vienna

LOKI_URL ?=

APP_PEPPER ?=
API_TOKEN ?=

LIMITER_DISABLED ?= false
REGISTRATION_ENABLED ?= true

VAPID_PUB_KEY ?=
VAPID_PRIV_KEY ?=
VAPID_SUBSCRIBER ?=

MAIL_HELO ?= localhost
MAIL_HOST ?=
MAIL_PORT ?= 25
MAIL_SASL_USR ?=
MAIL_SASL_PWD ?=

# go environment
#COMMON_BUILD_LDFLAGS=-s -w
COMMON_BUILD_LDFLAGS=
GOARCH := $(shell go env GOARCH)
GOCACHE ?= /home/${USER}/.cache/go-build
GOMODCACHE ?= /home/${USER}/go/pkg/mod
GOOS := $(shell go env GOOS)

# go build -race [...]
RACE_FLAG ?= ""

# docker environment
DOCKER_COMPOSE_FILE ?= deployments/docker-compose.yml
DOCKER_COMPOSE_TEST_FILE ?= deployments/docker-compose-test.yml
DOCKER_COMPOSE_OVERRIDE ?= deployments/docker-compose.override.yml
DOCKER_COMPOSE_TEST_OVERRIDE ?= deployments/docker-compose-test.override.yml
DOCKER_CONTAINER_NAME ?= ${PROJECT_NAME}-server

REGISTRY ?= ${APP_NAME}
DOCKER_BUILD_IMAGE ?= golang:${GOLANG_VERSION}-alpine
DOCKER_BUILD_IMAGE_RELEASE ?= alpine:${ALPINE_VERSION}
DOCKER_IMAGE_TAG ?= ${REGISTRY}/littr/backend:${APP_VERSION}-go${GOLANG_VERSION}

DOCKER_INTERNAL_PORT ?= 8080
DOCKER_NETWORK_NAME ?= traefik
DOCKER_USER ?= littr
DOCKER_VOLUME_DATA_NAME ?= littr-data
DOCKER_VOLUME_PIX_NAME ?= littr-pix

# define standard colors
# https://gist.github.com/rsperl/d2dfe88a520968fbc1f49db0a29345b9
ifneq (,$(findstring xterm,${TERM}))
	BLACK        := $(shell tput -Txterm setaf 0)
	RED          := $(shell tput -Txterm setaf 1)
	GREEN        := $(shell tput -Txterm setaf 2)
	YELLOW       := $(shell tput -Txterm setaf 3)
	LIGHTPURPLE  := $(shell tput -Txterm setaf 4)
	PURPLE       := $(shell tput -Txterm setaf 5)
	BLUE         := $(shell tput -Txterm setaf 6)
	WHITE        := $(shell tput -Txterm setaf 7)
	RESET        := $(shell tput -Txterm sgr0)
else
	BLACK        := ""
	RED          := ""
	GREEN        := ""
	YELLOW       := ""
	LIGHTPURPLE  := ""
	PURPLE       := ""
	BLUE         := ""
	WHITE        := ""
	RESET        := ""
endif

export


#
#  FUNCTIONS
#

define print_info
	@echo -e "\n>>> ${YELLOW}${1}${RESET}\n"
endef


#
#  TARGETS
#

.PHONY: all info

all: info

info: 
	@echo -e "\n${GREEN} ${PROJECT_NAME} / Makefile ${RESET}\n"

	@echo -e "${YELLOW} make config  --- check dev environment ${RESET}"
	@echo -e "${YELLOW} make fmt     --- reformat the go source (gofmt) ${RESET}"
	@echo -e "${YELLOW} make docs    --- render documentation from code (go doc) ${RESET}\n"

	@echo -e "${YELLOW} make build   --- build project (docker image) ${RESET}"
	@echo -e "${YELLOW} make run     --- run project ${RESET}"
	@echo -e "${YELLOW} make logs    --- fetch container's logs ${RESET}"
	@echo -e "${YELLOW} make stop    --- stop and purge project (only docker containers!) ${RESET}"
	@echo -e ""

#
#  deployment targets
#

.PHONY: dev prod

dev: version fmt build run logs

prod: run logs


#
#  development targets
#

.PHONY: fmt config push push_mirror sonar_scan test_local test_local_coverage

deps:
	$(call print_info, Fetching/upgrading dependencies...)
	@go get -u ./...

fmt:
	$(call print_info, Reformatting the code using gofmt tool...)
	@gofmt -w -s .

config:
	$(call print_info, Running the local environment configuration setup...)
	@go install github.com/swaggo/swag/cmd/swag@latest

push:
	$(call print_info, Pushing tagged commits to origin/master...)
	@git tag -fa 'v${APP_VERSION}' -m 'v${APP_VERSION}'
	@git push --follow-tags --set-upstream origin master
	
push_mirror:
	$(call print_info, Pushing tagged commits to mirror/master...)
	@git push --follow-tags mirror master
	
ifeq (${SONAR_URL}${SONAR_PROJECT_TOKEN},)
sonar_scan:
else
sonar_scan:
	$(call print_info, Starting the sonarqube code analysis...)
	sonar-scanner \
		-Dsonar.projectKey=${APP_NAME} \
		-Dsonar.sources=. \
		-Dsonar.host.url=${SONAR_URL}   \
		-Dsonar.login=${SONAR_PROJECT_TOKEN}
endif

test_local: fmt
	$(call print_info, Running Go unit/integration tests locally...)
	@go clean -testcache
	@go test $(shell go list ./... | grep -v cmd/sse_client | grep -v cmd/dbench | grep -v pkg/models | grep -v pkg/helpers | grep -v pkg/frontend)

test_local_coverage: fmt
	$(call print_info, Running Go code coverage analysis...)
	@go clean -testcache
	@go test -v -coverprofile coverage.out ./... && \
		go tool cover -html coverage.out


#
#  Versioning (semver incrementing) targets
#

.PHONY: major minor patch version

define update_semver
	@echo "Incrementing semver to ${1}"
	@sed -i 's|APP_VERSION=.*|APP_VERSION=${1}|' .env
	@sed -i 's|APP_VERSION=.*|APP_VERSION=${1}|' .env.example
	@sed -i 's/\/\/\(.*[[:blank:]]\)[0-9]*\.[0-9]*\.[0-9]*/\/\/\1${1}/' pkg/backend/router.go
endef

MAJOR := $(shell echo ${APP_VERSION} | cut -d. -f1)
MINOR := $(shell echo ${APP_VERSION} | cut -d. -f2)
PATCH := $(shell echo ${APP_VERSION} | cut -d. -f3)

major:
	$(eval APP_VERSION := $(shell echo $$(( ${MAJOR} + 1 )).0.0))
	$(call update_semver,${APP_VERSION})

minor:
	$(eval APP_VERSION := $(shell echo ${MAJOR}.$$(( ${MINOR} + 1 )).0))
	$(call update_semver,${APP_VERSION})

patch:
	$(eval APP_VERSION := $(shell echo ${MAJOR}.${MINOR}.$$(( ${PATCH} + 1 ))))
	$(call update_semver,${APP_VERSION})

version:
	$(call print_info, Current version: ${APP_VERSION}...)


#
#  build&run targets (CI mostly)
#

.PHONY: build docs docs_host push_to_registry run run_test_env

build: 
	@echo -e "\n${YELLOW} Building the project (docker compose build)... ${RESET}\n"
	@[ -f ".env" ] || cp .env.example .env
	@[ -f "${DOCKER_COMPOSE_OVERRIDE}" ] || touch ${DOCKER_COMPOSE_OVERRIDE}
	@if [ \( -z "${REGISTRY_USER}" \) -o \( -z "${REGISTRY_PASSWORD}" \) ]; \
	then \
		DOCKERFILE=full.dockerfile DOCKER_BUILDKIT=1 docker compose -f ${DOCKER_COMPOSE_FILE} -f ${DOCKER_COMPOSE_OVERRIDE} build; \
	else \
		echo "${REGISTRY_PASSWORD}" | docker login -u "${REGISTRY_USER}" --password-stdin "${REGISTRY}"; \
		DOCKERFILE=prebuilt.dockerfile DOCKER_BUILDKIT=1 docker compose -f ${DOCKER_COMPOSE_FILE} -f ${DOCKER_COMPOSE_OVERRIDE} build; \
	fi

docs: config
	@echo -e "\n${YELLOW} Generating OpenAPI documentation... ${RESET}\n"
	@~/go/bin/swag init --parseDependency -ot json -g router.go --dir pkg/backend/ 
	@mv docs/swagger.json api/swagger.json
	@[ -f ".env" ] || cp .env.example .env
	@[ -f ${DOCKER_COMPOSE_OVERRIDE} ] \
		&& docker compose -f ${DOCKER_COMPOSE_FILE} -f ${DOCKER_COMPOSE_OVERRIDE} up littr-swagger -d --force-recreate \
		|| docker compose -f ${DOCKER_COMPOSE_FILE} up littr-swagger -d --force-recreate

docs_host:
	@echo -e "\n${YELLOW} Updating the host for docs... ${RESET}\n"
	sed -i 's/\/\/.*\(@host[[:blank:]]*\)[a-z.0-9]*/\/\/ \1${APP_URL_MAIN}/' pkg/backend/router.go

push_to_registry:
	@echo -e "\n${YELLOW} Pushing new image to registry... ${RESET}\n"
	@[ -n "${REGISTRY}" ] || exit 10
	@echo "${REGISTRY_PASSWORD}" | docker login -u "${REGISTRY_USER}" --password-stdin "${REGISTRY}" && \
		docker push ${DOCKER_IMAGE_TAG}
	@docker logout ${REGISTRY} > /dev/null

run:	
	@echo -e "\n${YELLOW} Starting project (docker compose up)... ${RESET}\n"
	@[ -f ".env" ] || cp .env.example .env
	@[ -f "${DOCKER_COMPOSE_OVERRIDE}" ] || touch ${DOCKER_COMPOSE_OVERRIDE}
	@if [ \( -n "${REGISTRY_USER}" \) -a \( -n "${REGISTRY_PASSWORD}" \) ]; \
	then \
		echo "${REGISTRY_PASSWORD}" | docker login -u "${REGISTRY_USER}" --password-stdin "${REGISTRY}"; \
		docker compose -f ${DOCKER_COMPOSE_FILE} -f ${DOCKER_COMPOSE_OVERRIDE} up --force-recreate --detach --remove-orphans; \
		docker logout "${REGISTRY}" > /dev/null; \
	else \
		docker compose -f ${DOCKER_COMPOSE_FILE} -f ${DOCKER_COMPOSE_OVERRIDE} up --force-recreate --detach --remove-orphans; \
	fi

run_test_env:	
	@echo -e "\n${YELLOW} Starting test project (docker compose up)... ${RESET}\n"
	@[ -f ".env" ] || cp .env.example .env
	@[ -f ${DOCKER_COMPOSE_TEST_OVERRIDE} ] \
		&& docker compose -f ${DOCKER_COMPOSE_TEST_FILE} -f ${DOCKER_COMPOSE_TEST_OVERRIDE} up --force-recreate --detach --remove-orphans \
		|| docker compose -f ${DOCKER_COMPOSE_TEST_FILE} up --force-recreate --detach --remove-orphans

#
#  profiling targets
#

.PHONY: kill_proff run_proff

GO_TOOL_PPROF = go tool pprof
NOL := $(shell ps auxf | grep -w '${GO_TOOL_PPROF}' | wc -l | cut -d' ' -f1)
LIST = $(shell ps auxf | grep -w '${GO_TOOL_PPROF}' | tail -n $$(( $(NOL) - 2 )) | awk '{ print $$2 }')
kill_pprof:
	@echo -e "\n${YELLOW} Killing all profiling instances... ${RESET}\n"
	@for INST in ${LIST}; do kill $${INST}; done

PPROF_SOURCE ?= http://localhost:${DOCKER_INTERNAL_PORT}/debug/pprof
run_pprof: kill_pprof
	@echo -e "\n${YELLOW} Starting profiling instances (run kill_pprof to kill them)... ${RESET}\n"
	@go build -pkgdir pkg/ -o littr cmd/littr/main.go cmd/littr/http_server.go cmd/littr/init_client.go
	@ mkdir -p .tmp
	@curl -sL ${PPROF_SOURCE}/allocs?debug=1 > .tmp/alloc.out
	@curl -sL ${PPROF_SOURCE}/goroutine?debug=1 > .tmp/goroutine.out
	@curl -sL ${PPROF_SOURCE}/heap?debug=1 > .tmp/heap.out
	@go tool pprof -http=127.0.0.1:8081 littr .tmp/alloc.out &
	@go tool pprof -http=127.0.0.1:8082 littr .tmp/goroutine.out &
	@go tool pprof -http=127.0.0.1:8083 littr .tmp/heap.out &
	

#
#  runtime (live system operation) targets
#

.PHONY: backup fetch_running_dump flush kill logs sh sse_client stop

backup: fetch_running_dump
	@echo -e "\n${YELLOW} Making the backup archive... ${RESET}\n"
	@tar czvf /mnt/backup/littr/$(shell date +"%Y-%m-%d-%H:%M:%S").tar.gz ${RUN_DATA_DIR}

RUN_DATA_DIR=./.run_data
fetch_running_dump:
	@echo -e "\n${YELLOW} Copying dumped data from the container... ${RESET}\n"
	@mkdir -p ${RUN_DATA_DIR}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/polls.json ${RUN_DATA_DIR}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/posts.json ${RUN_DATA_DIR}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/requests.json ${RUN_DATA_DIR}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/subscriptions.json ${RUN_DATA_DIR}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/tokens.json ${RUN_DATA_DIR}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/users.json ${RUN_DATA_DIR}
	
flush:
	@echo -e "\n${YELLOW} Flushing app data... ${RESET}\n"
	@docker cp test/data/polls.json ${DOCKER_CONTAINER_NAME}:/opt/data/polls.json
	@docker cp test/data/posts.json ${DOCKER_CONTAINER_NAME}:/opt/data/posts.json
	@docker cp test/data/users.json ${DOCKER_CONTAINER_NAME}:/opt/data/users.json
	@docker cp test/data/subscriptions.json ${DOCKER_CONTAINER_NAME}:/opt/data/subscriptions.json
	@docker cp test/data/tokens.json ${DOCKER_CONTAINER_NAME}:/opt/data/tokens.json

kill:
	@echo -e "\n${YELLOW} Killing the container not to dump running caches... ${RESET}\n"
	@[ -f ".env" ] || cp .env.example .env
	@docker kill ${DOCKER_CONTAINER_NAME}

logs:
	@echo -e "\n${YELLOW} Fetching container's logs (CTRL-C to exit)... ${RESET}\n"
	@docker logs ${DOCKER_CONTAINER_NAME} -f

sh:
	@echo -e "\n${YELLOW} Attaching container's (${DOCKER_CONTAINER_NAME})... ${RESET}\n"
	@[ -f ".env" ] || cp .env.example .env
	@docker exec -it ${DOCKER_CONTAINER_NAME} sh

sse_client:
	@echo -e "\n${YELLOW} Starting the SSE client... ${RESET}\n"
	@go run cmd/sse_client/main.go

stop:  
	@echo -e "\n${YELLOW} Stopping and purging project (docker compose down)... ${RESET}\n"
	@[ -f ".env" ] || cp .env.example .env
	@docker compose -f ${DOCKER_COMPOSE_FILE} down

tweak:
	$(call print_info, jezisi kriste dopice uz)
