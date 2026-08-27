.PHONY: generate build imports tidy vet \
	test test-integration test-smoke test-chaos check \
	docker-config docker-build docker-run docker-reset docker-restart docker-stop docker-status docker-logs docker-clean

scenario ?= all
DOCKER_COMPOSE = docker compose

generate:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/hyperion/v1/hyperion.proto

build:
	go build -o ./bin/hyprd ./cmd/hyprd
	go build -o ./bin/hyprctl ./cmd/hyprctl

imports:
	goimports -w .

tidy:
	go mod tidy

vet:
	go vet ./...

test:
	go test ./...

test-integration:
	go test -tags=integration ./internal/test/integration

test-smoke:
	go test -tags=smoke ./internal/test/chaos

test-chaos:
	go run ./cmd/hyprchaos $(scenario)

check: tidy generate imports vet test build

clean:
	rm -f ./bin/hyprd ./bin/hyprctl

docker-config:
	$(DOCKER_COMPOSE) config --quiet

docker-build:
	$(DOCKER_COMPOSE) build

docker-run: docker-build
	$(DOCKER_COMPOSE) up -d

docker-reset: docker-clean docker-run

docker-restart:
	$(DOCKER_COMPOSE) restart

docker-stop:
	$(DOCKER_COMPOSE) down --remove-orphans

docker-status:
	$(DOCKER_COMPOSE) ps

docker-logs:
	$(DOCKER_COMPOSE) logs --tail=100 -f

docker-clean:
	$(DOCKER_COMPOSE) down --volumes --remove-orphans
