.PHONY: dependency unit-test integration-test docker-up docker-down protobuf

#protobuf:
#    @protoc -I commons/protocols/ commons/protocols/*.proto --go_out # = plugins=grpc:commons/protocols

protoc-gen:
	protoc -I commons/protocols/ commons/protocols/*.proto --go_out=plugins=grpc:commons/protocols

dependency:
	@go get -v ./...

integration-test: docker-up dependency
	@go test -v ./...

unit-test: dependency
	@go test -v -short ./...

docker-up:
	docker-compose up -d

docker-down:
	@docker-compose down

clear: docker-down

