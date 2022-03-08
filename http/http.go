package http

import (
	"context"
	"errors"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	ErrIntegrationConnectionTimeout = errors.New("connection timed out for integration")
	ErrIntegrationConnectionRefused = errors.New("integration connection refused")
)

type Client interface {
	Post(ctx context.Context, url string, headers map[string]string, body []byte, timeout time.Duration) ([]byte, int, error)
}

func IsConnRefused(err error) bool {
	return errors.Is(err, syscall.ECONNREFUSED)
}

func IsDeadlineExceededError(err error) bool {
	return errors.Is(err, fasthttp.ErrTimeout) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, os.ErrDeadlineExceeded) ||
		errors.Is(err, fasthttp.ErrDialTimeout) ||
		strings.Contains(err.Error(), "context deadline exceeded")
}

func IsKnownError(err error) error {
	if IsConnRefused(err) {
		err = ErrIntegrationConnectionRefused
	} else if IsDeadlineExceededError(err) {
		err = ErrIntegrationConnectionTimeout
	}

	return err
}
