package herdrsock

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"
)

const (
	// CurrentProtocol is the Herdr socket API protocol version in the
	// Herdr 0.7.3 source used for this package.
	CurrentProtocol uint32 = 16

	defaultTimeout = 30 * time.Second
)

var requestCounter atomic.Uint64

// Client talks to Herdr's newline-delimited JSON socket API.
//
// The client opens one connection per Call, matching Herdr's own CLI client.
// Long-lived event subscriptions use Subscribe and keep their connection open.
type Client struct {
	socketPath string
	timeout    time.Duration
}

// Option configures a Client.
type Option func(*clientConfig)

type clientConfig struct {
	socketPath string
	session    *string
	timeout    time.Duration
}

// New returns a client using Herdr's socket resolution order:
// explicit WithSocketPath, explicit WithSession, HERDR_SOCKET_PATH,
// HERDR_SESSION, then the default session socket.
func New(opts ...Option) (*Client, error) {
	cfg := clientConfig{timeout: defaultTimeout}
	for _, opt := range opts {
		opt(&cfg)
	}

	socketPath := cfg.socketPath
	if socketPath == "" {
		if cfg.session != nil {
			path, err := SocketPathForSession(*cfg.session)
			if err != nil {
				return nil, err
			}
			socketPath = path
		} else {
			var err error
			socketPath, err = ActiveSocketPath()
			if err != nil {
				return nil, err
			}
		}
	}

	return &Client{
		socketPath: socketPath,
		timeout:    cfg.timeout,
	}, nil
}

// MustNew is a convenience for package-level clients in small tools.
func MustNew(opts ...Option) *Client {
	client, err := New(opts...)
	if err != nil {
		panic(err)
	}
	return client
}

// WithSocketPath targets an explicit Herdr API socket path.
func WithSocketPath(path string) Option {
	return func(cfg *clientConfig) {
		cfg.socketPath = path
	}
}

// WithSession targets a named Herdr session. Use an empty string or "default"
// for the default session.
func WithSession(name string) Option {
	return func(cfg *clientConfig) {
		cfg.session = &name
	}
}

// WithTimeout sets the per-request connect/write/read timeout. A zero timeout
// disables the client timeout and leaves cancellation to ctx.
func WithTimeout(timeout time.Duration) Option {
	return func(cfg *clientConfig) {
		cfg.timeout = timeout
	}
}

// SocketPath returns the resolved Herdr API socket path.
func (c *Client) SocketPath() string {
	return c.socketPath
}

// Call sends one request and decodes the response result into result.
func (c *Client) Call(ctx context.Context, id, method string, params any, result any) error {
	response, err := c.CallRaw(ctx, id, method, params)
	if err != nil {
		return err
	}
	if result == nil {
		return nil
	}
	if len(response.Result) == 0 || string(response.Result) == "null" {
		return nil
	}
	if err := json.Unmarshal(response.Result, result); err != nil {
		return fmt.Errorf("decode %s result: %w", method, err)
	}
	return nil
}

// CallRaw sends one request and returns the raw JSON result payload.
func (c *Client) CallRaw(ctx context.Context, id, method string, params any) (*Response, error) {
	if method == "" {
		return nil, errors.New("herdrsock: method is required")
	}
	if id == "" {
		id = nextRequestID()
	}

	conn, err := c.dial(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if deadline, ok := c.deadline(ctx); ok {
		_ = conn.SetDeadline(deadline)
	}

	request := Request{
		ID:     id,
		Method: method,
		Params: normalizeParams(params),
	}
	if err := writeJSONLine(conn, request); err != nil {
		return nil, err
	}

	var response Response
	if err := readJSONLine(bufio.NewReader(conn), &response); err != nil {
		return nil, err
	}
	if response.ID != id {
		return nil, &IDMismatchError{Want: id, Got: response.ID}
	}
	if response.Error != nil {
		return nil, &ErrorResponse{ID: response.ID, Body: *response.Error}
	}
	return &response, nil
}

// Ping returns the running Herdr server version, protocol, and capabilities.
func (c *Client) Ping(ctx context.Context) (*PongResult, error) {
	var pong PongResult
	if err := c.Call(ctx, "", MethodPing, EmptyParams{}, &pong); err != nil {
		return nil, err
	}
	return &pong, nil
}

// RequireProtocol checks the running server protocol and returns
// ProtocolMismatchError when it differs from required.
func (c *Client) RequireProtocol(ctx context.Context, required uint32) (*PongResult, error) {
	pong, err := c.Ping(ctx)
	if err != nil {
		return nil, err
	}
	if pong.Protocol != required {
		return pong, &ProtocolMismatchError{
			Required: required,
			Actual:   pong.Protocol,
			Version:  pong.Version,
		}
	}
	return pong, nil
}

// RequireCurrentProtocol checks the server against this package's
// CurrentProtocol constant.
func (c *Client) RequireCurrentProtocol(ctx context.Context) (*PongResult, error) {
	return c.RequireProtocol(ctx, CurrentProtocol)
}

// StopServer asks the running Herdr server to stop.
func (c *Client) StopServer(ctx context.Context) error {
	return c.Call(ctx, "", MethodServerStop, EmptyParams{}, nil)
}

// ReloadConfig reloads Herdr config.
func (c *Client) ReloadConfig(ctx context.Context) (*ConfigReloadResult, error) {
	var out ConfigReloadResult
	if err := c.Call(ctx, "", MethodServerReloadConfig, EmptyParams{}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) dial(ctx context.Context) (net.Conn, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	dialCtx, cancel := c.contextWithTimeout(ctx)
	defer cancel()

	conn, err := dialLocal(dialCtx, c.socketPath)
	if err != nil {
		return nil, fmt.Errorf("connect herdr socket %q: %w", c.socketPath, err)
	}
	return conn, nil
}

func (c *Client) contextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.timeout <= 0 {
		return context.WithCancel(ctx)
	}
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, c.timeout)
}

func (c *Client) deadline(ctx context.Context) (time.Time, bool) {
	now := time.Now()
	var deadline time.Time
	if c.timeout > 0 {
		deadline = now.Add(c.timeout)
	}
	if ctx != nil {
		if ctxDeadline, ok := ctx.Deadline(); ok && (deadline.IsZero() || ctxDeadline.Before(deadline)) {
			deadline = ctxDeadline
		}
	}
	if deadline.IsZero() {
		return time.Time{}, false
	}
	return deadline, true
}

func writeJSONLine(w io.Writer, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	n, err := w.Write(encoded)
	if err == nil && n != len(encoded) {
		return io.ErrShortWrite
	}
	return err
}

func readJSONLine(r *bufio.Reader, value any) error {
	line, err := r.ReadBytes('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && len(line) == 0 {
			return ErrEmptyResponse
		}
		if len(line) == 0 {
			return err
		}
	}
	if len(line) == 0 {
		return ErrEmptyResponse
	}
	if err := json.Unmarshal(line, value); err != nil {
		return err
	}
	return nil
}

func normalizeParams(params any) any {
	if params == nil {
		return EmptyParams{}
	}
	return params
}

func nextRequestID() string {
	return fmt.Sprintf("go:%d", requestCounter.Add(1))
}
