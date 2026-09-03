.PHONY: generate build install imports tidy vet \
	test test-integration test-smoke test-chaos check \
	docker-config docker-build docker-run docker-reset docker-restart docker-stop docker-status docker-logs docker-clean \
	kube-start kube-stop kube-status kube-forward kube-test kube-smoke kube-logs

scenario ?= all

generate:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/hyperion/v1/hyperion.proto

build:
	go build -o ./bin/hyprd ./cmd/hyprd
	go build -o ./bin/hyprctl ./cmd/hyprctl

install: check
	go install ./cmd/hyprd ./cmd/hyprctl

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
	./scripts/docker.sh config

docker-build:
	./scripts/docker.sh build

docker-run:
	./scripts/docker.sh run

docker-reset:
	./scripts/docker.sh reset

docker-restart:
	./scripts/docker.sh restart

docker-stop:
	./scripts/docker.sh stop

docker-status:
	./scripts/docker.sh status

docker-logs:
	./scripts/docker.sh logs

docker-clean:
	./scripts/docker.sh clean

kube-start:
	./scripts/k8s.sh start

kube-stop:
	./scripts/k8s.sh stop

kube-status:
	./scripts/k8s.sh status

kube-forward:
	./scripts/k8s.sh forward

kube-smoke:
	./scripts/k8s.sh smoke

kube-logs:
	./scripts/k8s.sh logs $(pod)
