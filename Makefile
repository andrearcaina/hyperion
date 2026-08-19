.PHONY: build clean generate \
	test test-smoke test-chaos vet check \
	docker-config docker-build docker-run docker-stop docker-status docker-logs docker-clean

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

test-smoke:
	go test -tags=smoke ./internal/chaos

test-chaos:
	go run ./cmd/hyprchaos $(scenario)

check: vet test generate build

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
