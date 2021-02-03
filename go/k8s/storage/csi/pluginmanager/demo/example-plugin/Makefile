





install:
	go get github.com/golang/protobuf/protoc-gen-go

beta1:
	protoc --proto_path=v1beta1 --proto_path=../../../../../../vendor --go_out=plugins=grpc:v1beta1 v1beta1/api.proto

beta2:
	protoc --proto_path=v1beta2 --proto_path=../../../../../../vendor --go_out=plugins=grpc:v1beta2 v1beta2/api.proto


