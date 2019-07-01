

protobuf:
    protoc -I commons/protocols/ commons/protocols/*.proto --go_out=plugins=grpc:commons/protocols