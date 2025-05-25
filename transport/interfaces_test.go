package transport

import (
	"context"
	"testing"
)

type mockCodec struct {
	closed bool
}

func (m *mockCodec) Encode(v interface{}) error { return nil }
func (m *mockCodec) Decode(v interface{}) error { return nil }
func (m *mockCodec) Close() error {
	m.closed = true
	return nil
}

func TestNewConnection(t *testing.T) {
	ctx := context.Background()
	codec := &mockCodec{}
	
	conn := NewConnection(ctx, codec)
	
	if conn.Context() != ctx {
		t.Error("Expected same context")
	}
	
	if conn.Codec() != codec {
		t.Error("Expected same codec")
	}
	
	if err := conn.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	
	if !codec.closed {
		t.Error("Expected codec to be closed")
	}
}