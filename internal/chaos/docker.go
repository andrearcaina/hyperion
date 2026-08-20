package chaos

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type dockerController struct{}

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
	compose := append([]string{"compose"}, args...)

	return d.run(ctx, compose...)
}

func (d *dockerController) node(ctx context.Context, service string, args ...string) (string, error) {
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

func (d *dockerController) kill(ctx context.Context, service string) error {
	_, err := d.compose(ctx, "kill", "-s", "SIGKILL", service) // same as doing docker compose kill -s 9 <service>
	return err
}

func (d *dockerController) start(ctx context.Context, service string) error {
	_, err := d.compose(ctx, "start", service)
	return err
}

func (d *dockerController) container(ctx context.Context, service string) (string, error) {
	return d.compose(ctx, "ps", "-q", service)
}

func (d *dockerController) network(ctx context.Context, container string) (string, error) {
	return d.run(ctx, "inspect", "--format",
		"{{range $name, $_ := .NetworkSettings.Networks}}{{$name}}{{end}}", container)
}

func (d *dockerController) disconnect(ctx context.Context, network, container string) error {
	_, err := d.run(ctx, "network", "disconnect", network, container)
	return err
}

func (d *dockerController) connect(ctx context.Context, network, container, alias string) error {
	_, err := d.run(ctx, "network", "connect", "--alias", alias, network, container)
	return err
}
