.PHONY: dependency unit-test integration-test docker-up docker-down protobuf lint server

protoc-gen:
	protoc ortoo/model/*.proto \
			-I=./ortoo/model/ \
			--gofast_out=plugins=grpc,Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,:./ortoo/model/
#			--gotag_out=xxx="bson+\"-\"",output_path=./ortoo/model/:.
	protoc-go-inject-tag -input=./ortoo/model/model.pb.go

server:
	CGO_ENABLED=1 go build  -race -gcflags='all=-N -l' -o build/ortoo_server ./server/...

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

integration-test: docker-up dependency
	@go test -v -race ./...

unit-test: dependency
	@go test -v -short -race ./...

docker-up:
	@docker-compose up -d

docker-down:
	@docker-compose down

clear: docker-down

lint: dependency
	golint ./...