package http

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hyperleex/zenmcp/protocol"
	"github.com/hyperleex/zenmcp/transport"
)

type Transport struct {
	server   *http.Server
	listener net.Listener
	path     string
	mu       sync.RWMutex
	conns    map[*httpConnection]struct{}
	connChan chan transport.Connection
}

type Options struct {
	Addr         string
	Path         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func New(opts Options) *Transport {
	if opts.Addr == "" {
		opts.Addr = ":8080"
	}
	if opts.Path == "" {
		opts.Path = "/mcp"
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = 30 * time.Second
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = 30 * time.Second
	}

	t := &Transport{
		path:  opts.Path,
		conns: make(map[*httpConnection]struct{}),
	}

	mux := http.NewServeMux()
	mux.HandleFunc(opts.Path, t.handleMCP)

	t.server = &http.Server{
		Addr:         opts.Addr,
		Handler:      mux,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
	}

	return t
}

func (t *Transport) Accept(ctx context.Context) (transport.Connection, error) {
	if t.listener == nil {
		listener, err := net.Listen("tcp", t.server.Addr)
		if err != nil {
			return nil, fmt.Errorf("failed to listen: %w", err)
		}
		t.listener = listener

		go func() {
			if err := t.server.Serve(t.listener); err != nil && err != http.ErrServerClosed {
				// Log error but don't block
			}
		}()
	}

	// For HTTP transport, we use a channel-based approach to bridge 
	// the Accept pattern with HTTP request handling
	connChan := make(chan transport.Connection, 1)
	
	t.mu.Lock()
	t.connChan = connChan
	t.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case conn := <-connChan:
		return conn, nil
	}
}

func (t *Transport) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if client supports Server-Sent Events
	accept := r.Header.Get("Accept")
	supportsSSE := strings.Contains(accept, "text/event-stream")

	if supportsSSE {
		t.handleSSE(w, r)
	} else {
		t.handleRegularHTTP(w, r)
	}
}

func (t *Transport) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Create bidirectional stream
	conn := newHTTPConnection(r.Context(), r.Body, &sseWriter{w: w, flusher: flusher})
	
	t.mu.Lock()
	t.conns[conn] = struct{}{}
	if t.connChan != nil {
		select {
		case t.connChan <- conn:
		default:
		}
	}
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		delete(t.conns, conn)
		t.mu.Unlock()
		conn.Close()
	}()

	// Keep connection alive until context is done
	<-r.Context().Done()
}

func (t *Transport) handleRegularHTTP(w http.ResponseWriter, r *http.Request) {
	conn := newHTTPConnection(r.Context(), r.Body, &httpWriter{w: w})
	
	t.mu.Lock()
	if t.connChan != nil {
		select {
		case t.connChan <- conn:
		default:
		}
	}
	t.mu.Unlock()

	defer conn.Close()

	// For regular HTTP, we expect a single request-response
	// The server will handle this connection through Accept()
}

func (t *Transport) Close() error {
	t.mu.Lock()
	for conn := range t.conns {
		conn.Close()
	}
	t.mu.Unlock()

	if t.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return t.server.Shutdown(ctx)
	}
	return nil
}

type httpConnection struct {
	ctx   context.Context
	codec protocol.Codec
	rw    io.ReadWriteCloser
}

func newHTTPConnection(ctx context.Context, reader io.Reader, writer io.Writer) *httpConnection {
	rw := &readWriteCloser{
		reader: reader,
		writer: writer,
	}
	
	return &httpConnection{
		ctx:   ctx,
		codec: protocol.NewJSONCodec(rw),
		rw:    rw,
	}
}

func (c *httpConnection) Codec() protocol.Codec {
	return c.codec
}

func (c *httpConnection) Context() context.Context {
	return c.ctx
}

func (c *httpConnection) Close() error {
	return c.rw.Close()
}

type readWriteCloser struct {
	reader io.Reader
	writer io.Writer
}

func (rw *readWriteCloser) Read(p []byte) (n int, err error) {
	return rw.reader.Read(p)
}

func (rw *readWriteCloser) Write(p []byte) (n int, err error) {
	return rw.writer.Write(p)
}

func (rw *readWriteCloser) Close() error {
	if closer, ok := rw.reader.(io.Closer); ok {
		closer.Close()
	}
	if closer, ok := rw.writer.(io.Closer); ok {
		closer.Close()
	}
	return nil
}

type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func (s *sseWriter) Write(p []byte) (n int, err error) {
	// Write as SSE data event
	fmt.Fprintf(s.w, "data: %s\n\n", string(p))
	s.flusher.Flush()
	return len(p), nil
}

type httpWriter struct {
	w http.ResponseWriter
}

func (h *httpWriter) Write(p []byte) (n int, err error) {
	return h.w.Write(p)
}