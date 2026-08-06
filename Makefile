.PHONY: build clean generate test vet check \
	docker-build docker-run docker-stop docker-status docker-logs docker-clean

DOCKER_COMPOSE ?= docker compose

generate:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/hyperion.proto

build:
	go build -o ./bin/hyprd ./cmd/hyprd
	go build -o ./bin/hyprctl ./cmd/hyprctl

vet:
	go vet ./...

test:
	go test ./...

check: vet test generate build

clean:
	rm -f ./bin/hyprd ./bin/hyprctl

docker-build:
	$(DOCKER_COMPOSE) build

docker-run: docker-build
	$(DOCKER_COMPOSE) up -d

docker-stop:
	$(DOCKER_COMPOSE) down --remove-orphans

docker-status:
	$(DOCKER_COMPOSE) ps

docker-logs:
	$(DOCKER_COMPOSE) logs --tail=100 -f

docker-clean:
	$(DOCKER_COMPOSE) down --volumes --remove-orphans
