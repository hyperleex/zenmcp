package runtime

import (
	"context"
	"testing"

	"github.com/hyperleex/zenmcp/protocol"
)

func TestNewContext(t *testing.T) {
	ctx := context.Background()
	requestID := protocol.NewRequestID("test-123")
	
	runtimeCtx := NewContext(ctx, requestID)
	
	if runtimeCtx.Context != ctx {
		t.Error("Expected same underlying context")
	}
	
	if runtimeCtx.RequestID() != requestID {
		t.Error("Expected same request ID")
	}
}

func TestContext_WithProgressToken(t *testing.T) {
	ctx := context.Background()
	requestID := protocol.NewRequestID("test-123")
	runtimeCtx := NewContext(ctx, requestID)
	
	token := &protocol.ProgressToken{}
	runtimeCtx2 := runtimeCtx.WithProgressToken(token)
	
	if runtimeCtx2.ProgressToken() != token {
		t.Error("Expected same progress token")
	}
}

func TestContext_SetProgress(t *testing.T) {
	ctx := context.Background()
	requestID := protocol.NewRequestID("test-123")
	runtimeCtx := NewContext(ctx, requestID)
	
	total := 100.0
	runtimeCtx.SetProgress(50.0, &total)
	
	progress, gotTotal := runtimeCtx.Progress()
	if progress != 50.0 {
		t.Errorf("Expected progress 50.0, got %f", progress)
	}
	
	if gotTotal == nil || *gotTotal != 100.0 {
		t.Errorf("Expected total 100.0, got %v", gotTotal)
	}
}

func TestContext_Cancel(t *testing.T) {
	ctx := context.Background()
	requestID := protocol.NewRequestID("test-123")
	runtimeCtx := NewContext(ctx, requestID)
	
	if runtimeCtx.IsCancelled() {
		t.Error("Expected context not to be cancelled initially")
	}
	
	runtimeCtx.Cancel()
	
	if !runtimeCtx.IsCancelled() {
		t.Error("Expected context to be cancelled after Cancel()")
	}
}