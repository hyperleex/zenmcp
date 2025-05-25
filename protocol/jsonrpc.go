package protocol

import (
	"encoding/json"
	"fmt"
)

const JSONRPCVersion = "2.0"

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *RequestID      `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      *RequestID  `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type RequestID struct {
	value interface{}
}

func (r *RequestID) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		r.value = str
		return nil
	}
	
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		r.value = num
		return nil
	}
	
	if string(data) == "null" {
		r.value = nil
		return nil
	}
	
	return fmt.Errorf("invalid request ID type")
}

func (r *RequestID) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.value)
}

func (r *RequestID) Value() interface{} {
	return r.value
}

func NewRequestID(v interface{}) *RequestID {
	return &RequestID{value: v}
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

func NewError(code int, message string, data interface{}) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Data:    data,
	}
}