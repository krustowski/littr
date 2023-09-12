#
# litter-go / Makefile
#

#
# VARS
#

include .env.example
-include .env

PROJECT_NAME?=litter-go

DOCKER_IMAGE_TAG?=${PROJECT_NAME}-image
DOCKER_CONTAINER_NAME?=${PROJECT_NAME}-server

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
	@echo -e "${YELLOW} make doc     --- render documentation from code (go doc) ${RESET}\n"

	@echo -e "${YELLOW} make build   --- build project (docker image) ${RESET}"
	@echo -e "${YELLOW} make run     --- run project ${RESET}"
	@echo -e "${YELLOW} make logs    --- fetch container's logs ${RESET}"
	@echo -e "${YELLOW} make stop    --- stop and purge project (only docker containers!) ${RESET}"
	@echo -e ""


.PHONY: fmt
fmt: version
	@echo -e "\n${YELLOW} Code reformating (gofmt)... ${RESET}\n"
	@gofmt -w -s .

.PHONY: build
build: 
	@echo -e "\n${YELLOW} Building the project (docker compose build)... ${RESET}\n"
	@docker compose build --no-cache

.PHONY: run
run:	
	@echo -e "\n${YELLOW} Starting project (docker compose up)... ${RESET}\n"
	@docker compose up --force-recreate --detach --remove-orphans

.PHONY: logs
logs:
	@echo -e "\n${YELLOW} Fetching container's logs (CTRL-C to exit)... ${RESET}\n"
	@docker logs ${DOCKER_CONTAINER_NAME} -f

.PHONY: stop
stop:  
	@echo -e "\n${YELLOW} Stopping and purging project (docker compose down)... ${RESET}\n"
	@docker compose down

.PHONY: version
version: 
	@[ -f "./.env" ] && cat .env | \
		sed -e 's/\(APP_PEPPER\)=\(.*\)/\1=xxx/' | \
		sed -e 's/\(API_TOKEN\)=\(.*\)/\1=yyy/' > .env.example

.PHONY: push
push:
	@echo -e "\n${YELLOW} Pushing to git with tags... ${RESET}\n"
	@git tag -fa 'v${APP_VERSION}' -m 'v${APP_VERSION}'
	@git push --follow-tags
	
.PHONY: sh
sh:
	@echo -e "\n${YELLOW} Attaching container's (${DOCKER_CONTAINER_NAME})... ${RESET}\n"
	@docker exec -it ${DOCKER_CONTAINER_NAME} sh

.PHONY: flush
flush:
	@echo -e "\n${YELLOW} Flushing app data... ${RESET}\n"
	@docker cp data/users.json ${DOCKER_CONTAINER_NAME}:/opt/data/users.json
	@docker cp data/posts.json ${DOCKER_CONTAINER_NAME}:/opt/data/posts.json

.PHONY: kill
kill:
	@echo -e "\n${YELLOW} Killing the container not to dump running caches... ${RESET}\n"
	@docker kill ${DOCKER_CONTAINER_NAME}
