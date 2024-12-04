.PHONY: help build build-local up down logs ps
.DEFAULT_GOAL := help
DOCKER_TAG := latest

build:
	docker build -t seoppak/dbserver:${DOCKER_TAG} \
		-f ./dbServer/Dockerfile \
		--target deploy ./dbServer
	docker build -t seoppak/loginserver:${DOCKER_TAG} \
		-f ./loginServer/Dockerfile \
		--target deploy ./loginServer
	docker build -t seoppak/ocrserver:${DOCKER_TAG} \
		-f ./ocrServer/Dockerfile \
		--target deploy ./ocrServer

build-local:
	docker compose build --no-cache

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'