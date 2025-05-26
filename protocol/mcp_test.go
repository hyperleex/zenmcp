package protocol

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestProgressToken_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name         string
		input        interface{} // Input to NewProgressToken constructor (not used directly)
		jsonInput    string      // The JSON string to unmarshal from
		expectedValue interface{} // The expected p.value after unmarshal
	}{
		{
			name:         "string token",
			input:        "request-1",
			jsonInput:    `"request-1"`,
			expectedValue: "request-1",
		},
		{
			name:         "number token",
			input:        123,
			jsonInput:    `123`,
			expectedValue: float64(123), // Numbers unmarshal to float64
		},
		{
			name:      "null token",
			input:     nil,
			jsonInput: `null`,
			expectedValue: nil,
		},
		{
			name:      "boolean token true",
			input:     true,
			jsonInput: `true`,
			expectedValue: true,
		},
		{
			name:      "boolean token false",
			input:     false,
			jsonInput: `false`,
			expectedValue: false,
		},
		{
			name:      "array token",
			input:     []interface{}{"a", float64(1)},
			jsonInput: `["a", 1]`,
			expectedValue: []interface{}{"a", float64(1)},
		},
		{
			name:      "object token",
			input:     map[string]interface{}{"key": "value", "num": float64(2)},
			jsonInput: `{"key":"value", "num":2}`,
			expectedValue: map[string]interface{}{"key": "value", "num": float64(2)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Unmarshal
			var token ProgressToken
			err := json.Unmarshal([]byte(tt.jsonInput), &token)
			if err != nil {
				// The current UnmarshalJSON for ProgressToken has a fallback that attempts
				// json.Unmarshal(data, &p.value), which might succeed for types not explicitly string/number.
				// If the plan is to make it stricter (error on bool/array/object/null), this test would change.
				// For now, testing existing flexible behavior.
				t.Fatalf("UnmarshalJSON error: %v for input %s", err, tt.jsonInput)
			}

			if !reflect.DeepEqual(token.Value(), tt.expectedValue) {
				t.Errorf("After Unmarshal: Expected value %v (type %T), got %v (type %T)",
					tt.expectedValue, tt.expectedValue, token.Value(), token.Value())
			}

			// Test Marshal (using the value set by Unmarshal to ensure round trip)
			// Or, we can construct ProgressToken directly if NewProgressToken is available.
			// Assuming ProgressToken struct is directly settable for `value` or has a constructor.
			// Since there's no NewProgressToken, we'll use the unmarshaled token.
			
			marshaledData, err := json.Marshal(token)
			if err != nil {
				t.Fatalf("MarshalJSON error: %v", err)
			}

			// Unmarshal again to check if marshaled data is what we expect
			// This is more of a check that MarshalJSON produces something UnmarshalJSON can read back
			// to the same internal representation.
			var token2 ProgressToken
			err = json.Unmarshal(marshaledData, &token2)
			if err != nil {
				t.Fatalf("UnmarshalJSON (second pass) error: %v for marshaled data %s", err, string(marshaledData))
			}

			if !reflect.DeepEqual(token2.Value(), tt.expectedValue) {
				t.Errorf("After second Unmarshal: Expected value %v (type %T), got %v (type %T)",
					tt.expectedValue, tt.expectedValue, token2.Value(), token2.Value())
			}
			
			// Also, it might be useful to compare string(marshaledData) with tt.jsonInput,
			// but this can be tricky if tt.jsonInput has different spacing or key order for objects.
			// A canonical check is better: unmarshal the marshaled data and compare the Go struct.
		})
	}
}

func TestCapabilities_MarshalUnmarshal(t *testing.T) {
	t.Run("ServerCapabilities_Populated", func(t *testing.T) {
		original := ServerCapabilities{
			Tools:        &ToolsCapability{ListChanged: true},
			Resources:    &ResourcesCapability{Subscribe: true, ListChanged: false},
			Prompts:      &PromptsCapability{ListChanged: true},
			Logging:      &LoggingCapability{},
			Completion:   &CompletionCapability{CompletionClient: &CompletionClientCapability{Arguments: &CompletionArgumentsCapability{Name: "argName"}}},
			Experimental: map[string]interface{}{"exp_key": "exp_value", "exp_num": 123.0},
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var deserialized ServerCapabilities
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("ServerCapabilities (Populated) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})

	t.Run("ServerCapabilities_Omitted", func(t *testing.T) {
		original := ServerCapabilities{
			// All fields omitted or explicitly nil where pointers
			Tools:        nil, // Test omitempty
			Resources:    &ResourcesCapability{}, // Test with empty struct but non-nil pointer
			Prompts:      nil,
			Logging:      nil, // Test omitempty for non-pointer struct (will still be present if not pointer)
			Completion:   &CompletionCapability{CompletionClient: &CompletionClientCapability{Arguments: &CompletionArgumentsCapability{Name: ""}}}, // Test empty string
			Experimental: nil,
		}
		// LoggingCapability is not a pointer in ServerCapabilities, so it won't be omitted by json if ServerCapabilities is non-nil.
		// It will be {"logging":{}} if Logging: &LoggingCapability{} or Logging: LoggingCapability{}.
		// If Logging is nil pointer type, it would be omitted. Current struct has Logging *LoggingCapability.

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		
		// Check if "tools" is omitted
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		if _, ok := m["tools"]; ok {
			t.Errorf("Expected 'tools' to be omitted, but found in JSON: %s", string(data))
		}


		var deserialized ServerCapabilities
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		
		// Need to be careful with DeepEqual if original.Logging was nil and deserialized.Logging is &LoggingCapability{}
		// However, if original.Logging is nil, it should be omitted from JSON, and then deserialized.Logging should also be nil.
		// If original.Logging was &LoggingCapability{}, then JSON is {"logging":{}}, and deserialized.Logging is &LoggingCapability{}.

		if !reflect.DeepEqual(original, deserialized) {
			// For logging, if original.Logging is nil, and deserialized.Logging is non-nil but points to an empty struct,
			// this might be an acceptable outcome of omitempty behavior.
			// Let's compare field by field if DeepEqual fails for complex cases.
			if original.Tools != nil && deserialized.Tools == nil || (original.Tools == nil && deserialized.Tools != nil) {
				 // This check is too simple, DeepEqual should handle it.
			}
			t.Errorf("ServerCapabilities (Omitted) mismatch:\nOriginal:   %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})

	t.Run("ClientCapabilities_Populated", func(t *testing.T) {
		original := ClientCapabilities{
			Roots:        &RootsCapability{ListChanged: true},
			Sampling:     &SamplingCapability{}, // Empty struct
			Experimental: map[string]interface{}{"client_exp": true},
		}

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var deserialized ClientCapabilities
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("ClientCapabilities (Populated) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})

	t.Run("ClientCapabilities_Omitted", func(t *testing.T) {
		original := ClientCapabilities{
			Roots:    nil,
			Sampling: nil, // SamplingCapability is a struct, not pointer. So it will be present as {}.
			// To make Sampling truly omittable, it should be *SamplingCapability.
			// Given current struct def, Sampling: nil is not possible.
			// We test with it being its zero value.
			// Experimental: nil, // This will be omitted by omitempty
		}
		// Expected: {"sampling":{}} or just {} if all are omittable and nil/zero.
		// Since Sampling is not a pointer, it won't be omitted by `omitempty` unless ClientCapabilities itself is an interface or pointer.
		// It will be present as `{"sampling":{}}`.

		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}

		var deserialized ClientCapabilities
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}

		if !reflect.DeepEqual(original, deserialized) {
			// If original.Sampling was SamplingCapability{} (zero value), it will be {"sampling":{}}
			// deserialized.Sampling will also be SamplingCapability{}. DeepEqual should work.
			t.Errorf("ClientCapabilities (Omitted) mismatch:\nOriginal:   %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})
}

func TestInitializeRequest_MarshalUnmarshal(t *testing.T) {
	t.Run("Populated", func(t *testing.T) {
		original := InitializeRequest{
			ProtocolVersion: "1.0",
			Capabilities: ClientCapabilities{
				Roots:        &RootsCapability{ListChanged: true},
				Sampling:     &SamplingCapability{},
				Experimental: map[string]interface{}{"client_feature": true},
			},
			ClientInfo: ClientInfo{
				Name:    "TestClient",
				Version: "0.1.0",
			},
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var deserialized InitializeRequest
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("InitializeRequest (Populated) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})

	t.Run("OmittedOptionalFields", func(t *testing.T) {
		original := InitializeRequest{ // Capabilities and ClientInfo are not omitempty
			ProtocolVersion: "1.0",
			Capabilities:    ClientCapabilities{}, // Test with empty, non-nil struct
			ClientInfo:      ClientInfo{Name: "MinimalClient", Version: "0.0.1"},
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var deserialized InitializeRequest
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("InitializeRequest (OmittedOptionalFields) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})
}

func TestInitializeResult_MarshalUnmarshal(t *testing.T) {
	t.Run("Populated", func(t *testing.T) {
		original := InitializeResult{
			ProtocolVersion: "1.0",
			Capabilities: ServerCapabilities{
				Tools:     &ToolsCapability{ListChanged: true},
				Resources: &ResourcesCapability{Subscribe: true},
			},
			ServerInfo: ServerInfo{
				Name:    "TestServer",
				Version: "0.2.0",
			},
			Instructions: "Please proceed.",
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var deserialized InitializeResult
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("InitializeResult (Populated) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})

	t.Run("OmittedOptionalFields", func(t *testing.T) {
		original := InitializeResult{ // Capabilities and ServerInfo are not omitempty
			ProtocolVersion: "1.0",
			Capabilities:    ServerCapabilities{}, // Test with empty, non-nil struct
			ServerInfo:      ServerInfo{Name: "MinimalServer", Version: "0.0.2"},
			// Instructions is omitempty
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var deserialized InitializeResult
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("InitializeResult (OmittedOptionalFields) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})
}

func TestClientServerInfo_MarshalUnmarshal(t *testing.T) {
	// ClientInfo and ServerInfo are identical in structure
	t.Run("ClientInfo", func(t *testing.T) {
		original := ClientInfo{Name: "TestClient", Version: "1.2.3"}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("ClientInfo Marshal error: %v", err)
		}
		var deserialized ClientInfo
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("ClientInfo Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("ClientInfo mismatch:\nOriginal: %+v\nDeserialized: %+v", original, deserialized)
		}
	})
	t.Run("ServerInfo", func(t *testing.T) {
		original := ServerInfo{Name: "TestServer", Version: "4.5.6"}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("ServerInfo Marshal error: %v", err)
		}
		var deserialized ServerInfo
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("ServerInfo Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("ServerInfo mismatch:\nOriginal: %+v\nDeserialized: %+v", original, deserialized)
		}
	})
}

func TestToolDescriptor_MarshalUnmarshal(t *testing.T) {
	t.Run("Populated", func(t *testing.T) {
		original := ToolDescriptor{
			Name:        "testTool",
			Description: "A tool for testing.",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"param1": map[string]interface{}{"type": "string"}}},
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var deserialized ToolDescriptor
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("ToolDescriptor (Populated) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})
	t.Run("OmittedOptionalFields", func(t *testing.T) {
		original := ToolDescriptor{ // Description is omitempty
			Name:        "minimalTool",
			InputSchema: map[string]interface{}{"type": "string"},
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var deserialized ToolDescriptor
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("ToolDescriptor (OmittedOptionalFields) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})
}

func TestToolListRequest_MarshalUnmarshal(t *testing.T) {
	original := ToolListRequest{} // Empty struct
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	if string(data) != "{}" { // Empty struct marshals to {}
		t.Errorf("Expected ToolListRequest to marshal to '{}', got %s", string(data))
	}
	var deserialized ToolListRequest
	if err := json.Unmarshal(data, &deserialized); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if !reflect.DeepEqual(original, deserialized) {
		t.Errorf("ToolListRequest mismatch:\nOriginal: %+v\nDeserialized: %+v", original, deserialized)
	}
}

func TestToolListResult_MarshalUnmarshal(t *testing.T) {
	original := ToolListResult{
		Tools: []ToolDescriptor{
			{Name: "toolA", InputSchema: map[string]interface{}{"type": "string"}},
			{Name: "toolB", Description: "Tool B desc", InputSchema: map[string]interface{}{"type": "number"}},
		},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var deserialized ToolListResult
	if err := json.Unmarshal(data, &deserialized); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if !reflect.DeepEqual(original, deserialized) {
		t.Errorf("ToolListResult mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
	}

	// Test empty tools list
	originalEmpty := ToolListResult{Tools: []ToolDescriptor{}} // Marshals to {"tools":[]}
	dataEmpty, errEmpty := json.Marshal(originalEmpty)
	if errEmpty != nil {
		t.Fatalf("Marshal error (empty): %v", errEmpty)
	}
	if string(dataEmpty) != `{"tools":[]}` {
		t.Errorf("Expected empty ToolListResult to marshal to '{\"tools\":[]}', got %s", string(dataEmpty))
	}
	var deserializedEmpty ToolListResult
	if err := json.Unmarshal(dataEmpty, &deserializedEmpty); err != nil {
		t.Fatalf("Unmarshal error (empty): %v", err)
	}
	// DeepEqual should handle empty non-nil slices.
	// If Tools was nil, then after unmarshal it might be non-nil empty, or nil depending on JSON.
	// `{"tools":null}` -> nil slice. `{"tools":[]}` -> empty non-nil slice.
	// Our struct has `Tools []ToolDescriptor`, not a pointer, so it won't be nil from JSON `null`.
	if !reflect.DeepEqual(originalEmpty, deserializedEmpty) {
		t.Errorf("ToolListResult (empty) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", originalEmpty, deserializedEmpty, string(dataEmpty))
	}
	
	// Test nil tools list (should be same as empty for marshal if not omitempty, but good to check)
	originalNil := ToolListResult{Tools: nil} // Marshals to {"tools":null}
	dataNil, errNil := json.Marshal(originalNil)
	if errNil != nil {
		t.Fatalf("Marshal error (nil): %v", errNil)
	}
	if string(dataNil) != `{"tools":null}` {
		t.Errorf("Expected nil ToolListResult.Tools to marshal to '{\"tools\":null}', got %s", string(dataNil))
	}
	var deserializedNil ToolListResult
	if err := json.Unmarshal(dataNil, &deserializedNil); err != nil {
		t.Fatalf("Unmarshal error (nil): %v", err)
	}
	if deserializedNil.Tools != nil { // After unmarshalling "tools":null, Tools should be nil
		t.Errorf("ToolListResult (nil) mismatch: Expected Tools to be nil, got %+v", deserializedNil.Tools)
	}

}

func TestToolCallRequest_MarshalUnmarshal(t *testing.T) {
	t.Run("WithParams", func(t *testing.T) {
		original := ToolCallRequest{
			Name:      "doSomething",
			Arguments: json.RawMessage(`{"arg1":"val1","arg2":100}`),
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var deserialized ToolCallRequest
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("ToolCallRequest (WithParams) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})
	t.Run("WithoutParams", func(t *testing.T) { // Arguments is omitempty
		original := ToolCallRequest{Name: "doSomethingSimple"}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var deserialized ToolCallRequest
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("ToolCallRequest (WithoutParams) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})
}

func TestToolCallResult_MarshalUnmarshal(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		original := ToolCallResult{
			Content: []Content{
				{Type: "text", Text: "Tool executed successfully."},
				{Type: "status", Text: "complete"},
			},
			IsError: false, // omitempty should remove it
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var deserialized ToolCallResult
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("ToolCallResult (Success) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})
	t.Run("Error", func(t *testing.T) {
		original := ToolCallResult{
			Content: []Content{{Type: "text", Text: "Tool failed."}},
			IsError: true,
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var deserialized ToolCallResult
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("ToolCallResult (Error) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})
	t.Run("EmptyContent", func(t *testing.T) {
		original := ToolCallResult{Content: []Content{}} // Should be "content":[]
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var deserialized ToolCallResult
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("ToolCallResult (EmptyContent) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})
}

func TestContent_MarshalUnmarshal(t *testing.T) {
	original := Content{Type: "text/markdown", Text: "# Hello"}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var deserialized Content
	if err := json.Unmarshal(data, &deserialized); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if !reflect.DeepEqual(original, deserialized) {
		t.Errorf("Content mismatch:\nOriginal: %+v\nDeserialized: %+v", original, deserialized)
	}

	// Test with omitted Text (is omitempty)
	originalOmit := Content{Type: "status"}
	dataOmit, errOmit := json.Marshal(originalOmit)
	if errOmit != nil {
		t.Fatalf("Marshal error (omit text): %v", errOmit)
	}
	var deserializedOmit Content
	if err := json.Unmarshal(dataOmit, &deserializedOmit); err != nil {
		t.Fatalf("Unmarshal error (omit text): %v", errOmit)
	}
	if !reflect.DeepEqual(originalOmit, deserializedOmit) {
		t.Errorf("Content (omit text) mismatch:\nOriginal: %+v\nDeserialized: %+v", originalOmit, deserializedOmit)
	}
}

func TestPromptMessage_MarshalUnmarshal(t *testing.T) {
	original := PromptMessage{
		Role:    "user",
		Content: map[string]interface{}{"type": "text", "text": "Hello, assistant!"},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}
	var deserialized PromptMessage
	if err := json.Unmarshal(data, &deserialized); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if !reflect.DeepEqual(original, deserialized) {
		t.Errorf("PromptMessage mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
	}
}

func TestProgressNotification_MarshalUnmarshal(t *testing.T) {
	t.Run("WithTotal", func(t *testing.T) {
		total := float64(100)
		original := ProgressNotification{
			ProgressToken: ProgressToken{value: "token-abc"},
			Progress:      50.5,
			Total:         &total,
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var deserialized ProgressNotification
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("ProgressNotification (WithTotal) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})
	t.Run("WithoutTotal", func(t *testing.T) { // Total is omitempty
		original := ProgressNotification{
			ProgressToken: ProgressToken{value: 12345},
			Progress:      0.75,
		}
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Marshal error: %v", err)
		}
		var deserialized ProgressNotification
		if err := json.Unmarshal(data, &deserialized); err != nil {
			t.Fatalf("Unmarshal error: %v", err)
		}
		if !reflect.DeepEqual(original, deserialized) {
			t.Errorf("ProgressNotification (WithoutTotal) mismatch:\nOriginal: %+v\nDeserialized: %+v\nJSON: %s", original, deserialized, string(data))
		}
	})
}
