package registry

import (
	"encoding/json"
	"testing"

	"github.com/hyperleex/zenmcp/protocol"
)

type testHandler struct{}

func (h *testHandler) Call(ctx interface{}, args json.RawMessage) (*protocol.ToolCallResult, error) {
	return &protocol.ToolCallResult{
		Content: []protocol.Content{{Type: "text", Text: "test result"}},
	}, nil
}

type testArgs struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Optional string `json:"optional,omitempty"`
}

func TestRegistry_RegisterTool(t *testing.T) {
	registry := New()
	handler := &testHandler{}
	
	err := registry.RegisterTool("test_tool", "A test tool", handler, testArgs{})
	if err != nil {
		t.Fatalf("RegisterTool error: %v", err)
	}
	
	tool, exists := registry.GetTool("test_tool")
	if !exists {
		t.Fatal("Expected tool to exist")
	}
	
	if tool.Name != "test_tool" {
		t.Errorf("Expected name test_tool, got %s", tool.Name)
	}
	
	if tool.Description != "A test tool" {
		t.Errorf("Expected description 'A test tool', got %s", tool.Description)
	}
	
	if tool.Handler != handler {
		t.Error("Expected same handler")
	}
}

func TestRegistry_ListTools(t *testing.T) {
	registry := New()
	handler := &testHandler{}
	
	registry.RegisterTool("tool1", "Tool 1", handler, nil)
	registry.RegisterTool("tool2", "Tool 2", handler, nil)
	
	tools := registry.ListTools()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}
	
	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
	}
	
	if !names["tool1"] || !names["tool2"] {
		t.Error("Expected tools tool1 and tool2")
	}
}

func TestGenerateJSONSchema(t *testing.T) {
	schema, err := generateJSONSchema(testArgs{})
	if err != nil {
		t.Fatalf("generateJSONSchema error: %v", err)
	}
	
	if schema["type"] != "object" {
		t.Errorf("Expected type object, got %v", schema["type"])
	}
	
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be map[string]interface{}")
	}
	
	nameField, ok := properties["name"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected name field")
	}
	
	if nameField["type"] != "string" {
		t.Errorf("Expected name type string, got %v", nameField["type"])
	}
	
	ageField, ok := properties["age"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected age field")
	}
	
	if ageField["type"] != "integer" {
		t.Errorf("Expected age type integer, got %v", ageField["type"])
	}
	
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Expected required to be []string")
	}
	
	requiredMap := make(map[string]bool)
	for _, field := range required {
		requiredMap[field] = true
	}
	
	if !requiredMap["name"] || !requiredMap["age"] {
		t.Error("Expected name and age to be required")
	}
	
	if requiredMap["optional"] {
		t.Error("Expected optional to not be required")
	}
}

func TestGenerateJSONSchema_Nil(t *testing.T) {
	schema, err := generateJSONSchema(nil)
	if err != nil {
		t.Fatalf("generateJSONSchema error: %v", err)
	}
	
	if schema["type"] != "object" {
		t.Errorf("Expected type object, got %v", schema["type"])
	}
}