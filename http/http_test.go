package http

import (
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/valyala/fasthttp"
)

func Test_isConnRefused(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		if ok := IsConnRefused(errors.New("some error")); ok {
			t.Fatal("should be false")
		}
	})
	t.Run("true", func(t *testing.T) {
		if ok := IsConnRefused(fmt.Errorf("some error: %w", syscall.ECONNREFUSED)); !ok {
			t.Fatal("should be true")
		}
	})
}

func Test_isDeadLineExceededError(t *testing.T) {
	testCases := []struct {
		name string
		err  error
		ok   bool
	}{
		{
			name: "simple error",
			err:  errors.New("simple error"),
			ok:   false,
		},
		{
			name: "context deadline exceeded",
			err:  context.DeadlineExceeded,
			ok:   true,
		},
		{
			name: "os deadline exceeded",
			err:  os.ErrDeadlineExceeded,
			ok:   true,
		},
		{
			name: "fast http timeout",
			err:  fasthttp.ErrTimeout,
			ok:   true,
		},
		{
			name: "error contains 'timeout'",
			err:  errors.New("failed: context deadline exceeded"),
			ok:   true,
		},
		{
			name: "Dial timeout",
			err:  fmt.Errorf("failed:%w", fasthttp.ErrDialTimeout),
			ok:   true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ok := IsDeadlineExceededError(tt.err)
			if ok != tt.ok {
				t.Fatal("failed")
			}
		})
	}
}
