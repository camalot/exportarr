# set APP_ENV=dist if not set
APP_ENV ?= dist
-include .env
-include .env.$(APP_ENV)

.PHONY: build run check fmt tidy lint test

build:
	docker build . -t exportarr:local


run:
	@echo "Running exportarr with APP_ENV = ${APP_ENV}"
	
	@docker rm --force exportarr || echo ""
	docker run --name exportarr \
		-e PORT=9707 \
		-e URL="${APP_URL}" \
		-e APIKEY="${APP_API_KEY}" \
		-e LOG_LEVEL="debug" \
		-p 9707:9707 \
		exportarr:local ${APP_NAME}

check: fmt tidy lint test

fmt:
	go fmt ./...

tidy:
	go mod tidy

lint:
	golangci-lint run -c .github/lint/golangci.yaml

test:
	go test -v -race -covermode atomic -coverprofile=covprofile ./...