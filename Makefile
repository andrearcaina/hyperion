.PHONY: build clean generate test

generate:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/hyperion.proto

build:
	go build -o ./bin/hyprd ./cmd/hyprd
	go build -o ./bin/hyprctl ./cmd/hyprctl

clean:
	rm -f ./bin/hyprd ./bin/hyprctl

test:
	go test ./...
