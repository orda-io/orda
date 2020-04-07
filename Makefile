BUILD_DIR = build
GOSRCS := $(shell find . -path ./vendor -prune -o -type f -name '*.go' -print)

.PHONY: protoc-gen
protoc-gen:
	protoc ortoo/model/*.proto \
			-I=./ortoo/model/ \
			--gofast_out=plugins=grpc,Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,:./ortoo/model/
#			--gotag_out=xxx="bson+\"-\"",output_path=./ortoo/model/:.
	protoc-go-inject-tag -input=./ortoo/model/model.pb.go

.PHONY: server
server:
	mkdir -p $(BUILD_DIR)
	cd server && CGO_ENABLED=1 go build -race -gcflags='all=-N -l' -o ../$(BUILD_DIR)

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
	@docker-compose up -d

.PHONY: docker-down
docker-down:
	@docker-compose down

.PHONY: clear
clear: docker-down

.PHONY: lint
lint: dependency
	golint ./...