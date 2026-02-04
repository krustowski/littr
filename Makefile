#
#  littr / Makefile
#

#
#  VARS
#

include .env.example
-include .env

#
#  Basic common runtime vars
#

APP_ENVIRONMENT 	?= dev
APP_NAME 		:= littr
APP_URLS_TRAEFIK 	?= `${HOSTNAME}`
APP_URL_MAIN 		?= ${HOSTNAME}
PROJECT_NAME 		:= ${APP_NAME}-${APP_ENVIRONMENT}
TZ 			?= Europe/Vienna

LOKI_URL 		?=

APP_PEPPER 		?=
API_TOKEN 		?=

#
#  Defaults for backend
#

LIMITER_ENABLED 	?= true
REGISTRATION_ENABLED 	?= true

DATA_DUMP_FORMAT       	?= JSON
DATA_LOAD_FORMAT     	?= JSON

#
#  Subscription (webpush) vars
#

VAPID_PUB_KEY 		?=
VAPID_PRIV_KEY 		?=
VAPID_SUBSCRIBER 	?=

#
#  Mailing vars
#

MAIL_HELO 		?= localhost
MAIL_HOST 		?= localhost
MAIL_PORT 		?= 25
MAIL_SASL_USR 		?=
MAIL_SASL_PWD 		?=

#
#  Go environment vars
#

#COMMON_BUILD_LDFLAGS	:= -s -w
COMMON_BUILD_LDFLAGS 	:=
GOARCH 			:= $(shell go env GOARCH)
GOCACHE 		?= /home/${USER}/.cache/go-build
GOMODCACHE 		?= /home/${USER}/go/pkg/mod
GOOS 			:= $(shell go env GOOS)

# go build -race [...]
RACE_FLAG 		?= ""

#
#  Docker environment vars 
#

DOCKER_COMPOSE_FILE 		?= deployments/docker-compose.yml
DOCKER_COMPOSE_TEST_FILE 	?= deployments/docker-compose-test.yml
DOCKER_COMPOSE_OVERRIDE 	?= deployments/docker-compose.override.yml
DOCKER_COMPOSE_TEST_OVERRIDE 	?= deployments/docker-compose-test.override.yml
DOCKER_CONTAINER_NAME 		?= ${PROJECT_NAME}-server

REGISTRY 			?= ${APP_NAME}
DOCKER_BUILD_IMAGE 		?= golang:${GOLANG_VERSION}-alpine
DOCKER_BUILD_IMAGE_RELEASE 	?= alpine:${ALPINE_VERSION}
DOCKER_IMAGE_TAG 		?= ${REGISTRY}/littr/backend:${APP_VERSION}-go${GOLANG_VERSION}

DOCKER_INTERNAL_PORT 		?= 8080
DOCKER_NETWORK_NAME 		?= traefik
DOCKER_USER 			?= littr
DOCKER_VOLUME_DATA_NAME 	?= littr-data
DOCKER_VOLUME_PIX_NAME 		?= littr-pix

DOCKER_SWAGGER_CONTAINER_NAME 	?= littr-swagger

#
#  Define standard colors for CLI
# https://gist.github.com/rsperl/d2dfe88a520968fbc1f49db0a29345b9
#

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
#  Deployment targets (chains)
#

.PHONY: dev prod

dev: version fmt build check_docker run logs

prod: run logs

#
#  Development targets (inc CI tests)
#

DNS_NAMESERVER 		?= 1.1.1.1

.PHONY: check_docker check_env config fmt push push_mirror sonar_scan test_local test_local_coverage

check_docker:
	@docker inspect ${DOCKER_VOLUME_DATA_NAME} 2>&1 > /dev/null || docker volume create ${DOCKER_VOLUME_DATA_NAME}
	@docker inspect ${DOCKER_VOLUME_PIX_NAME} 2>&1 > /dev/null || docker volume create ${DOCKER_VOLUME_PIX_NAME}
	@docker inspect ${DOCKER_NETWORK_NAME} 2>&1 > /dev/null || docker network create ${DOCKER_NETWORK_NAME}
	@docker plugin inspect grafana/loki-docker-driver:latest 2>&1 > /dev/null || docker plugin install --grant-all-permissions grafana/loki-docker-driver:latest

check_env:
	@[ -f ".env" ] || cp .env.example .env

config:
	$(call print_info, Running the local environment configuration setup...)
	@go install github.com/swaggo/swag/cmd/swag@latest

deps:
	$(call print_info, Fetching/upgrading dependencies...)
	@go get -u ./...
	@go mod tidy

fmt:
	$(call print_info, Reformatting the code using gofmt tool...)
	@gofmt -w -s .

push:
	$(call print_info, Pushing tagged commits to origin/master...)
	@git tag -fa 'v${APP_VERSION}' -m 'v${APP_VERSION}'
	@git push --follow-tags --set-upstream origin master
	
push_mirror:
	$(call print_info, Pushing tagged commits to mirror/master...)
	@git push --follow-tags mirror master

ifeq (${SONAR_HOST_URL}${SONAR_TOKEN},)
sonar_check:
else
sonar_check:
	$(call print_info, Starting the sonarqube code analysis...)
	@docker run --rm \
		--dns ${DNS_NAMESERVER} \
		-e SONAR_HOST_URL="${SONAR_HOST_URL}" \
		-e SONAR_TOKEN="${SONAR_TOKEN}" \
		-v ".:/usr/src" \
		sonarsource/sonar-scanner-cli
endif

test_local: fmt
	$(call print_info, Running Go unit/integration tests locally...)
	@go clean -testcache
	@go test -tags server \
		$(shell go list ./... | grep -v cmd/sse_client | grep -v cmd/dbench | grep -v pkg/models | grep -v pkg/helpers | grep -v pkg/frontend)

test_local_coverage: fmt
	$(call print_info, Running Go code coverage analysis...)
	@go clean -testcache
	@go test -tags server -v -coverprofile coverage.out ./... && \
		go tool cover -html coverage.out

#
#  Versioning (semver incrementing) targets
#

.PHONY: major minor patch version

define update_semver
	$(call print_info, Incrementing semver to ${1}...)
	@[ -f ".env" ] || cp .env.example .env
	@sed -i 's|APP_VERSION=.*|APP_VERSION=${1}|' .env
	@sed -i 's|APP_VERSION=.*|APP_VERSION=${1}|' .env.example
	@sed -i 's|sonar.projectVersion=.*|sonar.projectVersion=${1}|' sonar-project.properties
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

ifeq (${REGISTRY},)
build: check_env
	$(call print_info, Building the docker image (docker compose build)...)
	@[ -f "${DOCKER_COMPOSE_OVERRIDE}" ] || touch ${DOCKER_COMPOSE_OVERRIDE}
	@DOCKERFILE=full.dockerfile DOCKER_BUILDKIT=1 \
		docker compose -f ${DOCKER_COMPOSE_FILE} -f ${DOCKER_COMPOSE_OVERRIDE} build
else
build: check_env
	$(call print_info, Building the docker image (docker compose build)...)
	@[ -f "${DOCKER_COMPOSE_OVERRIDE}" ] || touch ${DOCKER_COMPOSE_OVERRIDE}
	@echo "${REGISTRY_PASSWORD}" | docker login -u "${REGISTRY_USER}" --password-stdin "${REGISTRY}"
	@DOCKERFILE=prebuilt.dockerfile DOCKER_BUILDKIT=1 \
		docker compose -f ${DOCKER_COMPOSE_FILE} -f ${DOCKER_COMPOSE_OVERRIDE} build
endif

docs: config check_env
	$(call print_info, Generating OpenAPI Swagger documentation...)
	@~/go/bin/swag init --parseDependency -ot json -g router.go --dir pkg/backend/ 
	@mv docs/swagger.json api/swagger.json
	@[ -f "${DOCKER_COMPOSE_OVERRIDE}" ] || touch ${DOCKER_COMPOSE_OVERRIDE}
	@docker compose -f ${DOCKER_COMPOSE_FILE} -f ${DOCKER_COMPOSE_OVERRIDE} up littr-swagger -d --force-recreate

docs_host:
	$(call print_info, Updating the baseURL in OpenAPI docs according to env...)
	@sed -i 's/\/\/.*\(@host[[:blank:]]*\)[a-z.0-9]*/\/\/ \1${APP_URL_MAIN}/' pkg/backend/router.go

ifeq (${REGISTRY},)
push_to_registry:
else
push_to_registry:
	$(call print_info, Pushing tagged docker image to registry...)
	@echo "${REGISTRY_PASSWORD}" | docker login -u "${REGISTRY_USER}" --password-stdin "${REGISTRY}"
	@docker push ${DOCKER_IMAGE_TAG}
	@docker logout ${REGISTRY} > /dev/null
endif

ifeq (${REGISTRY},)
run: check_env check_docker
	$(call print_info, Starting the docker compose stack up...)
	@[ -f "${DOCKER_COMPOSE_OVERRIDE}" ] || touch ${DOCKER_COMPOSE_OVERRIDE}
	@docker compose -f ${DOCKER_COMPOSE_FILE} -f ${DOCKER_COMPOSE_OVERRIDE} up --force-recreate --detach --remove-orphans
else
run: check_env check_docker
	$(call print_info, Starting the docker compose stack up...)
	@[ -f "${DOCKER_COMPOSE_OVERRIDE}" ] || touch ${DOCKER_COMPOSE_OVERRIDE}
	@echo "${REGISTRY_PASSWORD}" | docker login -u "${REGISTRY_USER}" --password-stdin "${REGISTRY}"
	@docker compose -f ${DOCKER_COMPOSE_FILE} -f ${DOCKER_COMPOSE_OVERRIDE} up --force-recreate --detach --remove-orphans
	@docker logout "${REGISTRY}" > /dev/null
endif

run_test_env: check_env
	$(call print_info, Starting the docker compose stack up (test env)...)
	@[ -f "${DOCKER_COMPOSE_TEST_OVERRIDE}" ] || touch ${DOCKER_COMPOSE_TEST_OVERRIDE}
	@docker compose -f ${DOCKER_COMPOSE_TEST_FILE} -f ${DOCKER_COMPOSE_TEST_OVERRIDE} up --force-recreate --detach --remove-orphans

#
#  Profiling targets
#

GO_TOOL_PPROF 		:= go tool pprof
NOL  			:= $(shell ps auxf | grep -w '${GO_TOOL_PPROF}' | wc -l | cut -d' ' -f1)
LIST 			:= $(shell ps auxf | grep -w '${GO_TOOL_PPROF}' | tail -n $$(( $(NOL) - 2 )) | awk '{ print $$2 }')
PPROF_SOURCE 		?= http://localhost:${DOCKER_INTERNAL_PORT}/debug/pprof

.PHONY: kill_pprof run_pprof

kill_pprof:
	$(call print_info, Killing all profiling instances...)
	@for INST in ${LIST}; do kill $${INST}; done

run_pprof: kill_pprof
	$(call print_info, Starting the profiling instances to analyze the runtime...)
	@go build -tags server -o ./littr -pkgdir ./pkg/ ./cmd/littr/
	@mkdir -p .tmp
	@curl -sL ${PPROF_SOURCE}/allocs?debug=1 > .tmp/alloc.out
	@curl -sL ${PPROF_SOURCE}/goroutine?debug=1 > .tmp/goroutine.out
	@curl -sL ${PPROF_SOURCE}/heap?debug=1 > .tmp/heap.out
	@go tool pprof -http=127.0.0.1:8081 ./littr .tmp/alloc.out &
	@go tool pprof -http=127.0.0.1:8082 ./littr .tmp/goroutine.out &
	@go tool pprof -http=127.0.0.1:8083 ./littr .tmp/heap.out &
	
#
#  Runtime (live system operation) targets
#

BACKUP_PATH    		?= /mnt/backup/littr
RUN_DATA_PATH  		?= ./.run_data
DEMO_DATA_PATH 		?= test/data

.PHONY: backup fetch_running_dump flush kill logs sh sse_client stop

backup: fetch_running_dump
	$(call print_info, Creating a backup archive...)
	@[ -d "${BACKUP_DIR}" ] || exit 5
	@tar czvf ${BACKUP_PATH}/$(shell date +"%Y-%m-%d-%H:%M:%S").tar.gz ${RUN_DATA_PATH}

fetch_running_dump:
	$(call print_info, Fetching current dump data from the container...)
	@mkdir -p ${RUN_DATA_PATH}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/polls.json ${RUN_DATA_PATH}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/posts.json ${RUN_DATA_PATH}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/requests.json ${RUN_DATA_PATH}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/subscriptions.json ${RUN_DATA_PATH}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/tokens.json ${RUN_DATA_PATH}
	@docker cp ${DOCKER_CONTAINER_NAME}:/opt/data/users.json ${RUN_DATA_PATH}
	
flush:
	$(call print_info, Flushing the running app data...)
	@[ -d ${DEMO_DATA_PATH} ] || exit 6
	@docker cp ${DEMO_DATA_PATH}/polls.json ${DOCKER_CONTAINER_NAME}:/opt/data/polls.json
	@docker cp ${DEMO_DATA_PATH}/posts.json ${DOCKER_CONTAINER_NAME}:/opt/data/posts.json
	@docker cp ${DEMO_DATA_PATH}/subscriptions.json ${DOCKER_CONTAINER_NAME}:/opt/data/subscriptions.json
	@docker cp ${DEMO_DATA_PATH}/tokens.json ${DOCKER_CONTAINER_NAME}:/opt/data/tokens.json
	@docker cp ${DEMO_DATA_PATH}/users.json ${DOCKER_CONTAINER_NAME}:/opt/data/users.json

kill: check_env
	$(call print_info, Killing the running container not to dump its caches...)
	@docker kill ${DOCKER_CONTAINER_NAME}

logs:
	$(call print_info, Attaching and following the container's (${DOCKER_CONTAINER_NAME}) logs...)
	@docker logs ${DOCKER_CONTAINER_NAME} -f

sh: check_env
	$(call print_info, Attaching the container's (${DOCKER_CONTAINER_NAME}) shell...)
	@docker exec -it ${DOCKER_CONTAINER_NAME} sh

sse_client:
	$(call print_info, Starting the custom SSE Go client...)
	@go run ./cmd/sse_client/main.go

stop: check_env
	$(call print_info, Stopping and purging the docker stack (docker compose down)...)
	@docker compose -f ${DOCKER_COMPOSE_FILE} down

