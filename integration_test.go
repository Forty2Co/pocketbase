package pocketbase_test

import (
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Forty2Co/pocketbase"
	"github.com/Forty2Co/pocketbase/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultURL = "http://127.0.0.1:8090"
)

type User struct {
	AuthProviders    []interface{} `json:"authProviders"`
	UsernamePassword bool          `json:"usernamePassword"`
	EmailPassword    bool          `json:"emailPassword"`
	OnlyVerified     bool          `json:"onlyVerified"`
}

// Auth Integration Tests
func TestCollection_ListAuthMethods_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("get AuthMethods with invalid authorization", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword("foo", "bar"))

		resp, err := pocketbase.CollectionSet[User](defaultClient, "users").ListAuthMethods()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to authenticate.")
		assert.Empty(t, resp)
	})

	t.Run("get AuthMethods with valid authorization", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))

		resp, err := pocketbase.CollectionSet[User](defaultClient, "users").ListAuthMethods()
		assert.NoError(t, err)
		assert.True(t, resp.Password.Enabled)
		assert.False(t, resp.OAuth2.Enabled)
		assert.False(t, resp.MFA.Enabled)
		assert.False(t, resp.OTP.Enabled)
		assert.Empty(t, resp.AuthProviders)
		assert.False(t, resp.UsernamePassword)
		assert.True(t, resp.EmailPassword)
	})
}

func TestCollection_AuthWithPassword_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("authenticate with valid user credentials", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		response, err := pocketbase.CollectionSet[User](defaultClient, "users").AuthWithPassword("user@user.com", "user@user.com")
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Token)
		assert.Len(t, response.Token, 224)
		assert.Equal(t, response.Token, defaultClient.GetToken())
	})

	t.Run("authenticate with invalid user credentials", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		response, err := pocketbase.CollectionSet[User](defaultClient, "users").AuthWithPassword("foo", "bar")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to authenticate")
		assert.Empty(t, response.Token)
		assert.Len(t, response.Token, 0)
		assert.Equal(t, response.Token, defaultClient.GetToken())
	})
}

func TestCollection_AuthRefresh_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("refresh authentication without valid user auth token", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		_, err := pocketbase.CollectionSet[User](defaultClient, "users").AuthRefresh()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})

	t.Run("refresh authentication with invalid user auth token", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		defaultClient.SetToken(strings.Repeat("X", 207))
		_, err := pocketbase.CollectionSet[User](defaultClient, "users").AuthRefresh()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})

	t.Run("refresh authentication with valid user auth token", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		authResponse, err := pocketbase.CollectionSet[User](defaultClient, "users").AuthWithPassword("user@user.com", "user@user.com")
		require.NoError(t, err)
		require.NotEmpty(t, authResponse.Token)
		oldToken := authResponse.Token

		time.Sleep(1 * time.Second) // we need to wait to get another token expire time

		response, err := pocketbase.CollectionSet[User](defaultClient, "users").AuthRefresh()
		assert.NoError(t, err)
		assert.NotEmpty(t, response.Token)
		assert.Len(t, response.Token, 224)
		assert.Equal(t, response.Token, defaultClient.GetToken())
		assert.NotEqual(t, response.Token, oldToken)
	})
}

func TestCollection_RequestVerification_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("request verification with valid authorization and not existing user", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))

		err := pocketbase.CollectionSet[User](defaultClient, "users").RequestVerification("nouser@nouser.com")
		assert.NoError(t, err)
	})

	t.Run("request verification with valid authorization", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))

		err := pocketbase.CollectionSet[User](defaultClient, "users").RequestVerification("user@user.com")
		assert.NoError(t, err)
	})
}

func TestCollection_ConfirmVerification_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("confirm verification with an invalid verification token", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		err := pocketbase.CollectionSet[User](defaultClient, "users").ConfirmVerification("no-valid-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation_invalid_token_claims")
	})

	t.Run("confirm verification with an valid token but not for the test-environment verification token", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		err := pocketbase.CollectionSet[User](defaultClient, "users").ConfirmVerification("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiJfcGJfdXNlcnNfYXV0aF8iLCJlbWFpbCI6InVzZXJAdXNlci5jb20iLCJleHAiOjE3MTQwNzE0MzgsImlkIjoiOHZ4OWh1ZDZkZXAyMnV2IiwidHlwZSI6ImF1dGhSZWNvcmQifQ.UwHOhmd0F_kK4LdjvDYqzE7QMheXmIiipFM6i-gwEPQ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation_invalid_token")
	})
}

func TestCollection_RequestPasswordReset_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("request password reset with valid authorization and not existing user", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		err := pocketbase.CollectionSet[User](defaultClient, "users").RequestPasswordReset("nouser@nouser.com")
		assert.NoError(t, err)
	})

	t.Run("request password reset with valid authorization", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		err := pocketbase.CollectionSet[User](defaultClient, "users").RequestPasswordReset("user@user.com")
		assert.NoError(t, err)
	})
}

func TestCollection_ConfirmPasswordReset_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("confirm password reset with an invalid verification token", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		err := pocketbase.CollectionSet[User](defaultClient, "users").ConfirmPasswordReset("no-valid-token", "new-password-123", "new-password-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation_invalid_token")
	})

	t.Run("confirm password reset with an valid token but not for the test-environment verification token", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		err := pocketbase.CollectionSet[User](defaultClient, "users").ConfirmPasswordReset("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiJfcGJfdXNlcnNfYXV0aCIsImVtYWlsIjoidXNlckB1c2VyLmNvbSIsImV4cCI6MTcxMzQ3MTc5NSwiaWQiOiI4dng5aHVkNmRlcDIydXYiLCJ0eXBlIjoiYXV0aFJlY29yZCJ9.u_7_u1t0MueFfKAMmXPqe4o1mNBn_-oFEpdSSeGqlUs",
			"new-password-123",
			"new-password-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation_invalid_token")
	})
}

func TestCollection_RequestEmailChange_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("confirm pemail change without a valid login", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		err := pocketbase.CollectionSet[User](defaultClient, "users").RequestEmailChange("useruser@user.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})
}

func TestCollection_ConfirmEmailChange_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("confirm email change with an invalid verification token", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		err := pocketbase.CollectionSet[User](defaultClient, "users").ConfirmEmailChange("no-valid-token", "new-password-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation_invalid_token_payload")
	})

	t.Run("confirm email change with an valid token but not for the test-environment verification token", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)

		err := pocketbase.CollectionSet[User](defaultClient, "users").ConfirmEmailChange("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiJfcGJfdXNlcnNfYXV0aCIsImVtYWlsIjoidXNlckB1c2VyLmNvbSIsImV4cCI6MTcxMzQ3MTc5NSwiaWQiOiI4dng5aHVkNmRlcDIydXYiLCJ0eXBlIjoiYXV0aFJlY29yZCJ9.u_7_u1t0MueFfKAMmXPqe4o1mNBn_-oFEpdSSeGqlUs",
			"user@user.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation_invalid_token")
	})
}

// Realtime Integration Tests
func TestCollection_Subscribe_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	client := pocketbase.NewClient(defaultURL)
	defaultBody := map[string]interface{}{
		"field": "value_" + time.Now().Format(time.StampMilli),
	}
	collection := pocketbase.CollectionSet[map[string]any](client, migrations.PostsPublic)
	stream, err := collection.Subscribe()
	if err != nil {
		t.Error(err)
		return
	}
	defer stream.Unsubscribe()
	<-stream.Ready()

	ch := stream.Events()

	t.Run("subscribe event: create", func(t *testing.T) {
		resp, err := collection.Create(defaultBody)
		if err != nil {
			t.Error(err)
			return
		}
		e := <-ch
		assert.Equal(t, "create", e.Action)
		assert.Equal(t, resp.ID, e.Record["id"])
	})

	t.Run("subscribe event: update", func(t *testing.T) {
		resp, err := collection.Create(defaultBody)
		if err != nil {
			t.Error(err)
			return
		}
		<-ch // ignore create event
		body := map[string]interface{}{
			"field": "value_" + time.Now().Format(time.StampMilli),
		}
		err = collection.Update(resp.ID, body)
		if err != nil {
			t.Error(err)
			return
		}
		e := <-ch
		assert.Equal(t, "update", e.Action)
		assert.Equal(t, body["field"], e.Record["field"])
	})

	t.Run("subscribe event: delete", func(t *testing.T) {
		resp, err := collection.Create(defaultBody)
		if err != nil {
			t.Error(err)
			return
		}
		<-ch // ignore create event
		err = collection.Delete(resp.ID)
		if err != nil {
			t.Error(err)
			return
		}
		e := <-ch
		assert.Equal(t, "delete", e.Action)
		assert.Equal(t, resp.ID, e.Record["id"])
	})
}

func TestCollection_Unsubscribe_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	client := pocketbase.NewClient(defaultURL)
	defaultBody := map[string]interface{}{
		"field": "value_" + time.Now().Format(time.StampMilli),
	}
	collection := pocketbase.CollectionSet[map[string]any](client, migrations.PostsPublic)
	stream, err := collection.Subscribe()
	if err != nil {
		t.Error(err)
		return
	}
	<-stream.Ready()

	ch := stream.Events()

	resp, err := collection.Create(defaultBody)
	if err != nil {
		t.Error(err)
		return
	}
	e := <-ch
	assert.Equal(t, resp.ID, e.Record["id"])

	stream.Unsubscribe()

	if err := collection.Delete(resp.ID); err != nil {
		t.Error(err)
		return
	}

	if _, ok := <-ch; ok {
		t.Error("unsubscribe is not working.")
	}
}

func TestCollection_RealtimeReconnect_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping realtime reconnect in short mode")
		return
	}

	client := pocketbase.NewClient(defaultURL)
	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			conn, err := net.Dial(network, addr)
			if err == nil {
				// Simulate pocketbase closing realtime connection after 5m of inactivity
				time.AfterFunc(3*time.Second, func() {
					// Close connection silently - don't log errors after test completion
					conn.Close()
				})
			}
			return conn, err
		},
	}
	client.GetClient().SetTransport(transport)
	defaultBody := map[string]interface{}{
		"field": "value_" + time.Now().Format(time.StampMilli),
	}
	collection := pocketbase.CollectionSet[map[string]any](client, migrations.PostsPublic)
	stream, err := collection.Subscribe()
	if err != nil {
		t.Error(err)
		return
	}
	defer stream.Unsubscribe()
	<-stream.Ready()

	// Use atomic operations and channels for thread-safe communication
	var eventReceived int32
	done := make(chan struct{})
	
	go func() {
		defer close(done)
		for range stream.Events() {
			if atomic.CompareAndSwapInt32(&eventReceived, 0, 1) {
				return
			}
		}
	}()
	
	time.AfterFunc(13*time.Second, func() {
		if _, err := collection.Create(defaultBody); err != nil {
			t.Error(err)
		}
	})
	
	// Wait for either success or timeout
	select {
	case <-done:
		assert.Equal(t, int32(1), atomic.LoadInt32(&eventReceived))
	case <-time.After(16 * time.Second):
		t.Fatal("stream reconnect test timed out")
	}
}

// Collections Integration Tests
func TestCollection_List_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	defaultClient := pocketbase.NewClient(defaultURL)

	tests := []struct {
		name       string
		client     *pocketbase.Client
		collection string
		params     pocketbase.ParamsList
		wantResult bool
		wantErr    bool
	}{
		{
			name:       "List with no params",
			client:     defaultClient,
			collection: migrations.PostsPublic,
			wantErr:    false,
			wantResult: true,
		},
		{
			name:       "List no results - query",
			client:     defaultClient,
			collection: migrations.PostsPublic,
			params: pocketbase.ParamsList{
				Filters: "field='some_random_value'",
			},
			wantErr:    false,
			wantResult: false,
		},
		{
			name:       "List no results - invalid query",
			client:     defaultClient,
			collection: migrations.PostsPublic,
			params: pocketbase.ParamsList{
				Filters: "field~~~some_random_value'",
			},
			wantErr:    true,
			wantResult: false,
		},
		{
			name:       "List invalid collection",
			client:     defaultClient,
			collection: "invalid_collection",
			wantErr:    true,
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection := pocketbase.CollectionSet[map[string]any](tt.client, tt.collection)
			got, err := collection.List(tt.params)
			assert.Equal(t, tt.wantErr, err != nil, err)
			assert.Equal(t, tt.wantResult, got.TotalItems > 0)
		})
	}
}

func TestCollection_Delete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	client := pocketbase.NewClient(defaultURL)
	field := "value_" + time.Now().Format(time.StampMilli)
	collection := pocketbase.CollectionSet[map[string]any](client, migrations.PostsPublic)

	// delete non-existing item
	err := collection.Delete("non_existing_id")
	assert.Error(t, err)

	// create temporary item
	resultCreated, err := collection.Create(map[string]any{
		"field": field,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, resultCreated.ID)

	// confirm item exists
	resultList, err := collection.List(pocketbase.ParamsList{Filters: "id='" + resultCreated.ID + "'"})
	assert.NoError(t, err)
	assert.Len(t, resultList.Items, 1)

	// delete temporary item
	err = collection.Delete(resultCreated.ID)
	assert.NoError(t, err)

	// confirm item does not exist
	resultList, err = collection.List(pocketbase.ParamsList{Filters: "id='" + resultCreated.ID + "'"})
	assert.NoError(t, err)
	assert.Len(t, resultList.Items, 0)
}

func TestCollection_Update_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	client := pocketbase.NewClient(defaultURL)
	field := "value_" + time.Now().Format(time.StampMilli)
	collection := pocketbase.CollectionSet[map[string]any](client, migrations.PostsPublic)

	// update non-existing item
	err := collection.Update("non_existing_id", map[string]any{
		"field": field,
	})
	assert.Error(t, err)

	// create temporary item
	resultCreated, err := collection.Create(map[string]any{
		"field": field,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, resultCreated.ID)

	// confirm item exists
	resultList, err := collection.List(pocketbase.ParamsList{Filters: "id='" + resultCreated.ID + "'"})
	assert.NoError(t, err)
	require.Len(t, resultList.Items, 1)
	assert.Equal(t, field, resultList.Items[0]["field"])

	// update temporary item
	err = collection.Update(resultCreated.ID, map[string]any{
		"field": field + "_updated",
	})
	assert.NoError(t, err)

	// confirm changes
	resultList, err = collection.List(pocketbase.ParamsList{Filters: "id='" + resultCreated.ID + "'"})
	assert.NoError(t, err)
	require.Len(t, resultList.Items, 1)
	assert.Equal(t, field+"_updated", resultList.Items[0]["field"])
}

func TestCollection_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	defaultClient := pocketbase.NewClient(defaultURL)
	defaultBody := map[string]interface{}{
		"field": "value_" + time.Now().Format(time.StampMilli),
	}

	tests := []struct {
		name       string
		client     *pocketbase.Client
		collection string
		body       any
		wantErr    bool
		wantID     bool
	}{
		{
			name:       "Create with no body",
			client:     defaultClient,
			collection: migrations.PostsPublic,
			wantErr:    false,
			wantID:     true,
		},
		{
			name:       "Create with body",
			client:     defaultClient,
			collection: migrations.PostsPublic,
			body:       defaultBody,
			wantErr:    false,
			wantID:     true,
		},
		{
			name:       "Create invalid collections",
			client:     defaultClient,
			collection: "invalid_collection",
			body:       defaultBody,
			wantErr:    true,
		},
		{
			name:       "Create no auth",
			client:     defaultClient,
			collection: migrations.PostsUser,
			body:       defaultBody,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection := pocketbase.CollectionSet[any](tt.client, tt.collection)
			r, err := collection.Create(tt.body)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantID {
				assert.NotEmpty(t, r.ID)
			} else {
				assert.Empty(t, r.ID)
			}
		})
	}
}

func TestCollection_One_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	client := pocketbase.NewClient(defaultURL)
	field := "value_" + time.Now().Format(time.StampMilli)
	collection := pocketbase.CollectionSet[map[string]any](client, migrations.PostsPublic)

	// update non-existing item
	_, err := collection.One("non_existing_id")
	assert.Error(t, err)

	// create temporary item
	resultCreated, err := collection.Create(map[string]any{
		"field": field,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, resultCreated.ID)

	// confirm item exists
	item, err := collection.One(resultCreated.ID)
	assert.NoError(t, err)
	assert.Equal(t, field, item["field"])

	// update temporary item
	err = collection.Update(resultCreated.ID, map[string]any{
		"field": field + "_updated",
	})
	assert.NoError(t, err)

	// confirm changes
	item, err = collection.One(resultCreated.ID)
	assert.NoError(t, err)
	assert.Equal(t, field+"_updated", item["field"])
}

// Admin Backup Integration Tests
func TestBackup_FullList_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("without authorization", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)
		resp, err := defaultClient.Backup().FullList()
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})

	t.Run("with valid authentication, but no existing backups", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		resp, err := defaultClient.Backup().FullList()
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Empty(t, resp)
	})

	t.Run("with valid authentication, create backup and check", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		err := defaultClient.Backup().Create()
		require.NoError(t, err)

		resp, err := defaultClient.Backup().FullList()
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp)

		// cleanup
		_ = defaultClient.Backup().Delete(resp[0].Key)
	})
}

func TestBackup_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("without authorization", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)
		err := defaultClient.Backup().Create()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})

	t.Run("with valid authentication, create backup without name and check", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		resp, err := defaultClient.Backup().FullList()
		require.NoError(t, err)
		require.Empty(t, resp)

		err = defaultClient.Backup().Create()
		assert.NoError(t, err)

		resp, err = defaultClient.Backup().FullList()
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp)

		// cleanup
		_ = defaultClient.Backup().Delete(resp[0].Key)
	})

	t.Run("with valid authentication, create backup with name and check", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		const backupName = "foobar"
		err := defaultClient.Backup().Create(backupName)
		assert.NoError(t, err)

		assert.True(t, isBackupExisting(t, defaultClient, backupName+".zip"))

		// cleanup
		_ = defaultClient.Backup().Delete(backupName + ".zip")
	})

	t.Run("with valid authentication, create backup with name incl. zip-extension and check", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		const backupName = "barfoo.zip"
		err := defaultClient.Backup().Create(backupName)
		assert.NoError(t, err)

		assert.True(t, isBackupExisting(t, defaultClient, backupName))

		// cleanup
		_ = defaultClient.Backup().Delete(backupName)
	})
}

func TestBackup_Delete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("without authorization", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)
		err := defaultClient.Backup().Delete("foobar")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})

	t.Run("create a backup and delete it", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		backupName := "foobar.zip"

		err := defaultClient.Backup().Create(backupName)
		require.NoError(t, err)
		require.True(t, isBackupExisting(t, defaultClient, backupName))

		err = defaultClient.Backup().Delete(backupName)
		assert.NoError(t, err)
		assert.False(t, isBackupExisting(t, defaultClient, backupName))
	})

	t.Run("create a backup and delete a non existing", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		backupName := "foobar.zip"

		err := defaultClient.Backup().Create(backupName)
		require.NoError(t, err)
		require.True(t, isBackupExisting(t, defaultClient, backupName))

		err = defaultClient.Backup().Delete(backupName + backupName)
		assert.Error(t, err)
		assert.False(t, isBackupExisting(t, defaultClient, backupName+backupName))

		// cleanup
		require.NoError(t, defaultClient.Backup().Delete(backupName))
	})
}

func TestBackup_Restore_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("without authorization", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL)
		err := defaultClient.Backup().Restore("foobar")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "valid record authorization")
	})

	t.Run("restore a backup", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		backupName := "foobar.zip"

		err := defaultClient.Backup().Create(backupName)
		require.NoError(t, err)
		require.True(t, isBackupExisting(t, defaultClient, backupName))

		err = defaultClient.Backup().Restore(backupName)
		assert.NoError(t, err)

		// cleanup
		require.NoError(t, defaultClient.Backup().Delete(backupName))
	})

	t.Run("cannot restore a backup with a non existing key", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		backupName := "not_existing.zip"

		err := defaultClient.Backup().Restore(backupName)
		assert.Error(t, err)
	})
}

func TestBackup_GetDownloadURL_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Run("build URL with token and key", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		url, err := defaultClient.Backup().GetDownloadURL("token", "key")
		assert.NoError(t, err)
		assert.Equal(t, defaultURL+"/api/backups/key?token=token", url)
	})

	t.Run("build URL with no token and no key", func(t *testing.T) {
		defaultClient := pocketbase.NewClient(defaultURL, pocketbase.WithAdminEmailPassword(migrations.AdminEmailPassword, migrations.AdminEmailPassword))
		url, err := defaultClient.Backup().GetDownloadURL("", "")
		assert.Error(t, err)
		assert.Empty(t, url)
	})
}

func isBackupExisting(t *testing.T, defaultClient *pocketbase.Client, backupname string) bool {
	resp, err := defaultClient.Backup().FullList()
	require.NoError(t, err)

	for _, b := range resp {
		if b.Key == backupname {
			return true
		}
	}
	return false
}