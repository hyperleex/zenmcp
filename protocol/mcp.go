package protocol

import (
	"encoding/json"
)

type ServerCapabilities struct {
	Tools        *ToolsCapability        `json:"tools,omitempty"`
	Resources    *ResourcesCapability    `json:"resources,omitempty"`
	Prompts      *PromptsCapability      `json:"prompts,omitempty"`
	Logging      *LoggingCapability      `json:"logging,omitempty"`
	Completion   *CompletionCapability   `json:"completion,omitempty"`
	Experimental map[string]interface{}  `json:"experimental,omitempty"`
}

type ClientCapabilities struct {
	Roots        *RootsCapability        `json:"roots,omitempty"`
	Sampling     *SamplingCapability     `json:"sampling,omitempty"`
	Experimental map[string]interface{}  `json:"experimental,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type LoggingCapability struct{}

type CompletionCapability struct {
	CompletionClient *CompletionClientCapability `json:"completionClient,omitempty"`
}

type CompletionClientCapability struct {
	Arguments *CompletionArgumentsCapability `json:"arguments,omitempty"`
}

type CompletionArgumentsCapability struct {
	Name string `json:"name"`
}

type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type SamplingCapability struct{}

type InitializeRequest struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      ClientInfo         `json:"clientInfo"`
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ToolDescriptor struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type ToolListRequest struct{}

type ToolListResult struct {
	Tools []ToolDescriptor `json:"tools"`
}

type ToolCallRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type ToolCallResult struct {
	Content []Content `json:"content"`
	IsError bool      `json:"isError,omitempty"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type PromptMessage struct {
	Role    string                 `json:"role"`
	Content map[string]interface{} `json:"content"`
}

type ProgressToken struct {
	value interface{}
}

func (p *ProgressToken) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		p.value = str
		return nil
	}
	
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		p.value = num
		return nil
	}
	
	return json.Unmarshal(data, &p.value)
}

func (p *ProgressToken) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.value)
}

func (p *ProgressToken) Value() interface{} {
	return p.value
}

type ProgressNotification struct {
	ProgressToken ProgressToken `json:"progressToken"`
	Progress      float64       `json:"progress"`
	Total         *float64      `json:"total,omitempty"`
}

const (
	MethodInitialize    = "initialize"
	MethodInitialized   = "notifications/initialized"
	MethodShutdown      = "shutdown"
	MethodExit          = "exit"
	MethodToolsList     = "tools/list"
	MethodToolsCall     = "tools/call"
	MethodProgress      = "notifications/progress"
	MethodCancellation  = "notifications/cancelled"
	MethodToolsListChanged = "notifications/tools/list_changed"
)