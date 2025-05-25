package stdio

import (
	"context"
	"testing"
	"time"
)

func TestTransport_Accept(t *testing.T) {
	transport := New()
	
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	conn, err := transport.Accept(ctx)
	if err != nil {
		t.Fatalf("Accept error: %v", err)
	}

	if conn == nil {
		t.Fatal("Expected connection, got nil")
	}

	if conn.Codec() == nil {
		t.Fatal("Expected codec, got nil")
	}

	if err := conn.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

func TestTransport_Close(t *testing.T) {
	transport := New()
	
	ctx := context.Background()
	_, err := transport.Accept(ctx)
	if err != nil {
		t.Fatalf("Accept error: %v", err)
	}

	if err := transport.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

func TestStdioReadWriteCloser(t *testing.T) {
	rw := &stdioReadWriteCloser{}
	
	// Test that Close doesn't return error
	if err := rw.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}