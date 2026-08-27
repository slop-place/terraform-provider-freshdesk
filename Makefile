# terraform-provider-freshdesk
BINARY  := terraform-provider-freshdesk
VERSION ?= dev
GOBIN   := $(shell go env GOPATH)/bin

.PHONY: default
default: check

## check: the full gate — format, vet, lint, test.
.PHONY: check
check: fmt vet lint test

## build: compile the provider binary.
.PHONY: build
build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) .

## install: build into the Go bin directory.
.PHONY: install
install:
	go install -ldflags "-X main.version=$(VERSION)" .

## fmt: apply gofmt and goimports.
.PHONY: fmt
fmt:
	gofmt -w -s ./freshdesk ./internal .
	@command -v goimports >/dev/null && goimports -w ./freshdesk ./internal . || true

## vet: run go vet.
.PHONY: vet
vet:
	go vet ./...

## lint: run golangci-lint with every linter enabled.
.PHONY: lint
lint:
	$(GOBIN)/golangci-lint run ./...

## test: unit tests with the race detector.
.PHONY: test
test:
	go test -race -count=1 ./...

## cover: unit tests with a coverage report.
.PHONY: cover
cover:
	go test -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

## testacc: acceptance tests against a real Freshdesk account.
## Requires FRESHDESK_DOMAIN and FRESHDESK_API_KEY (see .env).
.PHONY: testacc
testacc:
	TF_ACC=1 go test -count=1 -timeout 30m -v ./internal/provider/...

## fieldaudit: check the client maps every field the live API returns.
## Requires FRESHDESK_DOMAIN and FRESHDESK_API_KEY, and a seeded sandbox
## (run `make seed` first if the account is empty).
.PHONY: fieldaudit
fieldaudit:
	python3 internal/apispec/fieldaudit.py

## seed: create one record of each kind the field audit would otherwise skip.
.PHONY: seed
seed:
	bash internal/apispec/seed.sh

## clean-sandbox: remove records left by the acceptance suite or the audit.
.PHONY: clean-sandbox
clean-sandbox:
	bash internal/apispec/cleanup.sh

## coverage-api: verify the client still covers the documented API surface.
.PHONY: coverage-api
coverage-api:
	python3 internal/apispec/coverage.py

## docs: regenerate the registry documentation.
.PHONY: docs
docs:
	$(GOBIN)/tfplugindocs generate --provider-name freshdesk

## tools: install the development tools this Makefile expects.
.PHONY: tools
tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	go install golang.org/x/tools/cmd/goimports@latest

## help: list the targets.
.PHONY: help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'
