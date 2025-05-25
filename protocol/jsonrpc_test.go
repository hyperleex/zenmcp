package protocol

import (
	"encoding/json"
	"testing"
)

func TestRequestID_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"string", "test-id"},
		{"number", 123.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := NewRequestID(tt.input)
			
			data, err := json.Marshal(id)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			var id2 RequestID
			if err := json.Unmarshal(data, &id2); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			if id2.Value() != tt.input {
				t.Errorf("Expected %v, got %v", tt.input, id2.Value())
			}
		})
	}
}

func TestRequest_Marshal(t *testing.T) {
	req := Request{
		JSONRPC: JSONRPCVersion,
		ID:      NewRequestID("test"),
		Method:  "test_method",
		Params:  json.RawMessage(`{"key":"value"}`),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var req2 Request
	if err := json.Unmarshal(data, &req2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if req2.JSONRPC != JSONRPCVersion {
		t.Errorf("Expected JSONRPC %s, got %s", JSONRPCVersion, req2.JSONRPC)
	}

	if req2.Method != "test_method" {
		t.Errorf("Expected method test_method, got %s", req2.Method)
	}
}

func TestError_Error(t *testing.T) {
	err := NewError(InvalidRequest, "test error", "test data")
	
	expected := "JSON-RPC error -32600: test error"
	if err.Error() != expected {
		t.Errorf("Expected %s, got %s", expected, err.Error())
	}
}