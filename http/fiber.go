package http

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

type FiberClient struct{}

const (
	MethodPost   = "POST"
	MethodGet    = "GET"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
)

func NewFiberClient() FiberClient {
	return FiberClient{}
}

func (f FiberClient) Post(
	_ context.Context,
	url string,
	headers map[string]string,
	body []byte,
	timeout time.Duration,
) ([]byte, int, error) {
	return f.Do(url, headers, body, timeout)
}

func (f FiberClient) Do(
	url string,
	headers map[string]string,
	body []byte,
	timeout time.Duration,
) ([]byte, int, error) {
	agent := makeAgent(body)
	req := makeRequest(agent, MethodPost, url)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return send(agent, req, timeout)
}

func makeAgent(body []byte) *fiber.Agent {
	return fiber.AcquireAgent().Body(body)
}

func makeRequest(agent *fiber.Agent, method, url string) *fiber.Request {
	req := agent.Request()
	req.SetRequestURI(url)
	req.Header.SetMethod(method)

	return req
}

func send(agent *fiber.Agent, req *fiber.Request, timeout time.Duration) ([]byte, int, error) {
	if err := agent.Parse(); err != nil {
		return nil, 0, fmt.Errorf("agent.Parse failed: %w", err)
	}

	resp := fiber.AcquireResponse()

	if err := agent.DoTimeout(req, resp, timeout); err != nil {
		err = IsKnownError(err)

		return nil, 0, err
	}

	return resp.Body(), resp.StatusCode(), nil
}
