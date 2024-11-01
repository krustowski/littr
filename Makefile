#
# littr / Makefile
#

#
# VARS
#

include .env.example
-include .env

APP_ENVIRONMENT ?= dev
APP_NAME=littr
APP_URLS_TRAEFIK ?= `${HOSTNAME}`
APP_URL_MAIN ?= ${HOSTNAME}
PROJECT_NAME=${APP_NAME}-${APP_ENVIRONMENT}
TZ=Europe/Vienna

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
GOARCH := $(shell go env GOARCH)
GOCACHE?=/home/${USER}/.cache/go-build
GOMODCACHE?=/home/${USER}/go/pkg/mod
GOOS := $(shell go env GOOS)

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
# TARGETS
#

all: info

.PHONY: info
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
# deployment targets
#

.PHONY: dev
dev: fmt build run logs

.PHONY: prod
prod: run logs


#
# helper targets
#

.PHONY: fmt
fmt: version
	@echo -e "\n${YELLOW} Code reformating (gofmt)... ${RESET}\n"
	@gofmt -w -s .

.PHONY: config
config:
	@echo -e "\n${YELLOW} Running local configuration setup... ${RESET}\n"
	@go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: docs
docs: config
	@echo -e "\n${YELLOW} Generating OpenAPI documentation... ${RESET}\n"
	@~/go/bin/swag init --parseDependency -ot json -g router.go --dir pkg/backend/ 
	@mv docs/swagger.json api/swagger.json
	@[ -f ".env" ] || cp .env.example .env
	@[ -f ${DOCKER_COMPOSE_OVERRIDE} ] \
		&& docker compose -f ${DOCKER_COMPOSE_FILE} -f ${DOCKER_COMPOSE_OVERRIDE} up littr-swagger -d --force-recreate \
		|| docker compose -f ${DOCKER_COMPOSE_FILE} up littr-swagger -d --force-recreate

.PHONY: test_local
test_local: 
	@echo -e "\n${YELLOW} Running unit/integration tests using the local runtime... ${RESET}\n"
	@go test $(shell go list ./... | grep -v cmd/sse_client | grep -v cmd/dbench | grep -v pkg/models | grep -v pkg/helpers | grep -v pkg/frontend)

.PHONY: build
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

.PHONY: run
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

.PHONY: run_test_env
run_test_env:	
	@echo -e "\n${YELLOW} Starting test project (docker compose up)... ${RESET}\n"
	@[ -f ".env" ] || cp .env.example .env
	@[ -f ${DOCKER_COMPOSE_TEST_OVERRIDE} ] \
		&& docker compose -f ${DOCKER_COMPOSE_TEST_FILE} -f ${DOCKER_COMPOSE_TEST_OVERRIDE} up --force-recreate --detach --remove-orphans \
		|| docker compose -f ${DOCKER_COMPOSE_TEST_FILE} up --force-recreate --detach --remove-orphans

.PHONY: logs
logs:
	@echo -e "\n${YELLOW} Fetching container's logs (CTRL-C to exit)... ${RESET}\n"
	@docker logs ${DOCKER_CONTAINER_NAME} -f

.PHONY: stop
stop:  
	@echo -e "\n${YELLOW} Stopping and purging project (docker compose down)... ${RESET}\n"
	@[ -f ".env" ] || cp .env.example .env
	@docker compose -f ${DOCKER_COMPOSE_FILE} down

.PHONY: version
version: 
	@[ -f "./.env" ] && head -n 8 .env | \
		sed -e 's/\(APP_PEPPER\)=\(.*\)/\1=xxx/' | \
		sed -e 's/\(REGISTRY\)=\(.*\)/\1=""/' | \
		sed -e 's/\(MAIL_SASL_USR\)=\(.*\)/\1=xxx/' | \
		sed -e 's/\(MAIL_SASL_PWD\)=\(.*\)/\1=xxx/' | \
		sed -e 's/\(MAIL_HOST\)=\(.*\)/\1=xxx/' | \
		sed -e 's/\(MAIL_PORT\)=\(.*\)/\1=xxx/' | \
		sed -e 's/\(GSC_TOKEN\)=\(.*\)/\1=xxx/' | \
		sed -e 's/\(GSC_URL\)=\(.*\)/\1=xxx/' | \
		sed -e 's/\(VAPID_PRIV_KEY\)=\(.*\)/\1=xxx/' | \
		sed -e 's/\(VAPID_PUB_KEY\)=\(.*\)/\1=xxx/' | \
		sed -e 's/\(VAPID_SUBSCRIBER\)=\(.*\)/\1=xxx/' | \
		sed -e 's/\(LOKI_URL\)=\(.*\)/\1=http:\/\/loki.example.com\/loki\/api\/v1\/push/' | \
		sed -e 's/\(APP_URLS_TRAEFIK\)=\(.*\)/\1=`littr.example.com`/' | \
		sed -e 's/\(API_TOKEN\)=\(.*\)/\1=xxx/' > .env.example && \
		sed -i 's/\/\/\(.*[[:blank:]]\)[0-9]*\.[0-9]*\.[0-9]*/\/\/\1${APP_VERSION}/' pkg/backend/router.go

.PHONY: docs_host
docs_host:
	@echo -e "\n${YELLOW} Updating the host for docs... ${RESET}\n"
	sed -i 's/\/\/.*\(@host[[:blank:]]*\)[a-z.0-9]*/\/\/ \1${APP_URL_MAIN}/' pkg/backend/router.go

.PHONY: push
push:
	@echo -e "\n${YELLOW} Pushing to git with tags... ${RESET}\n"
	@git tag -fa 'v${APP_VERSION}' -m 'v${APP_VERSION}'
	@git push --follow-tags --set-upstream origin master
	
.PHONY: sh
sh:
	@echo -e "\n${YELLOW} Attaching container's (${DOCKER_CONTAINER_NAME})... ${RESET}\n"
	@[ -f ".env" ] || cp .env.example .env
	@docker exec -it ${DOCKER_CONTAINER_NAME} sh

.PHONY: flush
flush:
	@echo -e "\n${YELLOW} Flushing app data... ${RESET}\n"
	@docker cp test/data/polls.json ${DOCKER_CONTAINER_NAME}:/opt/data/polls.json
	@docker cp test/data/posts.json ${DOCKER_CONTAINER_NAME}:/opt/data/posts.json
	@docker cp test/data/users.json ${DOCKER_CONTAINER_NAME}:/opt/data/users.json
	@docker cp test/data/subscriptions.json ${DOCKER_CONTAINER_NAME}:/opt/data/subscriptions.json
	@docker cp test/data/tokens.json ${DOCKER_CONTAINER_NAME}:/opt/data/tokens.json

.PHONY: kill
kill:
	@echo -e "\n${YELLOW} Killing the container not to dump running caches... ${RESET}\n"
	@[ -f ".env" ] || cp .env.example .env
	@docker kill ${DOCKER_CONTAINER_NAME}

RUN_DATA_DIR=./.run_data
.PHONY: fetch_running_dump
fetch_running_dump:
	@echo -e "\n${YELLOW} Copying dumped data from the container... ${RESET}\n"
	@mkdir -p ${RUN_DATA_DIR}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/users.json ${RUN_DATA_DIR}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/polls.json ${RUN_DATA_DIR}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/posts.json ${RUN_DATA_DIR}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/tokens.json ${RUN_DATA_DIR}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/subscriptions.json ${RUN_DATA_DIR}
	
.PHONY: backup
backup: fetch_running_dump
	@echo -e "\n${YELLOW} Making the backup archive... ${RESET}\n"
	@tar czvf /mnt/backup/littr/$(shell date +"%Y-%m-%d-%H:%M:%S").tar.gz ${RUN_DATA_DIR}

.PHONY: push_to_registry
push_to_registry:
	@echo -e "\n${YELLOW} Pushing new image to registry... ${RESET}\n"
	@[ -n "${REGISTRY}" ] || exit 10
	@echo "${REGISTRY_PASSWORD}" | docker login -u "${REGISTRY_USER}" --password-stdin "${REGISTRY}" && \
		docker push ${DOCKER_IMAGE_TAG}
	@docker logout ${REGISTRY} > /dev/null

.PHONY: sse_client
sse_client:
	@echo -e "\n${YELLOW} Starting the SSE client... ${RESET}\n"
	@go run cmd/sse_client/main.go

.PHONY: push_mirror
push_mirror:
	@echo -e "\n${YELLOW} Pushing to the mirror repository... ${RESET}\n"
	@git push mirror master --follow-tags
	
.PHONY: sonar_scan
sonar_scan:
	@if [ \( -n "${SONAR_URL}" \) -a \( -z "${SONAR_PROJECT_TOKEN}" \) ]; \
		then \
		sonar-scanner \
		-Dsonar.projectKey=${APP_NAME} \
		-Dsonar.sources=. \
		-Dsonar.host.url=${SONAR_URL}   \
		-Dsonar.login=${SONAR_PROJECT_TOKEN}; \
		fi

