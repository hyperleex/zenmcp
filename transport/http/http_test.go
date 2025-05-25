package http

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	transport := New(Options{})
	if transport == nil {
		t.Fatal("expected transport to be created")
	}

	if transport.path != "/mcp" {
		t.Errorf("expected default path /mcp, got %s", transport.path)
	}

	if transport.server.Addr != ":8080" {
		t.Errorf("expected default addr :8080, got %s", transport.server.Addr)
	}
}

func TestNewWithOptions(t *testing.T) {
	opts := Options{
		Addr:         ":9090",
		Path:         "/api/mcp",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	transport := New(opts)
	if transport.path != "/api/mcp" {
		t.Errorf("expected path %s, got %s", opts.Path, transport.path)
	}

	if transport.server.Addr != ":9090" {
		t.Errorf("expected addr %s, got %s", opts.Addr, transport.server.Addr)
	}

	if transport.server.ReadTimeout != 10*time.Second {
		t.Errorf("expected read timeout %v, got %v", 10*time.Second, transport.server.ReadTimeout)
	}

	if transport.server.WriteTimeout != 15*time.Second {
		t.Errorf("expected write timeout %v, got %v", 15*time.Second, transport.server.WriteTimeout)
	}
}

func TestTransportAccept(t *testing.T) {
	// Find a free port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	transport := New(Options{
		Addr: ":" + string(rune(port+'0')),
	})
	defer transport.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Accept should timeout since no connections are made
	_, err = transport.Accept(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("expected timeout error, got %v", err)
	}
}

func TestTransportClose(t *testing.T) {
	transport := New(Options{})
	err := transport.Close()
	if err != nil {
		t.Errorf("close failed: %v", err)
	}
}

func TestReadWriteCloser(t *testing.T) {
	data := []byte("test data")
	rw := &readWriteCloser{
		reader: &testReader{data: data},
		writer: &testWriter{},
	}

	// Test read
	buf := make([]byte, len(data))
	n, err := rw.Read(buf)
	if err != nil {
		t.Errorf("read failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected %d bytes read, got %d", len(data), n)
	}

	// Test write
	n, err = rw.Write(data)
	if err != nil {
		t.Errorf("write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected %d bytes written, got %d", len(data), n)
	}

	// Test close
	err = rw.Close()
	if err != nil {
		t.Errorf("close failed: %v", err)
	}
}

type testReader struct {
	data []byte
	pos  int
}

func (r *testReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, nil
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *testReader) Close() error {
	return nil
}

type testWriter struct {
	data []byte
}

func (w *testWriter) Write(p []byte) (n int, err error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w *testWriter) Close() error {
	return nil
}