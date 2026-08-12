package chaos

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type dockerCLI struct{}

func (d *dockerCLI) run(ctx context.Context, args ...string) (string, error) {
	var output bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker %s: %w: %s",
			strings.Join(args, " "),
			err,
			strings.TrimSpace(output.String()),
		)
	}
	return strings.TrimSpace(output.String()), nil
}

func (d *dockerCLI) compose(ctx context.Context, args ...string) (string, error) {
	return d.run(ctx, append([]string{"compose"}, args...)...)
}

func (d *dockerCLI) node(ctx context.Context, service string, args ...string) (string, error) {
	command := []string{
		"exec",
		"-T",
		service,
		"hyprctl",
		"--timeout",
		"2s",
	}

	return d.compose(ctx, append(command, args...)...)
}

func (d *dockerCLI) container(ctx context.Context, service string) (string, error) {
	return d.compose(ctx, "ps", "-q", service)
}

func (d *dockerCLI) network(ctx context.Context, container string) (string, error) {
	return d.run(ctx, "inspect", "--format",
		"{{range $name, $_ := .NetworkSettings.Networks}}{{$name}}{{end}}", container)
}

func (d *dockerCLI) disconnect(ctx context.Context, network, container string) error {
	_, err := d.run(ctx, "network", "disconnect", network, container)
	return err
}

func (d *dockerCLI) connect(ctx context.Context, network, container, alias string) error {
	_, err := d.run(ctx, "network", "connect", "--alias", alias, network, container)
	return err
}
