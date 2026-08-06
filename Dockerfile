ARG GO_VERSION=1.24.5

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/hyprd ./cmd/hyprd && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/hyprctl ./cmd/hyprctl

FROM alpine:3.22

RUN addgroup -S hyperion && \
    adduser -S -G hyperion -h /home/hyperion hyperion && \
    mkdir -p /home/hyperion/.hyperion/data && \
    chown -R hyperion:hyperion /home/hyperion/.hyperion

COPY --from=builder /out/hyprd /out/hyprctl /usr/local/bin/

USER hyperion
WORKDIR /home/hyperion

EXPOSE 8080 8081 9001
CMD ["hyprd"]
