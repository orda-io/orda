VERSION = 0.0.1
BUILD_DIR = bin

GIT_COMMIT := $(shell git rev-parse --short HEAD)

GOSRCS := $(shell find . -path ./vendor -prune -o -type f -name '*.go' -print)

GO_PROJECT = github.com/knowhunger/ortoo

GO_LDFLAGS ?=
GO_LDFLAGS += -X ${GO_PROJECT}/ortoo/version.GitCommit=${GIT_COMMIT}
GO_LDFLAGS += -X ${GO_PROJECT}/ortoo/version.Version=${VERSION}

.PHONY: protoc-gen
protoc-gen:
	-rm ./ortoo/model/model.pb.go
	protoc ortoo/model/*.proto \
			-I=./ortoo/model/ \
			--gofast_out=plugins=grpc,:./ortoo/model/
	protoc-go-inject-tag -input=./ortoo/model/model.pb.go

.PHONY: server
server:
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
	gofmt -w $(GOSRCS)
	goimports -w -local github.com/knowhunger $(GOSRCS)

.PHONY: integration-test
integration-test: docker-up dependency
	@go test -v -race ./...

.PHONY: unit-test
unit-test: dependency
	@go test -v -short -race ./...

.PHONY: docker-up
docker-up:
	@cd deployments; docker-compose up -d

.PHONY: docker-down
docker-down:
	@cd deployments; docker-compose down

.PHONY: run-local-server
run-local-server: docker-up server
	$(BUILD_DIR)/server --conf examples/local-config.json

.PHONY: clear
clear: docker-down

.PHONY: lint
lint: dependency
	golint ./...