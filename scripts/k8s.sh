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
    smoke)
        index=$(( $$ % 3 ))
        case "$index" in
            0) writer=hyperion-0; reader=hyperion-1 ;;
            1) writer=hyperion-1; reader=hyperion-2 ;;
            *) writer=hyperion-2; reader=hyperion-0 ;;
        esac
        key="k8s-smoke-$$"
        value=hello-from-k8s

        echo "Smoke test: writing through $writer and reading through $reader"
        result=$(kubectl exec -n hyperion "$writer" -- \
            hyprctl --addr "http://$writer.hyperion:8080" set "$key" "$value")
        actual=$(kubectl exec -n hyperion "$reader" -- \
            hyprctl --addr "http://$reader.hyperion:8080" get "$key")

        if [ "$result" != "$key=$value" ] || [ "$actual" != "$value" ]; then
            echo "Kubernetes smoke test failed: set=$result get=$actual" >&2
            exit 1
        fi

        kubectl exec -n hyperion "$writer" -- \
            hyprctl --addr "http://$writer.hyperion:8080" del "$key" >/dev/null
        echo "Kubernetes smoke test passed."
        ;;
    logs)
        kubectl logs -n hyperion "${2:-statefulset/hyperion}" -f
        ;;
    *)
        echo "Usage: $0 {start|stop|status|forward|test|smoke|logs [pod]}" >&2
        exit 2
        ;;
esac
