package chaos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type dockerController struct {
	networks map[string]string
}

const composeFile = "deployments/docker/compose.yml"

func newDockerCluster() *dockerController {
	return &dockerController{networks: make(map[string]string)}
}

func (d *dockerController) Nodes() []string {
	nodes := make([]string, 5)
	for i := range nodes {
		nodes[i] = fmt.Sprintf("node-%d", i+1)
	}
	return nodes
}

func (d *dockerController) Partition(ctx context.Context, service string) error {
	container, err := d.container(ctx, service)
	if err != nil {
		return err
	}
	container = strings.TrimSpace(container)
	if container == "" {
		return fmt.Errorf("no running container found for %s", service)
	}
	network, err := d.network(ctx, container)
	if err != nil {
		return err
	}
	d.networks[service] = network
	return d.disconnect(ctx, network, container)
}

func (d *dockerController) Heal(ctx context.Context, service string) error {
	container, err := d.container(ctx, service)
	if err != nil {
		return err
	}
	container = strings.TrimSpace(container)
	if container == "" {
		return fmt.Errorf("no running container found for %s", service)
	}
	network := d.networks[service]
	if network == "" {
		return fmt.Errorf("no remembered network for %s", service)
	}
	if err := d.connect(ctx, network, container, service); err != nil {
		return err
	}
	delete(d.networks, service)
	return nil
}

func (d *dockerController) run(ctx context.Context, args ...string) (string, error) {
	var output bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(
			"docker %s: %w: %s",
			strings.Join(args, " "),
			err,
			strings.TrimSpace(output.String()),
		)
	}

	return strings.TrimSpace(output.String()), nil
}

func (d *dockerController) compose(ctx context.Context, args ...string) (string, error) {
	compose := append([]string{"compose", "-f", composeFile}, args...)

	return d.run(ctx, compose...)
}

func (d *dockerController) Node(ctx context.Context, service string, args ...string) (string, error) {
	command := append([]string{
		"exec",
		"-T",
		service,
		"hyprctl",
		"--timeout",
		"2s",
	}, args...)

	return d.compose(ctx, command...)
}

func (d *dockerController) Kill(ctx context.Context, service string) error {
	_, err := d.compose(ctx, "kill", "-s", "SIGKILL", service) // same as doing docker compose kill -s 9 <service>
	return err
}

func (d *dockerController) Restart(ctx context.Context, service string) error {
	_, err := d.compose(ctx, "start", service)
	return err
}

func (d *dockerController) container(ctx context.Context, service string) (string, error) {
	return d.compose(ctx, "ps", "-q", service)
}

func (d *dockerController) network(ctx context.Context, container string) (string, error) {
	output, err := d.run(ctx, "inspect", "--format", "{{json .NetworkSettings.Networks}}", container)
	if err != nil {
		return "", err
	}
	return networkName(output, container)
}

func networkName(output, container string) (string, error) {
	networks := make(map[string]struct{})
	if err := json.Unmarshal([]byte(output), &networks); err != nil {
		return "", fmt.Errorf("parse networks for %s: %w", container, err)
	}
	for name := range networks {
		return name, nil
	}

	return "", fmt.Errorf("container %s has no attached networks", container)
}

func (d *dockerController) disconnect(ctx context.Context, network, container string) error {
	_, err := d.run(ctx, "network", "disconnect", network, container)
	return err
}

func (d *dockerController) connect(ctx context.Context, network, container, alias string) error {
	_, err := d.run(ctx, "network", "connect", "--alias", alias, network, container)
	return err
}
