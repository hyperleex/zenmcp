package protocol

import (
	"bytes"
	"testing"
)

type testReadWriteCloser struct {
	*bytes.Buffer
	closed bool
}

func (t *testReadWriteCloser) Close() error {
	t.closed = true
	return nil
}

func TestJSONCodec(t *testing.T) {
	buf := &testReadWriteCloser{Buffer: &bytes.Buffer{}}
	codec := NewJSONCodec(buf)

	req := Request{
		JSONRPC: JSONRPCVersion,
		ID:      NewRequestID("test"),
		Method:  "test_method",
	}

	if err := codec.Encode(req); err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	var decoded Request
	if err := codec.Decode(&decoded); err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	if decoded.Method != "test_method" {
		t.Errorf("Expected method test_method, got %s", decoded.Method)
	}

	if err := codec.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	if !buf.closed {
		t.Error("Expected buffer to be closed")
	}
}

func TestLengthPrefixedCodec(t *testing.T) {
	buf := &testReadWriteCloser{Buffer: &bytes.Buffer{}}
	codec := NewLengthPrefixedCodec(buf)

	req := Request{
		JSONRPC: JSONRPCVersion,
		ID:      NewRequestID("test"),
		Method:  "test_method",
	}

	if err := codec.Encode(req); err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	var decoded Request
	if err := codec.Decode(&decoded); err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	if decoded.Method != "test_method" {
		t.Errorf("Expected method test_method, got %s", decoded.Method)
	}
}

func TestLengthPrefixedCodec_InvalidContentLength(t *testing.T) {
	buf := bytes.NewBufferString("Content-Length: invalid\r\n\r\n")
	codec := NewLengthPrefixedCodec(&testReadWriteCloser{Buffer: buf})

	var req Request
	err := codec.Decode(&req)
	if err == nil {
		t.Error("Expected error for invalid Content-Length")
	}
}

func TestLengthPrefixedCodec_MissingContentLength(t *testing.T) {
	buf := bytes.NewBufferString("\r\n\r\n")
	codec := NewLengthPrefixedCodec(&testReadWriteCloser{Buffer: buf})

	var req Request
	err := codec.Decode(&req)
	if err == nil {
		t.Error("Expected error for missing Content-Length")
	}
}