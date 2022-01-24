


# https://grpc.io/docs/languages/go/quickstart/#regenerate-grpc-code
grpc:
	rm -f transport.pb.go
	rm -f transport_grpc.pb.go
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative transport.proto
