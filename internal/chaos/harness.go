package chaos

import (
	"context"
	"time"
)

type Harness struct {
	docker   *dockerController
	attempts int
	delay    time.Duration
}

func NewHarness() *Harness {
	return &Harness{
		docker:   &dockerController{},
		attempts: 30,
		delay:    time.Second,
	}
}

func (h *Harness) retry(ctx context.Context, action func() error) error {
	var err error
	for range h.attempts {
		if err = action(); err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(h.delay):
		}
	}
	return err
}
