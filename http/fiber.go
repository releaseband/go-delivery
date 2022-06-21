package http

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

type FiberClient struct {
	defaultTimeout time.Duration
}

func NewFiberClient() FiberClient {
	return FiberClient{}
}

func makePostRequest(agent *fiber.Agent, url string) *fiber.Request {
	req := agent.Request()
	req.SetRequestURI(url)
	req.Header.SetMethod("POST")

	return req
}

func makeAgent(body []byte) *fiber.Agent {
	return fiber.AcquireAgent().Body(body)
}

func send(ctx context.Context, agent *fiber.Agent, req *fiber.Request, timeout time.Duration) ([]byte, int, error) {
	if err := agent.Parse(); err != nil {
		return nil, 0, fmt.Errorf("agent.Parse failed: %w", err)
	}

	resp := fiber.AcquireResponse()

	start := time.Now()
	err := agent.DoTimeout(req, resp, timeout)
	record(ctx, start, err == nil)
	if err != nil {
		err = WrapKnownError(err)

		return nil, 0, fmt.Errorf("fiber.DoTimeout: timeout=%s: err: %w",
			timeout.String(), err)
	}

	return resp.Body(), resp.StatusCode(), nil
}

func (f FiberClient) Post(
	ctx context.Context, url string, headers map[string]string, body []byte, timeout time.Duration) ([]byte, int, error) {
	agent := makeAgent(body)
	req := makePostRequest(agent, url)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, code, err := send(ctx, agent, req, timeout)
	if err != nil {
		return nil, code, err
	}

	if isFailedHttpCode(code) {
		registerHttpStatus(ctx, code)
	}

	return resp, code, err
}
