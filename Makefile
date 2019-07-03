.PHONY: dependency unit-test integration-test docker-up docker-down

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

protobuf:
    protoc -I commons/protocols/ commons/protocols/*.proto --go_out=plugins=grpc:commons/protocols