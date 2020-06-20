PROJECT_NAME := "arch-weekly-digest"
PKG := "github.com/otanikotani/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

.PHONY: all

all: build

init.done:
	terraform init

deploy.done: init.done zip
	terraform apply -auto-approve

zip: clean.zip build
	zip -j ${PROJECT_NAME}.zip build/*

tidy: dep
	@go mod tidy

dep: ## Get the dependencies
	@go mod download

lint: ## Lint Golang files
	@golint -set_exit_status ${PKG_LIST}

vet: ## Run go vet
	@go vet ${PKG_LIST}

test: ## Run unittests
	@go test -short ${PKG_LIST}

build: tidy ## Build the binary file
	@go build -i -o build/${PROJECT_NAME} $(PKG)

clean.zip:
	@rm -rf build/
	rm ${PROJECT_NAME}.zip || true

clean:
	terraform destroy -auto-approve
	@rm -rf build/
	rm ${PROJECT_NAME}.zip || true
