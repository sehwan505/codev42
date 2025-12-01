.PHONY: build run proto db-diff db-apply test

build:
	go build -o bin/plan-service main.go

run:
	go run main.go

proto:
	rm -f pb/*.go
	protoc --proto_path=pb \
		--go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		pb/*.proto

db-diff:
	atlas migrate diff --env "plan"

db-apply:
	atlas migrate apply --env "plan"

test:
	go test -v -cover -short ./...

