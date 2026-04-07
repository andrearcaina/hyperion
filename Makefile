.PHONY: build clean

build:
	go build -o ./bin/hyprd ./cmd/hyprd
	go build -o ./bin/hyprctl ./cmd/hyprctl

clean:
	rm -f ./bin/hyprd ./bin/hyprctl
