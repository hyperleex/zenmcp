package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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
	var contentLength int
	
	for {
		line, err := c.reader.ReadString('\n')
		if err != nil {
			return err
		}
		
		if len(line) == 0 {
			continue
		}
		
		line = line[:len(line)-1] // Remove \n
		if line == "\r" || line == "" {
			break
		}
		
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1] // Remove \r
		}
		
		var key, value string
		if n, err := fmt.Sscanf(line, "%s %s", &key, &value); err != nil || n != 2 {
			continue
		}
		
		if key == "Content-Length:" {
			if _, err := fmt.Sscanf(value, "%d", &contentLength); err != nil {
				return fmt.Errorf("invalid Content-Length: %s", value)
			}
		}
	}
	
	if contentLength <= 0 {
		return fmt.Errorf("missing or invalid Content-Length header")
	}
	
	data := make([]byte, contentLength)
	if _, err := io.ReadFull(c.reader, data); err != nil {
		return err
	}
	
	return json.Unmarshal(data, v)
}

func (c *LengthPrefixedCodec) Close() error {
	return c.rw.Close()
}