package chaos

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type kubernetesCluster struct {
	namespace string
}

func newKubernetesCluster(namespace string) *kubernetesCluster {
	if namespace == "" {
		namespace = "hyperion"
	}
	return &kubernetesCluster{namespace: namespace}
}

func (k *kubernetesCluster) Nodes() []string {
	nodes := make([]string, 5)
	for i := range 5 {
		nodes[i] = fmt.Sprintf("hyperion-%d", i)
	}
	return nodes
}

func (k *kubernetesCluster) Node(ctx context.Context, pod string, args ...string) (string, error) {
	command := []string{"exec", "-n", k.namespace, pod, "--", "hyprctl", "--addr", "http://" + pod + ".hyperion:8080", "--timeout", "2s"}
	return k.runInput(ctx, "", append(command, args...)...)
}

func (k *kubernetesCluster) Kill(ctx context.Context, pod string) error {
	_, err := k.runInput(ctx, "", "delete", "pod", pod, "-n", k.namespace, "--grace-period=0", "--force")
	return err
}

func (k *kubernetesCluster) Restart(ctx context.Context, pod string) error {
	_, err := k.runInput(ctx, "", "wait", "-n", k.namespace, "--for=condition=Ready", "pod/"+pod, "--timeout=120s")
	return err
}

func (k *kubernetesCluster) Partition(ctx context.Context, pod string) error {
	policy := fmt.Sprintf(`apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: hyperion-chaos-partition-%s
  namespace: %s
spec:
  podSelector:
    matchLabels:
      statefulset.kubernetes.io/pod-name: %s
  policyTypes:
    - Ingress
    - Egress
`, strings.TrimPrefix(pod, "hyperion-"), k.namespace, pod)
	_, err := k.runInput(ctx, policy, "apply", "-f", "-")
	return err
}

func (k *kubernetesCluster) Heal(ctx context.Context, pod string) error {
	name := "hyperion-chaos-partition-" + strings.TrimPrefix(pod, "hyperion-")
	_, err := k.runInput(ctx, "", "delete", "networkpolicy", name, "-n", k.namespace, "--ignore-not-found")
	return err
}

func (k *kubernetesCluster) runInput(ctx context.Context, input string, args ...string) (string, error) {
	var output bytes.Buffer
	command := exec.CommandContext(ctx, "kubectl", args...)
	command.Stdin = strings.NewReader(input)
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("kubectl %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(output.String()))
	}
	return strings.TrimSpace(output.String()), nil
}
