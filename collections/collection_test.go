package collections

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClient implements ClientInterface for testing
type MockClient struct {
	mock.Mock
}

func (m *MockClient) Update(collection string, id string, body any) error {
	args := m.Called(collection, id, body)
	return args.Error(0)
}

func (m *MockClient) Create(collection string, body any) (ResponseCreate, error) {
	args := m.Called(collection, body)
	return args.Get(0).(ResponseCreate), args.Error(1)
}

func (m *MockClient) Delete(collection string, id string) error {
	args := m.Called(collection, id)
	return args.Error(0)
}

func (m *MockClient) List(collection string, params ParamsList) (ResponseList[map[string]any], error) {
	args := m.Called(collection, params)
	return args.Get(0).(ResponseList[map[string]any]), args.Error(1)
}

func (m *MockClient) FullList(collection string, params ParamsList) (ResponseList[map[string]any], error) {
	args := m.Called(collection, params)
	return args.Get(0).(ResponseList[map[string]any]), args.Error(1)
}

func (m *MockClient) Authorize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockClient) SetToken(token string) {
	m.Called(token)
}

func (m *MockClient) GetToken() string {
	args := m.Called()
	return args.String(0)
}

func TestCollectionSet(t *testing.T) {
	mockClient := &MockClient{}
	collectionName := "test_collection"

	collection := CollectionSet[map[string]any](mockClient, collectionName)

	assert.NotNil(t, collection)
	assert.Equal(t, collectionName, collection.Name)
	assert.Equal(t, mockClient, collection.Client)
	assert.Contains(t, collection.BaseCollectionPath, collectionName)
}

func TestCollection_Update(t *testing.T) {
	mockClient := &MockClient{}
	collection := CollectionSet[map[string]any](mockClient, "test")

	testData := map[string]any{"field": "value"}
	mockClient.On("Update", "test", "123", testData).Return(nil)

	err := collection.Update("123", testData)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestCollection_Create(t *testing.T) {
	mockClient := &MockClient{}
	collection := CollectionSet[map[string]any](mockClient, "test")

	testData := map[string]any{"field": "value"}
	expectedResponse := ResponseCreate{ID: "new_id", Created: "2023-01-01"}
	
	mockClient.On("Create", "test", testData).Return(expectedResponse, nil)

	response, err := collection.Create(testData)

	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)
	mockClient.AssertExpectations(t)
}

func TestCollection_Delete(t *testing.T) {
	mockClient := &MockClient{}
	collection := CollectionSet[map[string]any](mockClient, "test")

	mockClient.On("Delete", "test", "123").Return(nil)

	err := collection.Delete("123")

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestCollection_List(t *testing.T) {
	mockClient := &MockClient{}
	collection := CollectionSet[map[string]any](mockClient, "test")

	params := ParamsList{Page: 1, Size: 10}

	// The mock should expect the params with HackResponseRef set
	mockClient.On("List", "test", mock.MatchedBy(func(p ParamsList) bool {
		return p.Page == 1 && p.Size == 10 && p.HackResponseRef != nil
	})).Return(ResponseList[map[string]any]{}, nil)

	response, err := collection.List(params)

	assert.NoError(t, err)
	// The response should be of the correct type even if empty due to mocking
	assert.IsType(t, ResponseList[map[string]any]{}, response)
	mockClient.AssertExpectations(t)
}

func TestCollection_FullList(t *testing.T) {
	mockClient := &MockClient{}
	collection := CollectionSet[map[string]any](mockClient, "test")

	params := ParamsList{Filters: "field='test'"}

	mockClient.On("FullList", "test", mock.MatchedBy(func(p ParamsList) bool {
		return p.Filters == "field='test'" && p.HackResponseRef != nil
	})).Return(ResponseList[map[string]any]{}, nil)

	response, err := collection.FullList(params)

	assert.NoError(t, err)
	assert.IsType(t, ResponseList[map[string]any]{}, response)
	mockClient.AssertExpectations(t)
}

func TestCollection_TokenManagement(t *testing.T) {
	mockClient := &MockClient{}
	collection := CollectionSet[map[string]any](mockClient, "test")

	// Test setting and getting token
	testToken := "test_token_123"
	collection.setToken(testToken)
	
	assert.Equal(t, testToken, collection.getToken())
}

func TestCollection_TokenFallback(t *testing.T) {
	mockClient := &MockClient{}
	collection := CollectionSet[map[string]any](mockClient, "test")

	// Test that when collection token is empty, it falls back to client token
	clientToken := "client_token_456"
	mockClient.On("GetToken").Return(clientToken)

	// Collection token is empty, should fall back to client
	assert.Equal(t, "", collection.getToken())
	
	// This simulates the fallback logic used in AuthRefresh
	token := collection.getToken()
	if token == "" {
		token = collection.Client.GetToken()
	}
	
	assert.Equal(t, clientToken, token)
	mockClient.AssertExpectations(t)
}

func TestCollection_GetName(t *testing.T) {
	mockClient := &MockClient{}
	collectionName := "users"
	collection := CollectionSet[map[string]any](mockClient, collectionName)

	assert.Equal(t, collectionName, collection.GetName())
}

func TestCollection_BaseCrudPath(t *testing.T) {
	mockClient := &MockClient{}
	collection := CollectionSet[map[string]any](mockClient, "test")

	path := collection.baseCrudPath()
	assert.Contains(t, path, "/records/")
	assert.Contains(t, path, "test")
}

func TestResponseTypes(t *testing.T) {
	// Test ResponseList
	response := ResponseList[map[string]any]{
		Page:       1,
		PerPage:    10,
		TotalItems: 25,
		TotalPages: 3,
		Items:      []map[string]any{{"id": "1"}},
	}
	
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.PerPage)
	assert.Equal(t, 25, response.TotalItems)
	assert.Equal(t, 3, response.TotalPages)
	assert.Len(t, response.Items, 1)

	// Test ResponseCreate
	createResponse := ResponseCreate{
		ID:      "new_id",
		Created: "2023-01-01T00:00:00Z",
		Updated: "2023-01-01T00:00:00Z",
	}
	
	assert.Equal(t, "new_id", createResponse.ID)
	assert.Equal(t, "2023-01-01T00:00:00Z", createResponse.Created)
}

func TestParamsList(t *testing.T) {
	params := ParamsList{
		Page:    1,
		Size:    20,
		Filters: "status='active'",
		Sort:    "-created",
		Expand:  "user",
		Fields:  "id,name,email",
	}

	assert.Equal(t, 1, params.Page)
	assert.Equal(t, 20, params.Size)
	assert.Equal(t, "status='active'", params.Filters)
	assert.Equal(t, "-created", params.Sort)
	assert.Equal(t, "user", params.Expand)
	assert.Equal(t, "id,name,email", params.Fields)
}

func TestAuthTypes(t *testing.T) {
	// Test AuthWithPasswordResponse
	authResponse := AuthWithPasswordResponse{
		Record: Record{
			ID:       "user_123",
			Email:    "test@example.com",
			Username: "testuser",
			Verified: true,
		},
		Token: "jwt_token_here",
	}

	assert.Equal(t, "user_123", authResponse.Record.ID)
	assert.Equal(t, "test@example.com", authResponse.Record.Email)
	assert.Equal(t, "jwt_token_here", authResponse.Token)
	assert.True(t, authResponse.Record.Verified)

	// Test AuthMethodsResponse
	authMethods := AuthMethodsResponse{
		Password: passwordResponse{
			Enabled:        true,
			IdentityFields: []string{"email", "username"},
		},
		OAuth2: oauth2Response{
			Enabled: false,
		},
		EmailPassword: true,
	}

	assert.True(t, authMethods.Password.Enabled)
	assert.False(t, authMethods.OAuth2.Enabled)
	assert.True(t, authMethods.EmailPassword)
	assert.Contains(t, authMethods.Password.IdentityFields, "email")
}