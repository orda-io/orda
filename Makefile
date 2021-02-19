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

PROTOC_INCLUDE := -I=./proto -I=./proto/third_party
PROTOC_PROTO_FILES := ortoo.enum.proto ortoo.proto ortoo.grpc.proto

.PHONY: protoc-gen
protoc-gen:
	-rm ./pkg/model/*.pb.go ./pkg/model/*.pb.gw.go ./server/swagger-ui/ortoo.grpc.swagger.json
	protoc $(PROTOC_INCLUDE) \
		--gofast_out=,plugins=grpc,:./pkg/model/ \
		$(PROTOC_PROTO_FILES)
	protoc-go-inject-tag -input=./pkg/model/ortoo.pb.go
	protoc $(PROTOC_INCLUDE) \
		--grpc-gateway_out ./pkg/model \
		--grpc-gateway_opt logtostderr=true \
		--openapiv2_out ./server/swagger-ui \
		--openapiv2_opt logtostderr=true \
		ortoo.grpc.proto

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
	go get github.com/protoc-gen-swagger
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
	go get google.golang.org/protobuf/cmd/protoc-gen-go
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc


.PHONY: copy-assets
copy-assets:
	-rm -rf ./proto/third_party
	-mkdir -p ./proto/third_party/google
	-mkdir -p ./proto/third_party/protoc-gen-openapiv2
	cp -rf $(shell go env GOMODCACHE)/github.com/grpc-ecosystem/grpc-gateway/v2@v2.1.0/third_party/googleapis/google/api ./proto/third_party/google
	cp -rf $(shell go env GOMODCACHE)/github.com/grpc-ecosystem/grpc-gateway/v2@v2.1.0/protoc-gen-openapiv2/options ./proto/third_party/protoc-gen-openapiv2

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
