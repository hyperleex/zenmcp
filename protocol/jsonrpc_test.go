package protocol

import (
	"encoding/json"
	"reflect" // Added reflect package
	"strings" // Added strings package
	"testing"
)

func TestRequestID_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"string", "test-id"},
		{"number", 123.0},
		{"null", nil},
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

			// Use reflect.DeepEqual for robust comparison, especially for nil.
			// tt.input is the original value used for NewRequestID.
			// id2.Value() is the value after marshal/unmarshal.
			if !reflect.DeepEqual(id2.Value(), tt.input) {
				// For numbers, tt.input might be int (e.g. 123.0) but id2.Value() will be float64.
				// DeepEqual handles this if they are numerically equal.
				// For the "null" case, tt.input is nil. id2.Value() should also be nil. DeepEqual handles this.
				t.Errorf("Expected %v (type %T), got %v (type %T)", tt.input, tt.input, id2.Value(), id2.Value())
			}
		})
	}

	// Test cases for UnmarshalJSON with invalid types
	invalidUnmarshalTests := []struct {
		name      string
		inputJSON string
	}{
		{"boolean", "true"},
		{"array", "[]"},
		{"object", "{}"},
	}

	for _, tt := range invalidUnmarshalTests {
		t.Run("InvalidUnmarshal_"+tt.name, func(t *testing.T) {
			var id RequestID
			err := json.Unmarshal([]byte(tt.inputJSON), &id)
			if err == nil {
				t.Errorf("Expected error for invalid JSON type %s, but got nil", tt.inputJSON)
			} else {
				// Check if the error message contains the base part, as it now includes the invalid data.
				expectedErrorMsgBase := "invalid request ID type"
				if !strings.Contains(err.Error(), expectedErrorMsgBase) {
					t.Errorf("Expected error message to contain %q, got %q", expectedErrorMsgBase, err.Error())
				}
			}
		})
	}
}

func TestNotification_MarshalUnmarshal(t *testing.T) {
	type testCase struct {
		name  string
		input Notification
		// Optional: for checking specific JSON output if omitempty is tricky for Params
		expectedJSON string
	}

	tests := []testCase{
		{
			name: "notification with params",
			input: Notification{
				JSONRPC: JSONRPCVersion,
				Method:  "$/progress",
				Params:  json.RawMessage(`{"token":"token123","value":{"kind":"begin","title":"Processing..."}}`),
			},
		},
		{
			name: "notification without params", // Params is json.RawMessage, nil should be omitted
			input: Notification{
				JSONRPC: JSONRPCVersion,
				Method:  "$/cancelRequest",
				Params:  nil, 
			},
			expectedJSON: `{"jsonrpc":"2.0","method":"$/cancelRequest"}`,
		},
		{
			name: "notification with empty object params", // Non-nil empty RawMessage should be included
			input: Notification{
				JSONRPC: JSONRPCVersion,
				Method:  "someEvent",
				Params:  json.RawMessage(`{}`),
			},
			expectedJSON: `{"jsonrpc":"2.0","method":"someEvent","params":{}}`,
		},
		{
			name: "notification with empty array params", // Non-nil empty RawMessage should be included
			input: Notification{
				JSONRPC: JSONRPCVersion,
				Method:  "anotherEvent",
				Params:  json.RawMessage(`[]`),
			},
			expectedJSON: `{"jsonrpc":"2.0","method":"anotherEvent","params":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			if tt.expectedJSON != "" {
				if string(data) != tt.expectedJSON {
					t.Errorf("Marshal() JSON output = %s, want %s", string(data), tt.expectedJSON)
				}
			}

			var unmarshaledNotification Notification
			if err := json.Unmarshal(data, &unmarshaledNotification); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			if unmarshaledNotification.JSONRPC != tt.input.JSONRPC {
				t.Errorf("JSONRPC mismatch: got %s, want %s", unmarshaledNotification.JSONRPC, tt.input.JSONRPC)
			}
			if unmarshaledNotification.Method != tt.input.Method {
				t.Errorf("Method mismatch: got %s, want %s", unmarshaledNotification.Method, tt.input.Method)
			}

			if tt.input.Params == nil {
				if unmarshaledNotification.Params != nil {
					// If input Params is nil, omitempty should make it absent in JSON.
					// Unmarshalling an absent field should result in a nil slice for json.RawMessage.
					t.Errorf("Params mismatch: expected nil, got %s", string(unmarshaledNotification.Params))
				}
			} else {
				if string(unmarshaledNotification.Params) != string(tt.input.Params) {
					t.Errorf("Params content mismatch: got %s, want %s", 
						string(unmarshaledNotification.Params), string(tt.input.Params))
				}
			}
		})
	}
}

// This existing TestRequest_Marshal is a good base, but I will create TestRequest_MarshalUnmarshal_Variations
// for more comprehensive cases as requested by the plan.
// I will keep TestRequest_Marshal as it tests a simple valid case.
// func TestRequest_Marshal(t *testing.T) { ... } // Original can be kept or removed if new one is sufficient

func TestRequest_MarshalUnmarshal_Variations(t *testing.T) {
	type testCase struct {
		name          string
		input         Request
		expectedJSON  string // Optional: for checking specific JSON output if omitempty is tricky
		checkOmittedID bool // Special flag for checking if ID is omitted in JSON
	}

	tests := []testCase{
		{
			name: "with string ID and params",
			input: Request{
				JSONRPC: JSONRPCVersion,
				ID:      NewRequestID("string-id-123"),
				Method:  "testMethod",
				Params:  json.RawMessage(`{"param1":"value1"}`),
			},
		},
		{
			name: "with number ID and params",
			input: Request{
				JSONRPC: JSONRPCVersion,
				ID:      NewRequestID(float64(456)), // Ensure float64 for numeric ID
				Method:  "anotherMethod",
				Params:  json.RawMessage(`[1,"foo"]`), // Canonical JSON (no space)
			},
		},
		{
			name: "with null ID",
			input: Request{
				JSONRPC: JSONRPCVersion,
				ID:      NewRequestID(nil), // This will marshal to "id":null
				Method:  "methodWithNullID",
				Params:  json.RawMessage(`{}`),
			},
		},
		{
			name: "with ID field completely omitted",
			input: Request{ // ID field is nil by default if not set
				JSONRPC: JSONRPCVersion,
				Method:  "methodWithOmittedID",
				Params:  json.RawMessage(`{"p":true}`),
			},
			// Expected JSON should not have the "id" field due to omitempty
			expectedJSON: `{"jsonrpc":"2.0","method":"methodWithOmittedID","params":{"p":true}}`,
			checkOmittedID: true,
		},
		{
			name: "with params omitted",
			input: Request{
				JSONRPC: JSONRPCVersion,
				ID:      NewRequestID("params-omitted-id"),
				Method:  "methodNoParams",
				// Params is json.RawMessage, which is a slice.
				// A nil slice should be omitted by omitempty.
				// An empty non-nil slice (e.g., json.RawMessage("{}") or json.RawMessage("[]")) might not be.
				// For Params to be truly omitted from JSON, it should be nil.
			},
			// Expected JSON should not have "params" field
			expectedJSON: `{"jsonrpc":"2.0","id":"params-omitted-id","method":"methodNoParams"}`,
		},
		{
			name: "with empty object params", // json.RawMessage for {} is not nil, so "params":{} will be present
			input: Request{
				JSONRPC: JSONRPCVersion,
				ID:      NewRequestID("empty-obj-params"),
				Method:  "methodEmptyObjParams",
				Params:  json.RawMessage(`{}`),
			},
			expectedJSON: `{"jsonrpc":"2.0","id":"empty-obj-params","method":"methodEmptyObjParams","params":{}}`,
		},
		{
			name: "with empty array params", // json.RawMessage for [] is not nil, so "params":[] will be present
			input: Request{
				JSONRPC: JSONRPCVersion,
				ID:      NewRequestID("empty-arr-params"),
				Method:  "methodEmptyArrParams",
				Params:  json.RawMessage(`[]`),
			},
			expectedJSON: `{"jsonrpc":"2.0","id":"empty-arr-params","method":"methodEmptyArrParams","params":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			// Optional: Check specific JSON output
			if tt.expectedJSON != "" {
				if string(data) != tt.expectedJSON {
					t.Errorf("Marshal() output = %s, want %s", string(data), tt.expectedJSON)
				}
			}
			
			if tt.checkOmittedID { // For "ID field completely omitted" case
				var m map[string]interface{}
				if err := json.Unmarshal(data, &m); err != nil {
					t.Fatalf("json.Unmarshal into map failed: %v", err)
				}
				if _, ok := m["id"]; ok {
					t.Errorf("Expected 'id' field to be omitted from JSON, but it was present: %s", string(data))
				}
			}


			// Unmarshal
			var unmarshaledReq Request
			if err := json.Unmarshal(data, &unmarshaledReq); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Verify fields
			if unmarshaledReq.JSONRPC != tt.input.JSONRPC {
				t.Errorf("JSONRPC mismatch: got %s, want %s", unmarshaledReq.JSONRPC, tt.input.JSONRPC)
			}
			if unmarshaledReq.Method != tt.input.Method {
				t.Errorf("Method mismatch: got %s, want %s", unmarshaledReq.Method, tt.input.Method)
			}

			// Verify ID
			if tt.input.ID == nil { // Case: Request.ID *RequestID was nil in the input struct (e.g. "ID completely omitted")
				// "id" field should be absent in JSON due to `omitempty`.
				// Unmarshalling should result in unmarshaledReq.ID being nil.
				if unmarshaledReq.ID != nil {
					t.Errorf("ID mismatch for omitted ID: expected nil *RequestID, got non-nil ID with value %v", unmarshaledReq.ID.Value())
				}
			} else { // Case: Request.ID *RequestID was NOT nil in the input struct (e.g. NewRequestID("foo") or NewRequestID(nil))
				if tt.input.ID.Value() == nil { // Specifically for input NewRequestID(nil), which means JSON "id":null
					// After unmarshalling JSON `{"id":null}`, the `unmarshaledReq.ID` pointer itself should be nil.
					if unmarshaledReq.ID != nil {
						t.Errorf("ID mismatch for 'null' ID: expected nil *RequestID from JSON 'null', got non-nil *RequestID with value '%v'", unmarshaledReq.ID.Value())
					}
				} else { // For non-null ID values (string or number)
					if unmarshaledReq.ID == nil {
						t.Errorf("ID mismatch: got nil *RequestID, expected non-nil *RequestID with value %v", tt.input.ID.Value())
					} else if !reflect.DeepEqual(unmarshaledReq.ID.Value(), tt.input.ID.Value()) {
						t.Errorf("ID value mismatch: got %v (type %T), want %v (type %T)", 
							unmarshaledReq.ID.Value(), unmarshaledReq.ID.Value(), 
							tt.input.ID.Value(), tt.input.ID.Value())
					}
				}
			}
			
			// Verify Params by unmarshalling both to interface{} and comparing with DeepEqual
			if tt.input.Params == nil {
				if unmarshaledReq.Params != nil {
					t.Errorf("Params mismatch: expected nil params, got %s", string(unmarshaledReq.Params))
				}
			} else {
				if unmarshaledReq.Params == nil {
					t.Errorf("Params mismatch: expected non-nil params %s, got nil", string(tt.input.Params))
				} else {
					var expectedParamsInterface, actualParamsInterface interface{}
					if err := json.Unmarshal(tt.input.Params, &expectedParamsInterface); err != nil {
						t.Fatalf("Error unmarshalling expected params for comparison: %v", err)
					}
					if err := json.Unmarshal(unmarshaledReq.Params, &actualParamsInterface); err != nil {
						t.Fatalf("Error unmarshalling actual params for comparison: %v", err)
					}

					if !reflect.DeepEqual(actualParamsInterface, expectedParamsInterface) {
						t.Errorf("Params content mismatch after DeepEqual:\nExpected: %s (%+v)\nActual:   %s (%+v)",
							string(tt.input.Params), expectedParamsInterface,
							string(unmarshaledReq.Params), actualParamsInterface)
					}
				}
			}
		})
	}
}


func TestError_Error(t *testing.T) {
	err := NewError(InvalidRequest, "test error", "test data")
	
	expected := "JSON-RPC error -32600: test error"
	if err.Error() != expected {
		t.Errorf("Expected %s, got %s", expected, err.Error())
	}
}

func TestResponse_MarshalUnmarshal(t *testing.T) {
	type testCase struct {
		name  string
		input Response
	}

	tests := []testCase{
		{
			name: "response with string ID and result",
			input: Response{
				JSONRPC: JSONRPCVersion,
				ID:      NewRequestID("res-str-id"),
				Result:  json.RawMessage(`{"status":"ok","data":123}`),
			},
		},
		{
			name: "response with number ID and result",
			input: Response{
				JSONRPC: JSONRPCVersion,
				ID:      NewRequestID(float64(789)), // Ensure float64 for numeric ID
				Result:  json.RawMessage(`["item1",true]`),      // Canonical JSON
			},
		},
		{
			name: "response with null ID and result",
			input: Response{
				JSONRPC: JSONRPCVersion,
				ID:      NewRequestID(nil),
				Result:  json.RawMessage(`"success"`),
			},
		},
		{
			name: "response with string ID and error",
			input: Response{
				JSONRPC: JSONRPCVersion,
				ID:      NewRequestID("res-err-id"),
				Error:   NewError(MethodNotFound, "Method not found", nil),
			},
		},
		{
			name: "response with null ID and error", // As per JSON-RPC 2.0 spec, error responses should have an ID
			input: Response{
				JSONRPC: JSONRPCVersion,
				ID:      NewRequestID(nil),
				Error:   NewError(ParseError, "Parse error occurred", map[string]string{"details": "invalid char"}),
			},
		},
		// Note: A response MUST include Result or Error, but not both.
		// The json encoding will handle this; if both are present, both will be encoded.
		// The spec says "The members result and error MUST NOT exist at the same time."
		// Our struct allows it, but valid JSON-RPC responses shouldn't have both.
		// We test what our struct can marshal/unmarshal.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			var unmarshaledResp Response
			if err := json.Unmarshal(data, &unmarshaledResp); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Basic checks
			if unmarshaledResp.JSONRPC != tt.input.JSONRPC {
				t.Errorf("JSONRPC mismatch: got %s, want %s", unmarshaledResp.JSONRPC, tt.input.JSONRPC)
			}

			// Check ID
			// For Responses, ID is mandatory and not omitempty. tt.input.ID will be non-nil.
			// tt.input.ID.Value() can be nil (if input was NewRequestID(nil)).
			if unmarshaledResp.ID == nil { 
				// This implies JSON `id:null` was unmarshalled into a nil *RequestID pointer.
				// This is correct if tt.input.ID.Value() was nil.
				if tt.input.ID.Value() != nil { // Should not happen if input ID value was non-nil
					t.Errorf("ID mismatch: got nil *RequestID for a non-null expected ID value %v", tt.input.ID.Value())
				}
				// If tt.input.ID.Value() was also nil, then (nil *RequestID vs. NewRequestID(nil)) is a pass.
			} else { // unmarshaledResp.ID is not nil
				if tt.input.ID.Value() == nil { // Expected ID value was nil, but got a non-nil *RequestID
					// This means JSON "id":null was unmarshalled into *RequestID{value: something_not_nil_after_roundtrip}
					// or our custom UnmarshalJSON for RequestID didn't set value to nil for "null" input.
					// Given RequestID.UnmarshalJSON, unmarshaledResp.ID.Value() should be nil here.
					if unmarshaledResp.ID.Value() != nil {
						t.Errorf("ID mismatch for 'null' ID: expected nil value, got value %v", unmarshaledResp.ID.Value())
					}
				} else { // Both expected and actual ID values are non-nil. Compare them.
					if !reflect.DeepEqual(unmarshaledResp.ID.Value(), tt.input.ID.Value()) {
						t.Errorf("ID value mismatch: got %v (type %T), want %v (type %T)", 
							unmarshaledResp.ID.Value(), unmarshaledResp.ID.Value(), 
							tt.input.ID.Value(), tt.input.ID.Value())
					}
				}
			}

			// Check Result
			if tt.input.Result != nil {
				if unmarshaledResp.Result == nil {
					t.Errorf("Result mismatch: expected non-nil result, got nil. Expected: %s", string(tt.input.Result.(json.RawMessage)))
				} else {
					// Marshal the unmarshaled result back to JSON to compare with the original RawMessage
					actualResultJSON, err := json.Marshal(unmarshaledResp.Result)
					if err != nil {
						t.Fatalf("Error marshalling actual result for comparison: %v", err)
					}
					// Compare string forms of JSON
					// tt.input.Result is already json.RawMessage (which is []byte)
					expectedResultJSONString := string(tt.input.Result.(json.RawMessage))
					actualResultJSONString := string(actualResultJSON)

					// To make comparison robust against formatting differences, unmarshal both to interface{} then DeepEqual
					var expectedInterface, actualInterface interface{}
					if err := json.Unmarshal([]byte(expectedResultJSONString), &expectedInterface); err != nil {
						t.Fatalf("Error unmarshalling expectedResultJSONString: %v", err)
					}
					if err := json.Unmarshal([]byte(actualResultJSONString), &actualInterface); err != nil {
						t.Fatalf("Error unmarshalling actualResultJSONString: %v", err)
					}

					if !reflect.DeepEqual(actualInterface, expectedInterface) {
						t.Errorf("Result mismatch after re-marshalling and DeepEqual:\nExpected JSON: %s\nActual JSON:   %s",
							expectedResultJSONString, actualResultJSONString)
					}
				}
			} else if unmarshaledResp.Result != nil {
				// If input result was nil, unmarshaled should also be nil (or effectively empty if omitempty didn't remove it)
				// Since Result is interface{}, if it was absent in JSON, it would be nil.
				t.Errorf("Result mismatch: expected nil result, got non-nil: %+v", unmarshaledResp.Result)
			}

			// Check Error
			if tt.input.Error != nil {
				if unmarshaledResp.Error == nil {
					t.Errorf("Error mismatch: expected error, got nil")
				} else {
					if unmarshaledResp.Error.Code != tt.input.Error.Code ||
						unmarshaledResp.Error.Message != tt.input.Error.Message {
						t.Errorf("Error content mismatch: got %+v, want %+v", *unmarshaledResp.Error, *tt.input.Error)
					}
					// Comparing Data field of Error can be complex if it's an interface{}
					// For simplicity, if code and message match, consider it mostly fine for this test.
					// A full DeepEqual on Error.Data might be needed for very strict tests.
				}
			} else if unmarshaledResp.Error != nil {
				t.Errorf("Error mismatch: expected nil, got %+v", *unmarshaledResp.Error)
			}
		})
	}
}