package admin

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

func TestNewClient(t *testing.T) {
	client := resty.New()
	url := "http://example.com"
	authorizer := &MockAuthorizer{}

	adminClient := NewClient(client, url, authorizer)

	assert.NotNil(t, adminClient)
	assert.Equal(t, client, adminClient.client)
	assert.Equal(t, url, adminClient.url)
	assert.Equal(t, authorizer, adminClient.authorizer)
}

func TestClient_Authorize(t *testing.T) {
	client := resty.New()
	url := "http://example.com"
	
	t.Run("with authorizer", func(t *testing.T) {
		authorizer := &MockAuthorizer{}
		authorizer.On("Authorize").Return(nil)
		
		adminClient := NewClient(client, url, authorizer)
		err := adminClient.Authorize()
		
		assert.NoError(t, err)
		authorizer.AssertExpectations(t)
	})
	
	t.Run("without authorizer", func(t *testing.T) {
		adminClient := NewClient(client, url, nil)
		err := adminClient.Authorize()
		
		assert.NoError(t, err)
	})
}

func TestGetZIPName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "name without extension",
			input:    "backup",
			expected: "backup.zip",
		},
		{
			name:     "name with .zip extension",
			input:    "backup.zip",
			expected: "backup.zip",
		},
		{
			name:     "uppercase name",
			input:    "BACKUP",
			expected: "backup.zip",
		},
		{
			name:     "mixed case with extension",
			input:    "MyBackup.ZIP",
			expected: "mybackup.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetZIPName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResponseBackupFullList(t *testing.T) {
	response := ResponseBackupFullList{
		Key:      "backup_2023.zip",
		Size:     1024,
		Modified: "2023-01-01T00:00:00Z",
	}

	assert.Equal(t, "backup_2023.zip", response.Key)
	assert.Equal(t, 1024, response.Size)
	assert.Equal(t, "2023-01-01T00:00:00Z", response.Modified)
}

func TestCreateRequest(t *testing.T) {
	request := CreateRequest{
		Name: "test_backup.zip",
	}

	assert.Equal(t, "test_backup.zip", request.Name)
}

func TestBackup_GetDownloadURL(t *testing.T) {
	client := resty.New()
	url := "http://example.com"
	authorizer := &MockAuthorizer{}
	authorizer.On("Authorize").Return(nil)
	
	adminClient := NewClient(client, url, authorizer)
	backup := Backup{Client: adminClient}

	t.Run("build URL with token and key", func(t *testing.T) {
		downloadURL, err := backup.GetDownloadURL("test_token", "backup.zip")
		assert.NoError(t, err)
		assert.Equal(t, "http://example.com/api/backups/backup.zip?token=test_token", downloadURL)
	})

	t.Run("build URL with empty token", func(t *testing.T) {
		downloadURL, err := backup.GetDownloadURL("", "backup.zip")
		assert.Error(t, err)
		assert.Empty(t, downloadURL)
		assert.Contains(t, err.Error(), "missing token and/or key")
	})

	t.Run("build URL with empty key", func(t *testing.T) {
		downloadURL, err := backup.GetDownloadURL("test_token", "")
		assert.Error(t, err)
		assert.Empty(t, downloadURL)
		assert.Contains(t, err.Error(), "missing token and/or key")
	})
}