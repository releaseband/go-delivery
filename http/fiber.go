package http

import (
	"context"
	"fmt"
	"time"

	"github.com/releaseband/metrics/measure"

	"github.com/gofiber/fiber/v2"
)

type FiberClient struct {
	defaultTimeout time.Duration
}

func NewFiberClient(defaultTimeout time.Duration) FiberClient {
	return FiberClient{
		defaultTimeout: defaultTimeout,
	}
}

func makePostRequest(agent *fiber.Agent, url string) *fiber.Request {
	req := agent.Request()
	req.SetRequestURI(url)
	req.Header.SetMethod("POST")
	req.Header.SetContentType(applicationJson)

	return req
}

func makeAgent(body []byte) *fiber.Agent {
	return fiber.AcquireAgent().Body(body)
}

func send(ctx context.Context, agent *fiber.Agent, req *fiber.Request, timeout time.Duration) (int, []byte, error) {
	if err := agent.Parse(); err != nil {
		return 0, nil, fmt.Errorf("agent.Parse failed: %w", err)
	}

	resp := fiber.AcquireResponse()

	start := measure.Start()
	defer record(ctx, start)

	if err := agent.DoTimeout(req, resp, timeout); err != nil {
		err = parseError(err)

		return 0, nil, fmt.Errorf("fiber.DoTimeout: timeout=%s: err: %w",
			timeout.String(), err)
	}

	return resp.StatusCode(), resp.Body(), nil
}

func post(
	ctx context.Context,
	agent *fiber.Agent,
	req *fiber.Request,
	url string,
	timeout time.Duration,
) ([]byte, int, error) {
	ctx = wrapToLatencyContext(ctx, url)

	code, resp, err := send(ctx, agent, req, timeout)
	if err != nil {
		return nil, code, err
	}

	if code >= 400 {
		commitFailedHttpCode(ctx, url, code)
	}

	return resp, code, err
}

func isEmptyTimeout(timeout time.Duration) bool {
	return timeout == 0
}

func (f FiberClient) chooseTimeout(timeout time.Duration) time.Duration {
	if isEmptyTimeout(timeout) {
		return f.defaultTimeout
	}

	return timeout
}

func (f FiberClient) Post(
	ctx context.Context, url string, headers map[string]string, body []byte, timeout time.Duration) ([]byte, int, error) {
	agent := makeAgent(body)
	req := makePostRequest(agent, url)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	req.Header.Set("Content-Type", "application/json")

	return post(ctx, agent, req, url, f.chooseTimeout(timeout))
}
