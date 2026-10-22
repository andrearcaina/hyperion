#!/bin/sh

set -eu

cluster_name=hyperion
image=hyperion:local
manifest_dir=deployments/k8s

case "${1:-}" in
    start)
        if ! kind get clusters | grep -qx "$cluster_name"; then
            kind create cluster --name "$cluster_name" --config "$manifest_dir/kind.yaml" --wait 60s
        fi

        ./scripts/docker.sh build
        kind load docker-image "$image" --name "$cluster_name"

        if kubectl get namespace hyperion >/dev/null 2>&1; then
            kubectl delete job hyperion-bootstrap -n hyperion --ignore-not-found
        fi

        kubectl apply -k "$manifest_dir"
        kubectl rollout status statefulset/hyperion -n hyperion --timeout=120s
        kubectl wait --for=condition=complete job/hyperion-bootstrap -n hyperion --timeout=120s
        ;;
    stop)
        kind delete cluster --name "$cluster_name"
        ;;
    status)
        kubectl get nodes
        kubectl get all -n hyperion -o wide
        ;;
    forward)
        kubectl port-forward -n hyperion pod/hyperion-0 8080:8080
        ;;
    logs)
        kubectl logs -n hyperion "${2:-statefulset/hyperion}" -f
        ;;
    *)
        echo "Usage: $0 {start|stop|status|forward|logs [pod]}" >&2
        exit 2
        ;;
esac
