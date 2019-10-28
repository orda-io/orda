.PHONY: dependency unit-test integration-test docker-up docker-down protobuf lint

protoc-gen:
	protoc commons/model/*.proto \
			-I=./commons/model/ \
			--gofast_out=plugins=grpc,Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,:./commons/model/ \
			--gotag_out=xxx="bson+\"-\"",output_path=./commons/model:.
	protoc-go-inject-tag -input=./commons/model/model.pb.go

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