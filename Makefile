VERSION = 0.0.1

BUILD_DIR = bin
DEPLOY_DIR = deployments
CONFIG_DIR = config

GIT_COMMIT := $(shell git rev-parse --short HEAD)

GO_SRCS := $(shell find . -path ./vendor -prune -o -type f -name "*.go" -print)
GO_PROJECT = github.com/knowhunger/ortoo
GO_LDFLAGS ?=
GO_LDFLAGS += -X ${GO_PROJECT}/ortoo/version.GitCommit=${GIT_COMMIT}
GO_LDFLAGS += -X ${GO_PROJECT}/ortoo/version.Version=${VERSION}

.PHONY: protoc-gen
protoc-gen:
	-rm ./pkg/model/model.pb.go
	protoc ./pkg/model/*.proto \
			-I=./pkg/model/ \
			--gofast_out=plugins=grpc,:./pkg/model/
	protoc-go-inject-tag -input=./pkg/model/model.pb.go

.PHONY: build-server
build-server:
	echo $(GO_SRCS)
	mkdir -p $(BUILD_DIR)
	cd server && go build -gcflags='all=-N -l' -ldflags "${GO_LDFLAGS}" -o ../$(BUILD_DIR)

.PHONY: dependency
dependency:
	go get -v ./...
	go get github.com/gogo/protobuf/proto
	go get github.com/gogo/protobuf/gogoproto
	go get github.com/gogo/protobuf/protoc-gen-gogo
	go get github.com/gogo/protobuf/protoc-gen-gofast
	go get github.com/tebeka/go2xunit
	go get golang.org/x/lint/golint
	go get github.com/axw/gocov/gocov
	go get github.com/AlekSi/gocov-xml
	go get github.com/favadi/protoc-go-inject-tag
	go get github.com/amsokol/protoc-gen-gotag

.PHONY: fmt
fmt:
	gofmt -w $(GO_SRCS)
	goimports -w -local github.com/knowhunger $(GO_SRCS)

.PHONY: integration-test
integration-test: docker-up dependency
	@go test -v -race ./...

.PHONY: unit-test
unit-test: dependency
	@go test -v -short -race ./...

.PHONY: docker-up
docker-up:
	@cd $(DEPLOY_DIR); docker-compose up -d

.PHONY: docker-down
docker-down:
	@cd $(DEPLOY_DIR); docker-compose down

.PHONY: build-local-docker-server
build-local-docker-server: build-server
	-mkdir -p $(DEPLOY_DIR)/tmp
	cp $(BUILD_DIR)/server  $(DEPLOY_DIR)/tmp
	@cd $(DEPLOY_DIR) && docker build -t knowhunger/ortoo:$(VERSION) .
	-rm -rf $(DEPLOY_DIR)/tmp

.PHONY: clear
clear: docker-down

.PHONY: lint
lint: dependency
	golint ./...