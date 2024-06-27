package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type serverStates struct {
	sync.RWMutex
	errors []error
}

func (s *serverStates) reset() {
	s.Lock()
	s.errors = make([]error, 0, 10)
	s.Unlock()
}

func (s *serverStates) addError(err error) {
	s.Lock()
	s.errors = append(s.errors, err)
	s.Unlock()
}

type testServer struct {
	state  *serverStates
	server *httptest.Server
}

func (s *testServer) addSimpleHandler(mux *http.ServeMux, h *postHandler) {
	mux.HandleFunc(h.url, func(writer http.ResponseWriter, request *http.Request) {
		code, err := h.handler()
		if err != nil {
			s.state.addError(err)
		}

		writer.WriteHeader(code)
	})
}

func (s *testServer) addTimedOutHandler(mux *http.ServeMux, url string, duration time.Duration) {
	mux.HandleFunc(url, func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(duration)
		writer.WriteHeader(200)
	})
}

func (s *testServer) addHandler(mux *http.ServeMux, h *postHandler) {
	if h.timeOut != nil {
		s.addTimedOutHandler(mux, h.url, h.timeOut.time)
	} else {
		s.addSimpleHandler(mux, h)
	}
}

func newTestServer(handlers ...*postHandler) (*testServer, func()) {
	s := &testServer{
		state: &serverStates{
			errors: make([]error, 0, 10),
		},
	}

	mux := http.NewServeMux()
	for _, h := range handlers {
		s.addHandler(mux, h)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		s.server = httptest.NewServer(mux)
		wg.Done()
	}()

	wg.Wait()

	for _, h := range handlers {
		h.fullPath = s.server.URL + h.url
	}

	return s, s.server.Close
}

func (s *testServer) Close() {
	s.server.Close()
}

type postHandler struct {
	url      string
	fullPath string
	handler  func() (int, error)
	timeOut  *struct {
		time time.Duration
	}
}

func newPostHandler(url string, handler func() (int, error)) *postHandler {
	return &postHandler{
		url:     url,
		handler: handler,
	}
}

func (h postHandler) FullPath() string {
	return h.fullPath
}

func (s *testServer) GetErrors() []error {
	return s.state.errors
}

func (s *testServer) Reset() {
	s.state.reset()
}

func Test_send(t *testing.T) {
	const (
		timeout = time.Millisecond * 200

		successUrl = "/test/success"
		failedUrl  = "/test/failed"
		timeoutUrl = "/test/timeout"
	)

	successHandler := newPostHandler(successUrl, func() (int, error) {
		return 200, nil
	})

	errFailed := errors.New("failed")
	failedHandler := newPostHandler(failedUrl, func() (int, error) {
		return 500, errFailed
	})

	timeoutHandler := newPostHandler(timeoutUrl, func() (int, error) {
		time.Sleep(timeout * 2)
		return 201, nil
	})

	server, finish := newTestServer(successHandler, failedHandler, timeoutHandler)
	defer finish()

	testCases := []struct {
		name      string
		url       string
		serverErr error
		clientErr error
		code      int
	}{
		{
			name:      "send failed",
			url:       failedHandler.FullPath(),
			serverErr: errFailed,
			code:      500,
		},
		{
			name:      "timeout",
			url:       timeoutHandler.FullPath(),
			serverErr: nil,
			clientErr: ErrIntegrationConnectionTimeout,
			code:      0,
		},
		{
			name:      "success",
			url:       successHandler.FullPath(),
			serverErr: nil,
			clientErr: nil,
			code:      200,
		},
	}

	body := []byte(`{body}`)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			agent := makeAgent(body)
			req := makeRequest(agent, MethodPost, tt.url)
			resp, code, err := send(agent, req, timeout)
			if !errors.Is(err, tt.clientErr) {
				t.Fatal("error invalid")
			}

			internalErrors := server.GetErrors()
			if tt.serverErr != nil {
				if len(internalErrors) != 1 {
					t.Fatal("errors len should be 1")
				}

				if !errors.Is(internalErrors[0], tt.serverErr) {
					t.Fatal("invalid internal error")
				}

				if len(resp) != 0 {
					t.Fatal("resp should be zero value or nil")
				}
			} else {
				if len(internalErrors) != 0 {
					t.Fatal("internal errors len should be zero value")
				}
			}

			if code != tt.code {
				t.Fatalf("httpStatus invalid: exp='%d' got='%d'", tt.code, code)
			}

			server.Reset()
		})
	}
}
