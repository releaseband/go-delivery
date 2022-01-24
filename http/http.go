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

const (
	applicationJson = "application/json"
)

var (
	ErrIntegrationConnectionTimeout = errors.New( "connection timed out for integration")
	ErrIntegrationConnectionRefused = errors.New( "integration connection refused")
)

type Client interface {
	Post(ctx context.Context, url string, headers map[string]string, body []byte, timeout time.Duration) ([]byte, int, error)
}

func isConnRefused(err error) bool {
	return errors.Is(err, syscall.ECONNREFUSED)
}

func isDeadlineExceededError(err error) bool {
	return errors.Is(err, fasthttp.ErrTimeout) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, os.ErrDeadlineExceeded) ||
		errors.Is(err, fasthttp.ErrDialTimeout) ||
		strings.Contains(err.Error(), "context deadline exceeded")
}

func parseError(err error) error {
	if isConnRefused(err) {
		err = ErrIntegrationConnectionRefused
	} else if isDeadlineExceededError(err) {
		err = ErrIntegrationConnectionTimeout
	}

	return err
}
