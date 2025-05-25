package http

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/hyperleex/zenmcp/protocol"
	"github.com/hyperleex/zenmcp/transport"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	useSSE     bool
}

type ClientOptions struct {
	BaseURL    string
	HTTPClient *http.Client
	UseSSE     bool
}

func NewClient(opts ClientOptions) *Client {
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{}
	}
	if opts.BaseURL == "" {
		opts.BaseURL = "http://localhost:8080/mcp"
	}

	return &Client{
		baseURL:    opts.BaseURL,
		httpClient: opts.HTTPClient,
		useSSE:     opts.UseSSE,
	}
}

func (c *Client) Connect(ctx context.Context) (transport.Connection, error) {
	if c.useSSE {
		return c.connectSSE(ctx)
	}
	return c.connectHTTP(ctx)
}

func (c *Client) connectSSE(ctx context.Context) (transport.Connection, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	// Create SSE connection
	conn := &sseConnection{
		ctx:    ctx,
		resp:   resp,
		reader: &sseReader{body: resp.Body},
		writer: &sseClientWriter{client: c.httpClient, baseURL: c.baseURL},
	}

	conn.codec = protocol.NewJSONCodec(conn)
	return conn, nil
}

func (c *Client) connectHTTP(ctx context.Context) (transport.Connection, error) {
	conn := &httpClientConnection{
		ctx:        ctx,
		client:     c.httpClient,
		baseURL:    c.baseURL,
		requestBuf: &bytes.Buffer{},
	}

	conn.codec = protocol.NewJSONCodec(conn)
	return conn, nil
}

func (c *Client) Close() error {
	return nil
}

type sseConnection struct {
	ctx    context.Context
	resp   *http.Response
	codec  protocol.Codec
	reader *sseReader
	writer *sseClientWriter
}

func (c *sseConnection) Read(p []byte) (n int, err error) {
	return c.reader.Read(p)
}

func (c *sseConnection) Write(p []byte) (n int, err error) {
	return c.writer.Write(p)
}

func (c *sseConnection) Close() error {
	if c.resp != nil {
		return c.resp.Body.Close()
	}
	return nil
}

func (c *sseConnection) Codec() protocol.Codec {
	return c.codec
}

func (c *sseConnection) Context() context.Context {
	return c.ctx
}

type sseReader struct {
	body   interface{ Read([]byte) (int, error) }
	buffer []byte
}

func (r *sseReader) Read(p []byte) (n int, err error) {
	// Simplified SSE parsing - in production would need proper event parsing
	return r.body.Read(p)
}

type sseClientWriter struct {
	client  *http.Client
	baseURL string
}

func (w *sseClientWriter) Write(p []byte) (n int, err error) {
	// Send data back to server via separate HTTP request
	req, err := http.NewRequest(http.MethodPost, w.baseURL, bytes.NewReader(p))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	
	resp, err := w.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("server error: %d", resp.StatusCode)
	}

	return len(p), nil
}

type httpClientConnection struct {
	ctx        context.Context
	client     *http.Client
	baseURL    string
	codec      protocol.Codec
	requestBuf *bytes.Buffer
	response   *http.Response
}

func (c *httpClientConnection) Read(p []byte) (n int, err error) {
	if c.response == nil {
		return 0, fmt.Errorf("no response available")
	}
	return c.response.Body.Read(p)
}

func (c *httpClientConnection) Write(p []byte) (n int, err error) {
	// Buffer the request
	return c.requestBuf.Write(p)
}

func (c *httpClientConnection) Close() error {
	if c.response != nil {
		return c.response.Body.Close()
	}
	return nil
}

func (c *httpClientConnection) Codec() protocol.Codec {
	return c.codec
}

func (c *httpClientConnection) Context() context.Context {
	return c.ctx
}

func (c *httpClientConnection) Flush() error {
	if c.requestBuf.Len() == 0 {
		return nil
	}

	req, err := http.NewRequestWithContext(c.ctx, http.MethodPost, c.baseURL, bytes.NewReader(c.requestBuf.Bytes()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	c.response = resp
	c.requestBuf.Reset()

	return nil
}