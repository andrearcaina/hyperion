#!/bin/sh

set -eu

compose="docker compose -f deployments/docker/compose.yml"

case "${1:-}" in
    config)
        $compose config --quiet
        ;;
    build)
        $compose build
        ;;
    run)
        $compose build
        $compose up -d
        ;;
    reset)
        $compose down --volumes --remove-orphans
        $compose build
        $compose up -d
        ;;
    restart)
        $compose restart
        ;;
    stop)
        $compose down --remove-orphans
        ;;
    status)
        $compose ps
        ;;
    logs)
        $compose logs --tail=100 -f
        ;;
    clean)
        $compose down --volumes --remove-orphans
        ;;
    *)
        echo "Usage: $0 {config|build|run|reset|restart|stop|status|logs|clean}" >&2
        exit 2
        ;;
esac
