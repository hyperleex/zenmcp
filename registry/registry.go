package registry

import (
	"encoding/json"
	"reflect"

	"github.com/hyperleex/zenmcp/protocol"
)

type Registry struct {
	tools     map[string]*ToolDescriptor
	resources map[string]*ResourceDescriptor
	prompts   map[string]*PromptDescriptor
}

type ToolDescriptor struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Handler     LegacyToolHandler      `json:"-"`
}

type ResourceDescriptor struct {
	URI         string          `json:"uri"`
	Name        string          `json:"name,omitempty"`
	Description string          `json:"description,omitempty"`
	MimeType    string          `json:"mimeType,omitempty"`
	Handler     ResourceHandler `json:"-"`
}

type PromptDescriptor struct {
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	Arguments   []Argument            `json:"arguments,omitempty"`
	Handler     LegacyPromptHandler   `json:"-"`
}

type Argument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// LegacyToolHandler maintains compatibility with existing code
type LegacyToolHandler interface {
	Call(ctx interface{}, args json.RawMessage) (*protocol.ToolCallResult, error)
}

// ResourceHandler is the legacy interface for resource handlers
type ResourceHandler interface {
	Read(ctx interface{}, uri string) ([]byte, string, error)
}

// LegacyPromptHandler maintains compatibility
type LegacyPromptHandler interface {
	Get(ctx interface{}, args map[string]interface{}) (*PromptResult, error)
}

type PromptResult struct {
	Description string                   `json:"description,omitempty"`
	Messages    []protocol.PromptMessage `json:"messages"`
}

func New() *Registry {
	return &Registry{
		tools:     make(map[string]*ToolDescriptor),
		resources: make(map[string]*ResourceDescriptor),
		prompts:   make(map[string]*PromptDescriptor),
	}
}

func (r *Registry) RegisterTool(name, description string, handler LegacyToolHandler, inputType interface{}) error {
	schema, err := generateJSONSchema(inputType)
	if err != nil {
		return err
	}
	
	r.tools[name] = &ToolDescriptor{
		Name:        name,
		Description: description,
		InputSchema: schema,
		Handler:     handler,
	}
	return nil
}

func (r *Registry) RegisterResource(uri, name, description, mimeType string, handler ResourceHandler) {
	r.resources[uri] = &ResourceDescriptor{
		URI:         uri,
		Name:        name,
		Description: description,
		MimeType:    mimeType,
		Handler:     handler,
	}
}

func (r *Registry) RegisterPrompt(name, description string, args []Argument, handler LegacyPromptHandler) {
	r.prompts[name] = &PromptDescriptor{
		Name:        name,
		Description: description,
		Arguments:   args,
		Handler:     handler,
	}
}


func (r *Registry) GetTool(name string) (*ToolDescriptor, bool) {
	tool, exists := r.tools[name]
	return tool, exists
}

func (r *Registry) GetResource(uri string) (*ResourceDescriptor, bool) {
	resource, exists := r.resources[uri]
	return resource, exists
}

func (r *Registry) GetPrompt(name string) (*PromptDescriptor, bool) {
	prompt, exists := r.prompts[name]
	return prompt, exists
}

func (r *Registry) ListTools() []protocol.ToolDescriptor {
	var tools []protocol.ToolDescriptor
	for _, tool := range r.tools {
		tools = append(tools, protocol.ToolDescriptor{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		})
	}
	return tools
}

func generateJSONSchema(v interface{}) (map[string]interface{}, error) {
	if v == nil {
		return map[string]interface{}{
			"type": "object",
		}, nil
	}
	
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}
	
	if t.Kind() != reflect.Struct {
		return schema, nil
	}
	
	properties := schema["properties"].(map[string]interface{})
	var required []string
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}
		
		fieldName := field.Name
		if jsonTag != "" && jsonTag != "-" {
			if idx := len(jsonTag); idx > 0 {
				if commaIdx := 0; commaIdx < len(jsonTag) {
					for j, c := range jsonTag {
						if c == ',' {
							commaIdx = j
							break
						}
					}
					if commaIdx > 0 {
						fieldName = jsonTag[:commaIdx]
					} else {
						fieldName = jsonTag
					}
				}
			}
		}
		
		fieldSchema := getFieldSchema(field.Type)
		properties[fieldName] = fieldSchema
		
		if !hasOmitemptyTag(field.Tag.Get("json")) {
			required = append(required, fieldName)
		}
	}
	
	if len(required) > 0 {
		schema["required"] = required
	}
	
	return schema, nil
}

func getFieldSchema(t reflect.Type) map[string]interface{} {
	switch t.Kind() {
	case reflect.String:
		return map[string]interface{}{"type": "string"}
	case reflect.Int, reflect.Int32, reflect.Int64:
		return map[string]interface{}{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{"type": "number"}
	case reflect.Bool:
		return map[string]interface{}{"type": "boolean"}
	case reflect.Slice, reflect.Array:
		return map[string]interface{}{
			"type":  "array",
			"items": getFieldSchema(t.Elem()),
		}
	case reflect.Ptr:
		return getFieldSchema(t.Elem())
	default:
		return map[string]interface{}{"type": "object"}
	}
}

func hasOmitemptyTag(tag string) bool {
	if tag == "" {
		return false
	}
	for i, c := range tag {
		if c == ',' && i+1 < len(tag) {
			rest := tag[i+1:]
			return rest == "omitempty" || (len(rest) > 9 && rest[:9] == "omitempty")
		}
	}
	return false
}