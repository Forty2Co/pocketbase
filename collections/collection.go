// Package collections provides a Go interfaces for PocketBase collection operations.
//
// This package offers a clean abstraction layer for interacting with PocketBase collections,
// including CRUD operations, authentication, real-time subscriptions, and record management.
//
// The main entry point is the CollectionSet function, which creates a type-safe wrapper
// around a PocketBase collection:
//
//	client := pocketbase.NewClient("http://localhost:8090")
//	collection := collections.CollectionSet[MyStruct](client, "my_collection")
//
// Basic CRUD operations:
//
//	// Create a record
//	response, err := collection.Create(MyStruct{Name: "example"})
//
//	// Read a record
//	record, err := collection.One("record_id")
//
//	// Update a record
//	err := collection.Update("record_id", MyStruct{Name: "updated"})
//
//	// Delete a record
//	err := collection.Delete("record_id")
//
//	// List records with pagination
//	records, err := collection.List(collections.ParamsList{
//		Page: 1,
//		Size: 10,
//		Filters: "status='active'",
//		Sort: "-created",
//	})
//
// Authentication operations for auth collections:
//
//	// Authenticate with email/password
//	authResponse, err := collection.AuthWithPassword("user@example.com", "password")
//
//	// Refresh authentication token
//	refreshResponse, err := collection.AuthRefresh()
//
//	// Request password reset
//	err := collection.RequestPasswordReset("user@example.com")
//
// Real-time subscriptions:
//
//	// Subscribe to collection changes
//	stream, err := collection.Subscribe()
//	defer stream.Unsubscribe()
//
//	for event := range stream.Events() {
//		fmt.Printf("Action: %s, Record: %+v\n", event.Action, event.Record)
//	}
//
// The package maintains backward compatibility with the main pocketbase package
// while providing improved organization and type safety for collection-specific operations.
package collections

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Forty2Co/pocketbase/auth"
	"github.com/Forty2Co/pocketbase/realtime"
	"github.com/go-resty/resty/v2"
)

// ResponseList represents a paginated list response from PocketBase.
type ResponseList[T any] struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	TotalItems int `json:"totalItems"`
	TotalPages int `json:"totalPages"`
	Items      []T `json:"items"`
}

// ResponseCreate represents the response from creating a new record.
type ResponseCreate struct {
	ID      string `json:"id"`
	Created string `json:"created"`
	Field   string `json:"field"`
	Updated string `json:"updated"`
}

// ParamsList represents query parameters for PocketBase API requests including pagination, filtering, and sorting.
type ParamsList struct {
	Page    int
	Size    int
	Filters string
	Sort    string
	Expand  string
	Fields  string

	HackResponseRef any //hack for collection list
}

// ErrInvalidResponse is returned when PocketBase returns an invalid response.
var ErrInvalidResponse = auth.ErrInvalidResponse

// ClientInterface defines the interface that the collections package needs from the main client.
type ClientInterface interface {
	Update(collection string, id string, body any) error
	Create(collection string, body any) (ResponseCreate, error)
	Delete(collection string, id string) error
	List(collection string, params ParamsList) (ResponseList[map[string]any], error)
	FullList(collection string, params ParamsList) (ResponseList[map[string]any], error)
	Authorize() error
	SetToken(token string) // Add SetToken method to the interface
	GetToken() string      // Add GetToken method to the interface
}

// TokenUpdater defines the interface for updating the client's token
type TokenUpdater interface {
	SetToken(token string)
}

// Client represents the collections-specific client functionality.
type Client struct {
	client     *resty.Client
	url        string
	authorizer auth.Store
	sseDebug   bool
}

// NewClient creates a new collections client with the provided HTTP client, URL, and authorizer.
func NewClient(client *resty.Client, url string, authorizer auth.Store, sseDebug bool) *Client {
	return &Client{
		client:     client,
		url:        url,
		authorizer: authorizer,
		sseDebug:   sseDebug,
	}
}

// Collection represents a type-safe wrapper around a PocketBase collection.
type Collection[T any] struct {
	Client             ClientInterface
	Name               string
	BaseCollectionPath string
	client             *resty.Client
	url                string
	authorizer         auth.Store
	sseDebug           bool
	token              string
}

// CollectionSet creates a new type-safe collection wrapper for the specified collection.
func CollectionSet[T any](client ClientInterface, collection string) *Collection[T] {
	// We need to extract the underlying client details for direct API calls
	// This is a temporary solution until we refactor the client interface
	var restyClient *resty.Client
	var baseURL string
	var authorizer auth.Store
	var sseDebug bool
	
	// Type assertion to get the underlying client details
	// This assumes the client implements these methods
	if c, ok := client.(interface{ GetClient() *resty.Client }); ok {
		restyClient = c.GetClient()
	}
	if c, ok := client.(interface{ GetURL() string }); ok {
		baseURL = c.GetURL()
	}
	if c, ok := client.(interface{ AuthStore() auth.Store }); ok {
		authorizer = c.AuthStore()
	}
	if c, ok := client.(interface{ IsSSEDebugEnabled() bool }); ok {
		sseDebug = c.IsSSEDebugEnabled()
	}
	
	return &Collection[T]{
		Client:             client,
		Name:               collection,
		BaseCollectionPath: baseURL + "/api/collections/" + url.QueryEscape(collection),
		client:             restyClient,
		url:                baseURL,
		authorizer:         authorizer,
		sseDebug:           sseDebug,
	}
}

// Update updates a record in the collection with the specified ID.
func (c *Collection[T]) Update(id string, body T) error {
	return c.Client.Update(c.Name, id, body)
}

// Create creates a new record in the collection.
func (c *Collection[T]) Create(body T) (ResponseCreate, error) {
	return c.Client.Create(c.Name, body)
}

// Delete removes a record from the collection by ID.
func (c *Collection[T]) Delete(id string) error {
	return c.Client.Delete(c.Name, id)
}

// List retrieves a paginated list of records from the collection.
func (c *Collection[T]) List(params ParamsList) (ResponseList[T], error) {
	var response ResponseList[T]
	params.HackResponseRef = &response

	_, err := c.Client.List(c.Name, params)
	return response, err
}

// FullList retrieves all records from the collection without pagination.
func (c *Collection[T]) FullList(params ParamsList) (ResponseList[T], error) {
	var response ResponseList[T]
	params.HackResponseRef = &response

	_, err := c.Client.FullList(c.Name, params)
	return response, err
}

// Authorize performs authentication using the configured authorization method.
func (c *Collection[T]) Authorize() error {
	if c.authorizer != nil {
		return c.authorizer.Authorize()
	}
	return c.Client.Authorize()
}

// One retrieves a single record from the collection by ID.
func (c *Collection[T]) One(id string) (T, error) {
	var response T

	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", c.Name).
		SetPathParam("id", id)

	resp, err := request.Get(c.url + "/api/collections/{collection}/records/{id}")
	if err != nil {
		return response, fmt.Errorf("[one] can't send update request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[one] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[one] can't unmarshal response, err %w", err)
	}
	return response, nil
}

// OneWithParams retrieves a single record from the collection by ID with additional parameters.
// Only fields and expand parameters are supported.
func (c *Collection[T]) OneWithParams(id string, params ParamsList) (T, error) {
	var response T

	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", c.Name).
		SetPathParam("id", id).
		SetQueryParam("fields", params.Fields).
		SetQueryParam("expand", params.Expand)

	resp, err := request.Get(c.url + "/api/collections/{collection}/records/{id}")
	if err != nil {
		return response, fmt.Errorf("[one] can't send update request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[one] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[one] can't unmarshal response, err %w", err)
	}
	return response, nil
}
// GetName returns the collection name.
func (c *Collection[T]) GetName() string {
	return c.Name
}

// GetURL returns the base URL for the client.
func (c *Collection[T]) GetURL() string {
	return c.url
}

// GetClient returns the HTTP client.
func (c *Collection[T]) GetClient() realtime.HTTPClient {
	return c.client
}

// GetAuthorizer returns the authorizer for authentication.
func (c *Collection[T]) GetAuthorizer() realtime.Authorizer {
	return c.authorizer
}

// IsSSEDebugEnabled returns whether SSE debug is enabled.
func (c *Collection[T]) IsSSEDebugEnabled() bool {
	return c.sseDebug
}

// Subscribe creates a real-time subscription to the collection with default options.
func (c *Collection[T]) Subscribe(targets ...string) (*realtime.Stream[T], error) {
	return realtime.Subscribe[T](c, targets...)
}

// SubscribeWith creates a real-time subscription with custom options and target collections.
func (c *Collection[T]) SubscribeWith(opts realtime.SubscribeOptions, targets ...string) (*realtime.Stream[T], error) {
	return realtime.SubscribeWith[T](c, opts, targets...)
}

// Record-related types and functionality

type (
	// AuthMethod represents the available authentication methods for a collection.
	AuthMethod struct {
		AuthProviders    []AuthProvider `json:"authProviders"`
		UsernamePassword bool           `json:"usernamePassword"`
		EmailPassword    bool           `json:"emailPassword"`
		OnlyVerified     bool           `json:"onlyVerified"`
	}

	// AuthProvider represents an OAuth2 authentication provider configuration.
	AuthProvider struct {
		Name                string `json:"name"`
		DisplayName         string `json:"displayName"`
		State               string `json:"state"`
		AuthURL             string `json:"authUrl"`
		CodeVerifier        string `json:"codeVerifier"`
		CodeChallenge       string `json:"codeChallenge"`
		CodeChallengeMethod string `json:"codeChallengeMethod"`
	}

	// AuthWithPasswordResponse represents the response from password authentication.
	AuthWithPasswordResponse struct {
		Record Record `json:"record"`
		Token  string `json:"token"`
	}

	// Record represents a PocketBase record with common fields.
	Record struct {
		Avatar          string `json:"avatar"`
		CollectionID    string `json:"collectionId"`
		CollectionName  string `json:"collectionName"`
		Created         string `json:"created"`
		Email           string `json:"email"`
		EmailVisibility bool   `json:"emailVisibility"`
		ID              string `json:"id"`
		Name            string `json:"name"`
		Updated         string `json:"updated"`
		Username        string `json:"username"`
		Verified        bool   `json:"verified"`
	}

	// AuthWithOauth2Response represents the response from OAuth2 authentication.
	AuthWithOauth2Response struct {
		Token string `json:"token"`
	}

	// AuthRefreshResponse represents the response from authentication token refresh.
	AuthRefreshResponse struct {
		Record struct {
			Avatar          string `json:"avatar"`
			CollectionID    string `json:"collectionId"`
			CollectionName  string `json:"collectionName"`
			Created         string `json:"created"`
			Email           string `json:"email"`
			EmailVisibility bool   `json:"emailVisibility"`
			ID              string `json:"id"`
			Name            string `json:"name"`
			Updated         string `json:"updated"`
			Username        string `json:"username"`
			Verified        bool   `json:"verified"`
		} `json:"record"`
		Token string `json:"token"`
	}

	// ExternalAuthRequest represents an external authentication provider link.
	ExternalAuthRequest struct {
		ID           string `json:"id"`
		Created      string `json:"created"`
		Updated      string `json:"updated"`
		RecordID     string `json:"recordId"`
		CollectionID string `json:"collectionId"`
		Provider     string `json:"provider"`
		ProviderID   string `json:"providerId"`
	}
)

// Internal response types for auth methods
type (
	otpResponse struct {
		Enabled  bool  `json:"enabled"`
		Duration int64 `json:"duration"` // in seconds
	}

	mfaResponse struct {
		Enabled  bool  `json:"enabled"`
		Duration int64 `json:"duration"` // in seconds
	}

	passwordResponse struct {
		IdentityFields []string `json:"identityFields"`
		Enabled        bool     `json:"enabled"`
	}

	oauth2Response struct {
		Providers []providerInfo `json:"providers"`
		Enabled   bool           `json:"enabled"`
	}

	providerInfo struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		State       string `json:"state"`
		AuthURL     string `json:"authURL"`

		// Deprecated: use AuthURL instead
		//todo remove in future versions AuthUrl is a deprecated field because of wrong spelling (Url vs. URL), we need to keep it for backward compatibility with v0.22
		//nolint:all // ignore wrong spelling linter
		AuthUrl string `json:"authUrl"`

		// technically could be omitted if the provider doesn't support PKCE,
		// but to avoid breaking existing typed clients we'll return them as empty string
		CodeVerifier        string `json:"codeVerifier"`
		CodeChallenge       string `json:"codeChallenge"`
		CodeChallengeMethod string `json:"codeChallengeMethod"`
	}

	// AuthMethodsResponse represents the response structure for authentication methods.
	// Borrowed from https://github.com/pocketbase/pocketbase/blob/844f18cac379fc749493dc4dd73638caa89167a1/apis/record_auth_methods.go#L52
	AuthMethodsResponse struct {
		Password passwordResponse `json:"password"`
		OAuth2   oauth2Response   `json:"oauth2"`
		MFA      mfaResponse      `json:"mfa"`
		OTP      otpResponse      `json:"otp"`

		// legacy fields
		// @todo remove after dropping v0.22 support
		AuthProviders    []providerInfo `json:"authProviders"`
		UsernamePassword bool           `json:"usernamePassword"`
		EmailPassword    bool           `json:"emailPassword"`
	}
)

// setToken sets the authentication token for this collection instance
func (c *Collection[T]) setToken(token string) {
	c.token = token
}

// getToken gets the authentication token for this collection instance
func (c *Collection[T]) getToken() string {
	return c.token
}

// baseCrudPath returns the base CRUD path for this collection
func (c *Collection[T]) baseCrudPath() string {
	return c.BaseCollectionPath + "/records/"
}

// ListAuthMethods22 returns all available collection auth methods (legacy version).
func (c *Collection[T]) ListAuthMethods22() (AuthMethod, error) {
	var response AuthMethod
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json")

	resp, err := request.Get(c.BaseCollectionPath + "/auth-methods")
	if err != nil {
		return response, fmt.Errorf("[records] can't send ListAuthMethods request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal response, err %w", err)
	}
	return response, nil
}

// ListAuthMethods returns all available collection auth methods.
func (c *Collection[T]) ListAuthMethods() (AuthMethodsResponse, error) {
	var response AuthMethodsResponse
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json")

	resp, err := request.Get(c.BaseCollectionPath + "/auth-methods")
	if err != nil {
		return response, fmt.Errorf("[records] can't send ListAuthMethods request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal response, err %w", err)
	}
	return response, nil
}

// AuthWithPassword authenticate a single auth collection record via its username/email and password.
//
// On success, this method also automatically updates
// the client's AuthStore data and returns:
// - the authentication token via the AuthWithPasswordResponse
// - the authenticated record model
func (c *Collection[T]) AuthWithPassword(username string, password string) (AuthWithPasswordResponse, error) {
	var response AuthWithPasswordResponse
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"identity": username,
			"password": password,
		})

	resp, err := request.Post(c.BaseCollectionPath + "/auth-with-password")
	if err != nil {
		return response, fmt.Errorf("[records] can't send auth-with-password request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase returned status at auth-with-password: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal auth-with-password-response, err %w", err)
	}

	c.setToken(response.Token)
	// Also update the main client's token if it supports it
	if tokenUpdater, ok := c.Client.(TokenUpdater); ok {
		tokenUpdater.SetToken(response.Token)
	}
	return response, nil
}

// AuthWithOAuth2Code authenticate a single auth collection record with OAuth2 code.
//
// If you don't have an OAuth2 code you may also want to check `authWithOAuth2` method.
//
// On success, this method also automatically updates
// the client's AuthStore data and returns:
// - the authentication token via the model
// - the authenticated record model
// - the OAuth2 account data (eg. name, email, avatar, etc.)
func (c *Collection[T]) AuthWithOAuth2Code(provider string, code string, codeVerifier string, redirectURL string) (AuthWithOauth2Response, error) {
	var response AuthWithOauth2Response
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"provider":     provider,
			"code":         code,
			"codeVerifier": codeVerifier,
			"redirectUrl":  redirectURL,
			//"createData":   createData,
		})

	resp, err := request.Post(c.BaseCollectionPath + "/auth-with-oauth2")
	if err != nil {
		return response, fmt.Errorf("[records] can't send auth-with-oauth2 request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase returned status at auth-with-oauth2: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal auth-with-oauth2-response, err %w", err)
	}

	c.setToken(response.Token)
	// Also update the main client's token if it supports it
	if tokenUpdater, ok := c.Client.(TokenUpdater); ok {
		tokenUpdater.SetToken(response.Token)
	}
	return response, nil
}

// AuthRefresh refreshes the current authenticated record instance and
// * returns a new token and record data.
func (c *Collection[T]) AuthRefresh() (AuthRefreshResponse, error) {
	var response AuthRefreshResponse
	if err := c.Authorize(); err != nil {
		return response, err
	}

	// Get token from collection first, then fall back to client token
	token := c.getToken()
	if token == "" {
		token = c.Client.GetToken()
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token)

	resp, err := request.Post(c.BaseCollectionPath + "/auth-refresh")
	if err != nil {
		return response, fmt.Errorf("[records] can't send auth-refresh request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase returned status at auth-refresh: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal auth-refresh-response, err %w", err)
	}

	c.setToken(response.Token)
	// Also update the main client's token if it supports it
	if tokenUpdater, ok := c.Client.(TokenUpdater); ok {
		tokenUpdater.SetToken(response.Token)
	}
	return response, nil
}

// RequestVerification sends auth record verification email request.
func (c *Collection[T]) RequestVerification(email string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"email": email,
		})
	resp, err := request.Post(c.BaseCollectionPath + "/request-verification")
	if err != nil {
		return fmt.Errorf("[records] can't send request-verification request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at request-verification: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// ConfirmVerification confirms auth record email verification request.
//
// If the current `client.authStore.model` matches with the auth record from the token,
// then on success the `client.authStore.model.verified` will be updated to `true`.
func (c *Collection[T]) ConfirmVerification(verificationToken string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"token": verificationToken,
		})
	resp, err := request.Post(c.BaseCollectionPath + "/confirm-verification")
	if err != nil {
		return fmt.Errorf("[records] can't send confirm-verification request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at confirm-verification: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// RequestPasswordReset sends auth record password reset request
func (c *Collection[T]) RequestPasswordReset(email string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"email": email,
		})
	resp, err := request.Post(c.BaseCollectionPath + "/request-password-reset")
	if err != nil {
		return fmt.Errorf("[records] can't send request-password-reset request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at request-password-reset: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// ConfirmPasswordReset confirms auth record password reset request.
func (c *Collection[T]) ConfirmPasswordReset(passwordResetToken string, password string, passwordConfirm string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"token":           passwordResetToken,
			"password":        password,
			"passwordConfirm": passwordConfirm,
		})
	resp, err := request.Post(c.BaseCollectionPath + "/confirm-password-reset")
	if err != nil {
		return fmt.Errorf("[records] can't send confirm-password-reset request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at confirm-password-reset: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// RequestEmailChange sends an email change request to the authenticated record model.
func (c *Collection[T]) RequestEmailChange(newEmail string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"newEmail": newEmail,
		}).
		SetAuthToken(func() string {
			token := c.getToken()
			if token == "" {
				token = c.Client.GetToken()
			}
			return token
		}())

	resp, err := request.Post(c.BaseCollectionPath + "/request-email-change")
	if err != nil {
		return fmt.Errorf("[records] can't send request-email-change request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at request-email-change: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// ConfirmEmailChange confirms auth record's new email address.
func (c *Collection[T]) ConfirmEmailChange(emailChangeToken string, password string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetMultipartFormData(map[string]string{
			"token":    emailChangeToken,
			"password": password,
		}).
		SetAuthToken(func() string {
			token := c.getToken()
			if token == "" {
				token = c.Client.GetToken()
			}
			return token
		}())

	resp, err := request.Post(c.BaseCollectionPath + "/confirm-email-change")
	if err != nil {
		return fmt.Errorf("[records] can't send confirm-email-change request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at confirm-email-change: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}

// ListExternalAuths22 lists all linked external auth providers for the specified auth record.
func (c *Collection[T]) ListExternalAuths22(recordID string) ([]ExternalAuthRequest, error) {
	var response []ExternalAuthRequest
	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json")

	resp, err := request.Get(c.baseCrudPath() + url.QueryEscape(recordID) + "/external-auths")
	if err != nil {
		return response, fmt.Errorf("[records] can't send list external-auths request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[records] pocketbase request for list external-auths returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[records] can't unmarshal list external-auths response, err %w", err)
	}
	return response, nil
}

// UnlinkExternalAuth22 unlinks a single external auth provider from the specified auth record.
func (c *Collection[T]) UnlinkExternalAuth22(recordID string, provider string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json")

	resp, err := request.Delete(c.baseCrudPath() + url.QueryEscape(recordID) + "/external-auths/" + url.QueryEscape(provider))
	if err != nil {
		return fmt.Errorf("[records] can't send unlink-external-auth-request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[records] pocketbase returned status at unlink-external-auth-: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}
	return nil
}