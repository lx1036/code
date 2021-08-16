





install:
	go get github.com/golang/protobuf/protoc-gen-go

beta1:
	protoc --proto_path=../../../../../../vendor --go_out=plugins=grpc:v1beta1 v1beta1/api.proto

# --go_out=plugins=grpc:v1beta2中，v1beta2表示生成的go文件放在当前目录的/v1beta2相对目录下
beta2:
	protoc --proto_path=../../../../../../vendor --go_out=plugins=grpc:v1beta2 v1beta2/api.proto


