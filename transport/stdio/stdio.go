package stdio

import (
	"context"
	"os"

	"github.com/hyperleex/zenmcp/protocol"
	"github.com/hyperleex/zenmcp/transport"
)

type Transport struct {
	codec    protocol.Codec
	accepted bool
}

func New() *Transport {
	return &Transport{}
}

func (t *Transport) Accept(ctx context.Context) (transport.Connection, error) {
	if t.accepted {
		// stdio transport can only have one connection
		<-ctx.Done()
		return nil, ctx.Err()
	}
	
	if t.codec == nil {
		rwc := &stdioReadWriteCloser{
			reader: os.Stdin,
			writer: os.Stdout,
		}
		t.codec = protocol.NewLengthPrefixedCodec(rwc)
	}
	
	t.accepted = true
	return transport.NewConnection(ctx, t.codec), nil
}

func (t *Transport) Close() error {
	if t.codec != nil {
		return t.codec.Close()
	}
	return nil
}

type stdioReadWriteCloser struct {
	reader *os.File
	writer *os.File
}

func (rw *stdioReadWriteCloser) Read(p []byte) (n int, err error) {
	return rw.reader.Read(p)
}

func (rw *stdioReadWriteCloser) Write(p []byte) (n int, err error) {
	return rw.writer.Write(p)
}

func (rw *stdioReadWriteCloser) Close() error {
	return nil
}