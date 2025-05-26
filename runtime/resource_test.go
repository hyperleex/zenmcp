package runtime

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	// "github.com/hyperleex/zenmcp/registry" // Not strictly needed if NewServer(nil) is fine
)

// --- Mock io.ReadCloser ---

type mockReadCloser struct {
	*strings.Reader
	closeFunc func() error
	closed    bool
}

func newMockReadCloser(content string, closeErr error) *mockReadCloser {
	mrc := &mockReadCloser{
		Reader: strings.NewReader(content),
		closed: false, // Explicitly initialize
	}
	mrc.closeFunc = func() error {
		mrc.closed = true
		return closeErr
	}
	return mrc
}

func (m *mockReadCloser) Close() error {
	// Ensure closed is set before returning, even if closeFunc is nil
	m.closed = true
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

type errorReader struct {
	closed bool
}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mock read error")
}

func (er *errorReader) Close() error {
	er.closed = true
	return nil
}

type errorReadCloserWithCloseError struct {
	readErr  error
	closeErr error
	closed   bool
}

func (er *errorReadCloserWithCloseError) Read(p []byte) (n int, err error) {
	return 0, er.readErr
}

func (er *errorReadCloserWithCloseError) Close() error {
	er.closed = true
	return er.closeErr
}

// --- Helper Functions ---

func newTestResource(uri, name, mimeType string, readerContent *string, readerCreationErr error, readErr error, closeErr error) Resource {
	return Resource{
		URI:      uri,
		Name:     name,
		MimeType: mimeType,
		Reader: func() (io.ReadCloser, error) {
			if readerCreationErr != nil {
				return nil, readerCreationErr
			}
			if readerContent == nil {
				return nil, errors.New("reader content is nil, but no creation error specified")
			}
			if readErr != nil {
				// Use errorReadCloserWithCloseError to handle both read and close errors
				return &errorReadCloserWithCloseError{readErr: readErr, closeErr: closeErr, closed: false}, nil
			}
			return newMockReadCloser(*readerContent, closeErr), nil
		},
	}
}

func newMockProvider(resources []Resource, err error) func(context.Context) ([]Resource, error) {
	return func(ctx context.Context) ([]Resource, error) {
		if err != nil {
			return nil, err
		}
		return resources, nil
	}
}

// --- Tests for handleResourcesList ---

func TestHandleResourcesList_Empty(t *testing.T) {
	server := NewServer(nil) // Pass nil for registry
	// Directly call the unexported method handleResourcesList
	result, err := server.handleResourcesList(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	resources, ok := result.([]Resource)
	if !ok {
		t.Fatalf("expected result to be []Resource, got %T", result)
	}
	if len(resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(resources))
	}
}

func TestHandleResourcesList_SingleProviderSingleResource(t *testing.T) {
	server := NewServer(nil)
	res := newTestResource("uri1", "name1", "mime1", nil, nil, nil, nil)
	server.Resources("provider1", newMockProvider([]Resource{res}, nil))

	result, err := server.handleResourcesList(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	resources, ok := result.([]Resource)
	if !ok {
		t.Fatalf("expected result to be []Resource, got %T", result)
	}
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	expectedRes := Resource{URI: "uri1", Name: "name1", MimeType: "mime1"}
	gotRes := Resource{URI: resources[0].URI, Name: resources[0].Name, MimeType: resources[0].MimeType}
	if !reflect.DeepEqual(gotRes, expectedRes) {
		t.Errorf("expected resource %+v, got %+v", expectedRes, gotRes)
	}
}

func TestHandleResourcesList_SingleProviderMultipleResources(t *testing.T) {
	server := NewServer(nil)
	res1 := newTestResource("uri1", "name1", "mime1", nil, nil, nil, nil)
	res2 := newTestResource("uri2", "name2", "mime2", nil, nil, nil, nil)
	server.Resources("provider1", newMockProvider([]Resource{res1, res2}, nil))

	result, err := server.handleResourcesList(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	resources, ok := result.([]Resource)
	if !ok {
		t.Fatalf("expected result to be []Resource, got %T", result)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}
	expectedResources := []Resource{
		{URI: "uri1", Name: "name1", MimeType: "mime1"},
		{URI: "uri2", Name: "name2", MimeType: "mime2"},
	}
	// Create comparable versions (without Reader func)
	gotComparableResources := make([]Resource, len(resources))
	for i, r := range resources {
		gotComparableResources[i] = Resource{URI: r.URI, Name: r.Name, MimeType: r.MimeType}
	}

	if !reflect.DeepEqual(gotComparableResources, expectedResources) {
		t.Errorf("expected resources %+v, got %+v", expectedResources, gotComparableResources)
	}
}

func TestHandleResourcesList_MultipleProviders(t *testing.T) {
	server := NewServer(nil)
	res1 := newTestResource("uri1", "name1", "mime1", nil, nil, nil, nil)
	res2 := newTestResource("uri2", "name2", "mime2", nil, nil, nil, nil)
	server.Resources("provider1", newMockProvider([]Resource{res1}, nil))
	server.Resources("provider2", newMockProvider([]Resource{res2}, nil))

	result, err := server.handleResourcesList(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	resources, ok := result.([]Resource)
	if !ok {
		t.Fatalf("expected result to be []Resource, got %T", result)
	}
	if len(resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(resources))
	}
	expectedURIs := map[string]bool{"uri1": true, "uri2": true}
	for _, r := range resources {
		if !expectedURIs[r.URI] {
			t.Errorf("unexpected resource URI %s found", r.URI)
		}
		delete(expectedURIs, r.URI) // Mark as found
	}
	if len(expectedURIs) != 0 {
		t.Errorf("not all expected URIs found, missing: %v", expectedURIs)
	}
}

func TestHandleResourcesList_ProviderError(t *testing.T) {
	server := NewServer(nil)
	providerErr := errors.New("provider error")
	res1 := newTestResource("uri1", "name1", "mime1", nil, nil, nil, nil)
	server.Resources("providerError", newMockProvider(nil, providerErr))
	server.Resources("providerSuccess", newMockProvider([]Resource{res1}, nil))

	result, err := server.handleResourcesList(context.Background()) // This is the method under test
	if err != nil {
		t.Fatalf("expected no error from handleResourcesList itself, got %v", err)
	}
	resources, ok := result.([]Resource)
	if !ok {
		t.Fatalf("expected result to be []Resource, got %T", result)
	}
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource from the succeeding provider, got %d", len(resources))
	}
	if resources[0].URI != "uri1" {
		t.Errorf("expected resource URI 'uri1', got '%s'", resources[0].URI)
	}
}

func TestHandleResourcesList_CorrectData(t *testing.T) {
	server := NewServer(nil)
	res := newTestResource("test/uri", "Test Resource", "application/test", nil, nil, nil, nil)
	server.Resources("provider1", newMockProvider([]Resource{res}, nil))

	result, err := server.handleResourcesList(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	resources, ok := result.([]Resource)
	if !ok {
		t.Fatalf("expected result to be []Resource, got %T", result)
	}
	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}
	expected := Resource{URI: "test/uri", Name: "Test Resource", MimeType: "application/test"}
	got := Resource{URI: resources[0].URI, Name: resources[0].Name, MimeType: resources[0].MimeType}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected resource data %+v, got %+v", expected, got)
	}
}

// --- Tests for handleResourcesRead ---

func TestHandleResourcesRead_Success(t *testing.T) {
	server := NewServer(nil)
	content := "test content"
	resURI := "uri/read/success"

	var testReaderInstance *mockReadCloser // To check .closed status
	providerFunc := func(ctx context.Context) ([]Resource, error) {
		// Create the mockReadCloser here so we can get a reference to it.
		testReaderInstance = newMockReadCloser(content, nil)
		return []Resource{
			{
				URI:      resURI,
				Name:     "name",
				MimeType: "text/plain",
				Reader: func() (io.ReadCloser, error) {
					return testReaderInstance, nil
				},
			},
		}, nil
	}
	server.Resources("providerRead", providerFunc)

	params := ResourcesReadParams{URI: resURI}
	// Directly call the unexported method handleResourcesRead
	result, err := server.handleResourcesRead(context.Background(), params)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	data, ok := result.([]byte)
	if !ok {
		t.Fatalf("expected result to be []byte, got %T", result)
	}
	if string(data) != content {
		t.Errorf("expected content '%s', got '%s'", content, string(data))
	}
	if testReaderInstance == nil { // Should be set by the provider
		t.Fatal("testReaderInstance was not initialized by the provider")
	}
	if !testReaderInstance.closed {
		t.Errorf("expected reader to be closed")
	}
}

func TestHandleResourcesRead_NotFound(t *testing.T) {
	server := NewServer(nil)
	content := "test content"
	res := newTestResource("uri/exists", "name", "mime", &content, nil, nil, nil)
	server.Resources("provider1", newMockProvider([]Resource{res}, nil))

	params := ResourcesReadParams{URI: "uri/does/not/exist"}
	_, err := server.handleResourcesRead(context.Background(), params)

	if !errors.Is(err, ErrResourceNotFound) {
		t.Fatalf("expected ErrResourceNotFound, got %v", err)
	}
}

func TestHandleResourcesRead_ProviderErrorOnDiscovery(t *testing.T) {
	server := NewServer(nil)
	providerErr := errors.New("provider discovery error")
	server.Resources("providerError", newMockProvider(nil, providerErr))

	params := ResourcesReadParams{URI: "any/uri"}
	_, err := server.handleResourcesRead(context.Background(), params)

	if !errors.Is(err, ErrResourceNotFound) {
		t.Fatalf("expected ErrResourceNotFound due to provider error and no other provider having the resource, got %v", err)
	}
}

func TestHandleResourcesRead_ReaderCreationError(t *testing.T) {
	server := NewServer(nil)
	creationErr := errors.New("reader creation error")
	resURI := "uri/reader/creation/error"
	res := newTestResource(resURI, "name", "mime", nil, creationErr, nil, nil)
	server.Resources("provider1", newMockProvider([]Resource{res}, nil))

	params := ResourcesReadParams{URI: resURI}
	_, err := server.handleResourcesRead(context.Background(), params)

	if err == nil || !strings.Contains(err.Error(), creationErr.Error()) {
		t.Fatalf("expected error containing '%s', got %v", creationErr.Error(), err)
	}
}

func TestHandleResourcesRead_ReadError(t *testing.T) {
	server := NewServer(nil)
	resURI := "uri/read/error"
	readErr := errors.New("mock read error") // This is the specific error from errorReader
	// The 'content' variable was here but was unused. It has been removed.

	var testErrorReader *errorReadCloserWithCloseError
	providerFunc := func(ctx context.Context) ([]Resource, error) {
		testErrorReader = &errorReadCloserWithCloseError{readErr: readErr, closeErr: nil, closed: false}
		return []Resource{
			{
				URI:      resURI,
				Name:     "name",
				MimeType: "text/plain",
				Reader: func() (io.ReadCloser, error) {
					return testErrorReader, nil
				},
			},
		}, nil
	}
	server.Resources("providerReadError", providerFunc)

	params := ResourcesReadParams{URI: resURI}
	_, err := server.handleResourcesRead(context.Background(), params)

	if err == nil || !strings.Contains(err.Error(), readErr.Error()) {
		t.Fatalf("expected error containing '%s', got %v", readErr.Error(), err)
	}
	if testErrorReader == nil {
		t.Fatal("testErrorReader was not initialized by the provider")
	}
	if !testErrorReader.closed {
		t.Errorf("expected reader to be closed even on read error")
	}
}

func TestHandleResourcesRead_ReaderCloseError(t *testing.T) {
	server := NewServer(nil)
	content := "test content for close error"
	resURI := "uri/read/success/closeerror"
	closeErr := errors.New("mock close error")

	var testReaderInstance *mockReadCloser
	providerFunc := func(ctx context.Context) ([]Resource, error) {
		testReaderInstance = newMockReadCloser(content, closeErr)
		return []Resource{
			{
				URI:      resURI,
				Name:     "name",
				MimeType: "text/plain",
				Reader: func() (io.ReadCloser, error) {
					return testReaderInstance, nil
				},
			},
		}, nil
	}
	server.Resources("providerReadCloseError", providerFunc)

	params := ResourcesReadParams{URI: resURI}
	result, err := server.handleResourcesRead(context.Background(), params)

	// The error from Close() in a defer is typically not returned by the function
	// unless the defer func modifies a named return value.
	// Current implementation of handleResourcesRead does: defer reader.Close().
	// This error will be "lost" from the perspective of the caller of handleResourcesRead.
	if err != nil {
		t.Fatalf("expected no error from handleResourcesRead due to Close error, got %v", err)
	}
	data, ok := result.([]byte)
	if !ok {
		t.Fatalf("expected result to be []byte, got %T", result)
	}
	if string(data) != content {
		t.Errorf("expected content '%s', got '%s'", content, string(data))
	}
	if testReaderInstance == nil {
		t.Fatal("testReaderInstance was not initialized by the provider")
	}
	if !testReaderInstance.closed {
		t.Errorf("expected reader to be closed")
	}
	// The actual closeErr is not checked here as it's not propagated by handleResourcesRead.
}

func TestHandleResourcesRead_NilReaderFuncInResource(t *testing.T) {
	server := NewServer(nil)
	resURI := "uri/nil/reader"
	res := Resource{
		URI:      resURI,
		Name:     "Nil Reader Resource",
		MimeType: "application/octet-stream",
		Reader:   nil, // Explicitly nil reader func
	}
	server.Resources("providerWithNilReader", newMockProvider([]Resource{res}, nil))

	params := ResourcesReadParams{URI: resURI}
	_, err := server.handleResourcesRead(context.Background(), params)

	if err == nil {
		t.Fatalf("expected an error for nil reader func, got nil")
	}
	expectedErrorMsg := "resource " + resURI + " has no reader defined"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("expected error message to contain '%s', got '%s'", expectedErrorMsg, err.Error())
	}
}

func TestHandleResourcesRead_NotFoundInPopulatedProvider(t *testing.T) {
	server := NewServer(nil)
	content := "extant content"
	resExisting := newTestResource("uri/extant", "Extant Name", "text/plain", &content, nil, nil, nil)
	server.Resources("populatedProvider", newMockProvider([]Resource{resExisting}, nil))

	params := ResourcesReadParams{URI: "uri/nonextant"}
	_, err := server.handleResourcesRead(context.Background(), params)

	if !errors.Is(err, ErrResourceNotFound) {
		t.Fatalf("expected ErrResourceNotFound, got %v (URI: %s)", err, params.URI)
	}
}

// Test case where a resource reader function returns a nil reader and nil error
// This is a potentially problematic case for the main code.
func TestHandleResourcesRead_ReaderFuncReturnsNilReaderNilError(t *testing.T) {
	server := NewServer(nil)
	resURI := "uri/nilreader/nilerror"
	res := Resource{
		URI:      resURI,
		Name:     "Nil Reader Nil Error",
		MimeType: "application/octet-stream",
		Reader: func() (io.ReadCloser, error) {
			return nil, nil // Problematic case
		},
	}
	server.Resources("providerNilReaderNilError", newMockProvider([]Resource{res}, nil))

	params := ResourcesReadParams{URI: resURI}
	_, err := server.handleResourcesRead(context.Background(), params)

	if err == nil {
		t.Fatalf("expected an error when reader func returns nil, nil, got nil")
	}
	// Check for a specific error message. The current code would likely panic or return a generic error.
	// The current handleResourcesRead code:
	// reader, err := resource.Reader()
	// if err != nil { return nil, fmt.Errorf("failed to get reader for resource %s: %w", params.URI, err) }
	// // What if reader is nil and err is nil?
	// defer reader.Close() // This would panic if reader is nil.
	// Let's assume the code should handle this, e.g. by returning an error.
	// The previous fix in handleResourcesRead (Subtask 8) ensures it returns a specific error
	// when a Reader() func returns (nil, nil). This test now asserts that specific error.

	if err == nil {
		t.Fatalf("expected an error when reader func returns (nil, nil), but got nil")
	}

	// Assert the specific error message from the fix applied in subtask 8.
	// The fix returns: fmt.Errorf("reader for resource %q is nil, but no error was reported by the reader function", params.URI)
	// Note: params.URI would be resURI in this test's context.
	expectedSpecificErrorMsg := fmt.Sprintf("reader for resource %q is nil, but no error was reported by the reader function", resURI)
	if !strings.Contains(err.Error(), expectedSpecificErrorMsg) {
		t.Errorf("expected error message to contain %q, got %q", expectedSpecificErrorMsg, err.Error())
	}
}
