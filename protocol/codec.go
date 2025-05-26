package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Codec interface {
	Encode(v interface{}) error
	Decode(v interface{}) error
	Close() error
}

type JSONCodec struct {
	enc *json.Encoder
	dec *json.Decoder
	rw  io.ReadWriteCloser
}

func NewJSONCodec(rw io.ReadWriteCloser) *JSONCodec {
	return &JSONCodec{
		enc: json.NewEncoder(rw),
		dec: json.NewDecoder(rw),
		rw:  rw,
	}
}

func (c *JSONCodec) Encode(v interface{}) error {
	return c.enc.Encode(v)
}

func (c *JSONCodec) Decode(v interface{}) error {
	return c.dec.Decode(v)
}

func (c *JSONCodec) Close() error {
	return c.rw.Close()
}

type LengthPrefixedCodec struct {
	rw     io.ReadWriteCloser
	reader *bufio.Reader
}

func NewLengthPrefixedCodec(rw io.ReadWriteCloser) *LengthPrefixedCodec {
	return &LengthPrefixedCodec{
		rw:     rw,
		reader: bufio.NewReader(rw),
	}
}

func (c *LengthPrefixedCodec) Encode(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := c.rw.Write([]byte(header)); err != nil {
		return err
	}
	
	_, err = c.rw.Write(data)
	return err
}

func (c *LengthPrefixedCodec) Decode(v interface{}) error {
	contentLength := -1
	contentLengthHeaderFound := false
	headerLinesRead := 0
	maxHeaders := 32 

	for {
		headerLinesRead++
		if headerLinesRead > maxHeaders {
			return fmt.Errorf("too many header lines (> %d)", maxHeaders)
		}

		line, errReadString := c.reader.ReadString('\n')

		if len(line) > 0 { // Process any data returned by ReadString, even if there's an error
			currentLine := strings.TrimSuffix(line, "\n")
			currentLine = strings.TrimSuffix(currentLine, "\r")

			if currentLine == "" { // Empty line marks end of headers
				break
			}

			parts := strings.SplitN(currentLine, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if strings.EqualFold(key, "Content-Length") {
					if value == "" { 
						return fmt.Errorf("invalid Content-Length value: empty")
					}
					parsedVal, errConv := strconv.Atoi(value)
					if errConv != nil { 
						return fmt.Errorf("invalid Content-Length value %q: %w", value, errConv)
					}
					contentLength = parsedVal
					contentLengthHeaderFound = true
				}
			}
		}

		// Now handle errors from ReadString
		if errReadString != nil {
			if errReadString == io.EOF { // EOF means end of input.
				break // Break from header loop. Checks after loop will see if CL was found.
			}
			// bufio.ErrBufferFull might occur if a header line is extremely long.
			// ReadString returns data read so far + ErrBufferFull. Data is processed above.
			// If we want to support >4k headers, more complex logic needed here.
			// For now, consider it an error that stops processing.
			if errReadString == bufio.ErrBufferFull {
				return fmt.Errorf("header line too long (exceeded bufio buffer): %w", errReadString)
			}
			return errReadString // Other unhandled error (e.g. underlying reader error)
		}
	}

	if !contentLengthHeaderFound {
		return fmt.Errorf("missing Content-Length header")
	}
	if contentLength < 0 {
		return fmt.Errorf("invalid Content-Length: %d", contentLength)
	}

	data := make([]byte, contentLength)
	n_read, bodyReadErr := io.ReadFull(c.reader, data)
	
	if bodyReadErr != nil {
		return fmt.Errorf("reading body: expected %d bytes, ReadFull read %d: %w", contentLength, n_read, bodyReadErr)
	}

	unmarshalErr := json.Unmarshal(data, v)
	if unmarshalErr != nil {
		preview := string(data)
		maxPreview := 100 
		if len(preview) > maxPreview {
			preview = preview[:maxPreview] + "..."
		}
		return fmt.Errorf("unmarshalling body (Content-Length: %d, ReadFull read: %d): %w; data preview: %q", contentLength, n_read, unmarshalErr, preview)
	}
	return nil
}

// The commented out previewBytes and hexDumpForError functions are confirmed to be absent
// in the current file state based on the previous read_files. This replacement block
// correctly represents the Decode and Close functions without them.

func (c *LengthPrefixedCodec) Close() error {
	return c.rw.Close()
}