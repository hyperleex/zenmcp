package protocol

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

// testReadWriteCloser is a helper for testing that implements io.ReadWriteCloser.
// It uses a bytes.Reader for reading and a bytes.Buffer for writing.
type testReadWriteCloser struct {
	reader io.Reader
	writer *bytes.Buffer // Expose buffer for reading written data
	closed bool
}

// newTestReadWriteCloser creates a new testReadWriteCloser.
// input is the data to be read.
// outputBuffer is an optional buffer to use for writing. If nil, a new one is created.
func newTestReadWriteCloser(input []byte, outputBuffer *bytes.Buffer) *testReadWriteCloser {
	if outputBuffer == nil {
		outputBuffer = new(bytes.Buffer)
	}
	return &testReadWriteCloser{
		reader: bytes.NewReader(input),
		writer: outputBuffer,
	}
}

func (trwc *testReadWriteCloser) Read(p []byte) (n int, err error) {
	return trwc.reader.Read(p)
}

func (trwc *testReadWriteCloser) Write(p []byte) (n int, err error) {
	return trwc.writer.Write(p)
}

func (trwc *testReadWriteCloser) Close() error {
	trwc.closed = true
	// Try to close the reader if it's a closer
	if c, ok := trwc.reader.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (trwc *testReadWriteCloser) WrittenData() []byte {
	return trwc.writer.Bytes()
}

// Test Data Structures
type SimpleData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type ComplexData struct {
	ID      int      `json:"id"`
	Message string   `json:"message"`
	Tags    []string `json:"tags"`
}

type EmptyData struct{}

// TestJSONCodec (adapted from original, if it existed, or as a good practice)
// This test ensures the basic JSONCodec works as expected, as it's part of the package.
func TestJSONCodec(t *testing.T) {
	sharedBuffer := new(bytes.Buffer) // Data written by Encode will be read by Decode

	// Setup for encoding: write to sharedBuffer
	encodeTrwc := newTestReadWriteCloser(nil, sharedBuffer)
	jsonCodecEnc := NewJSONCodec(encodeTrwc)

	testData := SimpleData{Name: "json_codec_test", Value: 987}
	if err := jsonCodecEnc.Encode(testData); err != nil {
		t.Fatalf("JSONCodec.Encode() error = %v", err)
	}
	if err := jsonCodecEnc.Close(); err != nil { // Close after encoding
		t.Fatalf("JSONCodec.Encode().Close() error = %v", err)
	}


	// Setup for decoding: read from sharedBuffer (which now contains the encoded data)
	// Pass nil for outputBuffer as decode doesn't write using the main Write method.
	decodeTrwc := newTestReadWriteCloser(sharedBuffer.Bytes(), nil)
	jsonCodecDec := NewJSONCodec(decodeTrwc)

	var decodedData SimpleData
	if err := jsonCodecDec.Decode(&decodedData); err != nil {
		t.Fatalf("JSONCodec.Decode() error = %v", err)
	}

	if !reflect.DeepEqual(decodedData, testData) {
		t.Errorf("JSONCodec.Decode() got = %+v, want %+v", decodedData, testData)
	}
	if err := jsonCodecDec.Close(); err != nil { // Close after decoding
		t.Fatalf("JSONCodec.Decode().Close() error = %v", err)
	}
	if !decodeTrwc.closed { 
		t.Error("JSONCodec: Expected decode ReadWriteCloser to be closed")
	}
}


func TestLengthPrefixedCodec_Encode(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string // Expected output string (headers + json body)
		wantErr bool
	}{
		{
			name:  "simple data",
			input: SimpleData{Name: "test", Value: 123},
			want:  "Content-Length: 25\r\n\r\n{\"name\":\"test\",\"value\":123}",
		},
		{
			name:  "complex data with spaces and escapes",
			input: ComplexData{ID: 1, Message: "Hello \"world\"!\nNew line.", Tags: []string{"tag1", "tag with space"}},
			want:  "Content-Length: 87\r\n\r\n{\"id\":1,\"message\":\"Hello \\\"world\\\"!\\nNew line.\",\"tags\":[\"tag1\",\"tag with space\"]}",
		},
		{
			name:  "empty data struct",
			input: EmptyData{},
			want:  "Content-Length: 2\r\n\r\n{}",
		},
		{
			name:  "nil input (marshals to 'null')",
			input: nil,
			want:  "Content-Length: 4\r\n\r\nnull",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trwc := newTestReadWriteCloser(nil, nil) // Input is nil for encode, output buffer auto-created
			codec := NewLengthPrefixedCodec(trwc)

			err := codec.Encode(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("LengthPrefixedCodec.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				got := string(trwc.WrittenData())
				// Reconstruct expected string to ensure JSON part is identical to how Go marshals it
				expectedJSON, _ := json.Marshal(tt.input)
				expectedHeader := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(expectedJSON))
				expectedFullMessage := expectedHeader + string(expectedJSON)

				if got != expectedFullMessage {
					t.Errorf("LengthPrefixedCodec.Encode() got = \n%q\nwant = \n%q", got, expectedFullMessage)
					t.Logf("Got raw: %s", got) // Using Logf for more debug info
					t.Logf("Expected raw: %s", expectedFullMessage)
					t.Logf("Got JSON part: %s", got[strings.Index(got, "\r\n\r\n")+4:])
					t.Logf("Expected JSON part: %s", string(expectedJSON))
				}
			}
		})
	}
}

func TestLengthPrefixedCodec_Decode(t *testing.T) {
	type TargetType struct {
		Name    string   `json:"name,omitempty"`
		Value   int      `json:"value,omitempty"`
		ID      int      `json:"id,omitempty"`
		Message string   `json:"message,omitempty"`
		Tags    []string `json:"tags,omitempty"`
	}

	tests := []struct {
		name               string
		input              string
		want               TargetType
		wantErrIs          error  // For errors.Is checks (e.g., io.EOF)
		wantErrMsgContains string // For checking specific error messages from fmt.Errorf
	}{
		{
			name:  "simple data CRLF",
			input: "Content-Length: 25\r\n\r\n{\"name\":\"test\",\"value\":123}",
			want:  TargetType{Name: "test", Value: 123},
		},
		{
			name:  "header value with spaces (Custom-Header is ignored)",
			input: "Custom-Header: Value with spaces\r\nContent-Length: 26\r\n\r\n{\"message\":\"header space\"}", // Corrected CL from 30 to 26
			want:  TargetType{Message: "header space"},
		},
		{
			name:  "Content-Length not first header",
			input: "User-Agent: TestClient\r\nContent-Length: 26\r\nAccept: application/json\r\n\r\n{\"name\":\"order\",\"value\":456}", // JSON is 26 bytes
			want:  TargetType{Name: "order", Value: 456},
		},
		{
			name:  "case-insensitive Content-Length (content-length)",
			input: "content-length: 25\r\n\r\n{\"name\":\"case\",\"value\":789}",
			want:  TargetType{Name: "case", Value: 789},
		},
		{
			name:  "case-insensitive Content-Length (CONTENT-LENGTH)",
			input: "CONTENT-LENGTH: 25\r\n\r\n{\"name\":\"case\",\"value\":789}",
			want:  TargetType{Name: "case", Value: 789},
		},
		{
			name:  "Content-Length: 0 (empty json object)",
			input: "Content-Length: 2\r\n\r\n{}", // JSON "{}" is 2 bytes
			want:  TargetType{},
		},
		{
			name:  "Content-Length: 0 (json null)",
			input: "Content-Length: 4\r\n\r\nnull", // JSON "null" is 4 bytes
			want:  TargetType{}, // Unmarshals to zero struct
		},
		{
			name:               "Content-Length: 0 (empty body - invalid JSON)",
			input:              "Content-Length: 0\r\n\r\n",
			wantErrMsgContains: "unexpected end of JSON input", // Error from json.Unmarshal
		},
		{
			name:  "extraneous headers ignored",
			input: "X-Request-ID: 123\r\nContent-Length: 26\r\nCache-Control: no-cache\r\n\r\n{\"name\":\"extra\",\"value\":111}", // JSON is 26 bytes
			want:  TargetType{Name: "extra", Value: 111},
		},
		{
			name:  "Unix line endings (LF only)",
			input: "Content-Length: 25\n\n{\"name\":\"unix\",\"value\":222}",
			want:  TargetType{Name: "unix", Value: 222},
		},
		{
			name:  "Mixed line endings (CRLF and LF)",
			input: "User-Agent: TestClient\r\nContent-Length: 26\nAccept: application/json\n\n{\"name\":\"mixed\",\"value\":333}", // JSON is 26 bytes
			want:  TargetType{Name: "mixed", Value: 333},
		},
		{
			name:               "malformed Content-Length (non-numeric)",
			input:              "Content-Length: abc\r\n\r\n{\"name\":\"badval\"}",
			wantErrMsgContains: "invalid Content-Length value: \"abc\"",
		},
		{
			name:               "malformed Content-Length (negative)",
			input:              "Content-Length: -5\r\n\r\nhello",
			wantErrMsgContains: "invalid Content-Length: -5",
		},
		{
			name:               "missing Content-Length header",
			input:              "Some-Header: value\r\n\r\n{\"name\":\"missingcl\"}",
			wantErrMsgContains: "missing Content-Length header",
		},
		{
			name:      "premature end of headers (no blank line)",
			input:     "Content-Length: 5", 
			wantErrIs: io.ErrUnexpectedEOF,
		},
		{
			name:      "premature end of headers (header line abruptly ends)",
			input:     "Content-Length: 5\r\nUser-Agen", 
			wantErrIs: io.ErrUnexpectedEOF,
		},
		{
			name:  "header line with key but no value (ignored)",
			input: "MyHeader:\r\nContent-Length: 26\r\n\r\n{\"name\":\"noval\",\"value\":555}", // JSON is 26 bytes
			want:  TargetType{Name: "noval", Value: 555},
		},
		{
			name:  "header line with key and colon but empty value (ignored)",
			input: "MyHeader: \r\nContent-Length: 29\r\n\r\n{\"name\":\"emptyval\",\"value\":666}", // JSON is 29 bytes
			want:  TargetType{Name: "emptyval", Value: 666},
		},
		{
			name:      "no headers at all, just EOF",
			input:     "",
			wantErrIs: io.ErrUnexpectedEOF,
		},
		{
			name:               "only CRLF then EOF (empty headers)",
			input:              "\r\n\r\n",
			wantErrMsgContains: "missing Content-Length header",
		},
		{
			name:               "only LF then EOF (empty headers)",
			input:              "\n\n",
			wantErrMsgContains: "missing Content-Length header",
		},
		{
			name:  "header without colon, then valid C-L (malformed header skipped)",
			input: "InvalidHeaderLine\r\nContent-Length: 25\r\n\r\n{\"name\":\"skip\",\"value\":777}",
			want:  TargetType{Name: "skip", Value: 777},
		},
		{
			name:      "EOF after headers before body (Content-Length > 0)",
			input:     "Content-Length: 10\r\n\r\n", 
			wantErrIs: io.EOF, // ReadFull returns EOF if len(buf)>0 and 0 bytes read
		},
		{
			name:      "EOF in the middle of body",
			input:     "Content-Length: 10\r\n\r\nshort", 
			wantErrIs: io.ErrUnexpectedEOF,
		},
		{
			name:  "complex data decode",
			// Actual marshaled length of {"id":1,"message":"Hello \"world\"!\nNew line.","tags":["tag1","tag with space"]} is 79
			input: "Content-Length: 79\r\n\r\n{\"id\":1,\"message\":\"Hello \\\"world\\\"!\\nNew line.\",\"tags\":[\"tag1\",\"tag with space\"]}", // Corrected CL from 87 to 79
			want:  TargetType{ID: 1, Message: "Hello \"world\"!\nNew line.", Tags: []string{"tag1", "tag with space"}},
		},
		{
			name:  "Header with only spaces as key (ignored)",
			input: "   : value\r\nContent-Length: 2\r\n\r\n{}", // Key "   " becomes "" after TrimSpace
			want:  TargetType{},
		},
		{
			name:  "Header with key and value, but key becomes empty after trim (ignored)",
			input: "  : value\r\nContent-Length: 2\r\n\r\n{}", // Similar to above
			want:  TargetType{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trwc := newTestReadWriteCloser([]byte(tt.input), nil)
			codec := NewLengthPrefixedCodec(trwc)

			var got TargetType
			err := codec.Decode(&got)

			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Fatalf("Decode() error = %v (%T), wantErrIs %v (%T)", err, err, tt.wantErrIs, tt.wantErrIs)
				}
			} else if tt.wantErrMsgContains != "" {
				if err == nil {
					t.Fatalf("Decode() error = nil, want err containing %q", tt.wantErrMsgContains)
				}
				if !strings.Contains(err.Error(), tt.wantErrMsgContains) {
					t.Fatalf("Decode() error = %q, want err containing %q", err.Error(), tt.wantErrMsgContains)
				}
			} else { // No error expected
				if err != nil {
					t.Fatalf("Decode() error = %v, wantErr nil", err)
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Decode() got = %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func TestLengthPrefixedCodec_Decode_InterfaceTarget(t *testing.T) {
	input := "Content-Length: 25\r\n\r\n{\"name\":\"test\",\"value\":123}"
	trwc := newTestReadWriteCloser([]byte(input), nil)
	codec := NewLengthPrefixedCodec(trwc)

	var got interface{}
	err := codec.Decode(&got)
	if err != nil {
		t.Fatalf("Decode into interface{} failed: %v", err)
	}

	expected := map[string]interface{}{"name": "test", "value": 123.0} // JSON numbers unmarshal to float64 in interface{}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Decode into interface{} got = %#v (%T), want %#v (%T)", got, got, expected, expected)
	}
}

func TestLengthPrefixedCodec_Close(t *testing.T) {
	trwc := newTestReadWriteCloser(nil, nil)
	codec := NewLengthPrefixedCodec(trwc)
	if err := codec.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
	if !trwc.closed {
		t.Error("Expected underlying ReadWriteCloser to be closed")
	}
}
