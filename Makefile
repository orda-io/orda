VERSION := $(shell cat version)

BUILD_DIR = bin
RESOURCES_DIR = resources
DEPLOY_DIR = deployments
CONFIG_DIR = config

GIT_COMMIT := $(shell git rev-parse --short HEAD)

GO_SRCS := $(shell find . -path ./vendor -prune -o -type f -name "*.go" -print)
GO_PROJECT = github.com/orda-io/orda
GO_LDFLAGS ?=
GO_LDFLAGS += -X '${GO_PROJECT}/pkg/constants.GitCommit=${GIT_COMMIT}'
GO_LDFLAGS += -X '${GO_PROJECT}/pkg/constants.Version=${VERSION}'

PROJECT_ROOT = $(shell pwd)
DOCKER_PROJECT_ROOT = /app

PROTOC_INCLUDE := -I=./proto/thirdparty -I=./proto
PROTOC_PROTO_FILES := orda.enum.proto orda.proto orda.grpc.proto

ORDA_BUILDER := docker run --platform linux/amd64 --network host --rm -v $(PROJECT_ROOT):${DOCKER_PROJECT_ROOT} -w ${DOCKER_PROJECT_ROOT} orda-builder sh -c

.PHONY: init
init:
	docker build --platform linux/amd64 -t orda-builder .

.PHONY: protoc-gen
protoc-gen:
	- $(ORDA_BUILDER) "rm -rf ./proto/thirdparty ./pkg/model/*.pb.go ./pkg/model/*.gw.go"
	$(ORDA_BUILDER) "mkdir -p ./proto/thirdparty/google"
	$(ORDA_BUILDER) "cp -rf /root/googleapis/google/api ./proto/thirdparty/google/"
	$(ORDA_BUILDER) "mkdir -p ./proto/thirdparty/protoc-gen-openapiv2"
	$(ORDA_BUILDER) "cp -rf /root/grpc-gateway/protoc-gen-openapiv2/options ./proto/thirdparty/protoc-gen-openapiv2/"
	$(ORDA_BUILDER) "protoc $(PROTOC_INCLUDE) --go_out=,plugins=grpc,:. $(PROTOC_PROTO_FILES)"
	$(ORDA_BUILDER) "protoc-go-inject-tag -input=./pkg/model/orda.pb.go"
	$(ORDA_BUILDER) "protoc $(PROTOC_INCLUDE) \
		--grpc-gateway_out ./pkg/model \
		--grpc-gateway_opt paths=source_relative \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt generate_unbound_methods=true \
		--openapiv2_out $(RESOURCES_DIR) \
		--openapiv2_opt logtostderr=true \
		orda.grpc.proto"

.PHONY: install-golibs
install-golibs:
	go install github.com/tebeka/go2xunit@latest
	go install golang.org/x/lint/golint@latest
	go install github.com/axw/gocov/gocov@latest
	go install github.com/AlekSi/gocov-xml@latest
	go install github.com/favadi/protoc-go-inject-tag@latest
	go install github.com/amsokol/protoc-gen-gotag@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/golang/protobuf/protoc-gen-go@latest

.PHONY: dependency
dependency: install-golibs
	go get -v ./...

.PHONY: fmt
fmt:
	$(ORDA_BUILDER) "gofmt -w $(GO_SRCS)"
	$(ORDA_BUILDER) "goimports -w -local github.com/orda-io $(GO_SRCS)"

.PHONY: lint
lint:
	$(ORDA_BUILDER) "golint ./..."

.PHONY: test
test:
	$(ORDA_BUILDER) "go test -v --race ./..."

.PHONY: build-server
build-server:
	echo $(PROJECT_ROOT)
	$(ORDA_BUILDER) "mkdir -p $(BUILD_DIR)"
	$(ORDA_BUILDER) "cd server && go build -gcflags='all=-N -l' -ldflags '${GO_LDFLAGS}' -o ../$(BUILD_DIR)"

.PHONY: build-docker-server
build-docker-server: build-server
	- $(ORDA_BUILDER) "mkdir -p $(DEPLOY_DIR)/tmp"
	$(ORDA_BUILDER) "cp $(BUILD_DIR)/server  $(DEPLOY_DIR)/tmp"
	$(ORDA_BUILDER) "cp -rf $(RESOURCES_DIR) $(DEPLOY_DIR)/tmp"
	@cd $(DEPLOY_DIR) && docker build --platform linux/amd64 -t orda-io/orda:$(VERSION) .
	-$(ORDA_BUILDER) "rm -rf $(DEPLOY_DIR)/tmp"

.PHONY: clear
clear: docker-down

.PHONY: docker-up
docker-up:
	@cd $(DEPLOY_DIR); VERSION=$(VERSION) docker-compose up -d

.PHONY: docker-down
docker-down:
	@cd $(DEPLOY_DIR); VERSION=$(VERSION) docker-compose down

.PHONY: check-before-pr
check-before-pr: lint
