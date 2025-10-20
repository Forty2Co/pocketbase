package realtime

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthorizer implements Authorizer for testing
type MockAuthorizer struct {
	mock.Mock
}

func (m *MockAuthorizer) Authorize() error {
	args := m.Called()
	return args.Error(0)
}

// MockHTTPClient implements HTTPClient for testing
type MockHTTPClient struct {
	mock.Mock
	client *resty.Client
}

func (m *MockHTTPClient) R() *resty.Request {
	args := m.Called()
	return args.Get(0).(*resty.Request)
}

// MockCollection implements Collection for testing
type MockCollection struct {
	mock.Mock
}

func (m *MockCollection) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCollection) GetURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCollection) GetClient() HTTPClient {
	args := m.Called()
	return args.Get(0).(HTTPClient)
}

func (m *MockCollection) GetAuthorizer() Authorizer {
	args := m.Called()
	return args.Get(0).(Authorizer)
}

func (m *MockCollection) IsSSEDebugEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestNewClient(t *testing.T) {
	client := resty.New()
	url := "http://example.com"
	authorizer := &MockAuthorizer{}
	sseDebug := true

	realtimeClient := NewClient(client, url, authorizer, sseDebug)

	assert.NotNil(t, realtimeClient)
	assert.Equal(t, client, realtimeClient.client)
	assert.Equal(t, url, realtimeClient.url)
	assert.Equal(t, authorizer, realtimeClient.authorizer)
	assert.Equal(t, sseDebug, realtimeClient.sseDebug)
}

func TestEvent(t *testing.T) {
	event := Event[map[string]any]{
		Action: "create",
		Record: map[string]any{"id": "123", "name": "test"},
		Error:  nil,
	}

	assert.Equal(t, "create", event.Action)
	assert.Equal(t, "123", event.Record["id"])
	assert.Equal(t, "test", event.Record["name"])
	assert.NoError(t, event.Error)
}

func TestNewCollectionSubscriber(t *testing.T) {
	mockCollection := &MockCollection{}
	
	subscriber := NewCollectionSubscriber[map[string]any](mockCollection)
	
	assert.NotNil(t, subscriber)
	assert.Equal(t, mockCollection, subscriber.collection)
}

func TestStream_NewStream(t *testing.T) {
	stream := newStream[map[string]any]()
	
	assert.NotNil(t, stream)
	assert.NotNil(t, stream.channel)
	assert.NotNil(t, stream.ready)
	assert.NotNil(t, stream.onceCleanup)
}

func TestStream_Events(t *testing.T) {
	stream := newStream[map[string]any]()
	
	eventsChan := stream.Events()
	assert.NotNil(t, eventsChan)
	
	// Test that we can receive from the channel (it should be non-blocking for this test)
	select {
	case <-eventsChan:
		// Channel is ready to receive
	default:
		// Channel is not ready, which is expected for a new stream
	}
}