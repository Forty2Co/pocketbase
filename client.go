// Package pocketbase provides a Go SDK for interacting with PocketBase APIs.
//
// This package offers type-safe, idiomatic Go interfaces for PocketBase operations
// including authentication, CRUD operations, real-time subscriptions, and backup management.
//
// Example usage:
//
//	client := pocketbase.NewClient("http://localhost:8090")
//	records, err := client.List("posts", pocketbase.ParamsList{})
//
// For type-safe operations, use CollectionSet:
//
//	collection := pocketbase.CollectionSet[MyStruct](client, "posts")
//	records, err := collection.List(pocketbase.ParamsList{})
package pocketbase

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/go-resty/resty/v2"
	"github.com/pocketbase/pocketbase/core"
)

// ErrInvalidResponse is returned when PocketBase returns an invalid response.
var ErrInvalidResponse = errors.New("invalid response")

type (
	// Client represents a PocketBase API client with authentication and HTTP capabilities.
	Client struct {
		client     *resty.Client
		url        string
		authorizer authStore
		token      string
		sseDebug   bool
		restDebug  bool
	}
	// ClientOption is a function type for configuring Client instances.
	ClientOption func(*Client)
)

// EnvIsTruthy checks if an environment variable is set to a truthy value (1, true, yes).
func EnvIsTruthy(key string) bool {
	val := strings.ToLower(os.Getenv(key))
	return val == "1" || val == "true" || val == "yes"
}

// NewClient creates a new PocketBase API client with the specified URL and options.
func NewClient(url string, opts ...ClientOption) *Client {
	client := resty.New()
	client.
		SetRetryCount(3).
		SetRetryWaitTime(3 * time.Second).
		SetRetryMaxWaitTime(10 * time.Second)

	c := &Client{
		client:     client,
		url:        url,
		authorizer: authorizeNoOp{},
	}
	opts = append([]ClientOption{}, opts...)
	if EnvIsTruthy("REST_DEBUG") {
		opts = append(opts, WithRestDebug())
	}
	if EnvIsTruthy("SSE_DEBUG") {
		opts = append(opts, WithSseDebug())
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithRestDebug enables REST API debug logging for the client.
func WithRestDebug() ClientOption {
	return func(c *Client) {
		c.restDebug = true
		c.client.SetDebug(true)
	}
}

// WithSseDebug enables Server-Sent Events debug logging for the client.
func WithSseDebug() ClientOption {
	return func(c *Client) {
		c.sseDebug = true
	}
}

// WithAdminEmailPassword22 configures admin authentication using email and password (legacy version).
func WithAdminEmailPassword22(email, password string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeEmailPassword(c.client, c.url+"/api/admins/auth-with-password", email, password)
	}
}

// WithTimeout set the timeout for requests
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.client.SetTimeout(timeout)
	}
}

// WithRetry set the retry settings for requests (defaults: count=3, waitTime=3s, maxWaitTime=10s)
func WithRetry(count int, waitTime, maxWaitTime time.Duration) ClientOption {
	return func(c *Client) {
		c.client.SetRetryCount(count)
		c.client.SetRetryWaitTime(waitTime)
		c.client.SetRetryMaxWaitTime(maxWaitTime)
	}
}

// WithAdminEmailPassword configures admin authentication using email and password.
func WithAdminEmailPassword(email, password string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeEmailPassword(c.client, c.url+fmt.Sprintf("/api/collections/%s/auth-with-password", core.CollectionNameSuperusers), email, password)
	}
}

// WithUserEmailPassword configures user authentication using email and password.
func WithUserEmailPassword(email, password string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeEmailPassword(c.client, c.url+"/api/collections/users/auth-with-password", email, password)
	}
}

// WithUserEmailPasswordAndCollection configures user authentication for a specific collection.
func WithUserEmailPasswordAndCollection(email, password, collection string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeEmailPassword(c.client, c.url+"/api/collections/"+collection+"/auth-with-password", email, password)
	}
}

// WithAdminToken22 configures admin authentication using a token (legacy version).
func WithAdminToken22(token string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeToken(c.client, c.url+"/api/admins/auth-refresh", token)
	}
}

// WithAdminToken configures admin authentication using a token.
func WithAdminToken(token string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeToken(c.client, c.url+fmt.Sprintf("/api/collections/%s/auth-refresh", core.CollectionNameSuperusers), token)
	}
}

// WithUserToken configures user authentication using a token.
func WithUserToken(token string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeToken(c.client, c.url+"/api/collections/users/auth-refresh", token)
	}
}

// Authorize performs authentication using the configured authorization method.
func (c *Client) Authorize() error {
	return c.authorizer.authorize()
}

// Update updates a record in the specified collection.
func (c *Client) Update(collection string, id string, body any) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", collection).
		SetBody(body)

	resp, err := request.Patch(c.url + "/api/collections/{collection}/records/" + id)
	if err != nil {
		return fmt.Errorf("[update] can't send update request to pocketbase, err %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("[update] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	return nil
}

// Get performs a GET request to the specified path with optional request/response hooks.
func (c *Client) Get(path string, result any, onRequest func(*resty.Request), onResponse func(*resty.Response)) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json")
	if onRequest != nil {
		onRequest(request)
	}

	resp, err := request.Get(c.url + path)
	if err != nil {
		return fmt.Errorf("[get] can't send get request to pocketbase, err %w", err)
	}
	if onResponse != nil {
		onResponse(resp)
	}
	if resp.IsError() {
		return fmt.Errorf("[get] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), result); err != nil {
		return fmt.Errorf("[get] failed to unmarshal response: %w", err)
	}

	return nil
}

// Create creates a new record in the specified collection.
func (c *Client) Create(collection string, body any) (ResponseCreate, error) {
	var response ResponseCreate

	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", collection).
		SetBody(body).
		SetResult(&response)

	resp, err := request.Post(c.url + "/api/collections/{collection}/records")
	if err != nil {
		return response, fmt.Errorf("[create] can't send update request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[create] pocketbase returned status: %d, msg: %s, body: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			fmt.Sprintf("%+v", body), // TODO remove that after debugging
			ErrInvalidResponse,
		)
	}

	return *resp.Result().(*ResponseCreate), nil
}

// Delete removes a record from the specified collection.
func (c *Client) Delete(collection string, id string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", collection).
		SetPathParam("id", id)

	resp, err := request.Delete(c.url + "/api/collections/{collection}/records/{id}")
	if err != nil {
		return fmt.Errorf("[delete] can't send update request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[delete] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	return nil
}

// One retrieves a single record from the specified collection.
func (c *Client) One(collection string, id string) (map[string]any, error) {
	var response map[string]any

	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", collection).
		SetPathParam("id", id)

	resp, err := request.Get(c.url + "/api/collections/{collection}/records/{id}")
	if err != nil {
		return response, fmt.Errorf("[one] can't send get request to pocketbase, err %w", err)
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

// OneTo retrieves a single record and unmarshals it into the provided result.
func (c *Client) OneTo(collection string, id string, result any) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", collection).
		SetPathParam("id", id)

	resp, err := request.Get(c.url + "/api/collections/{collection}/records/{id}")
	if err != nil {
		return fmt.Errorf("[oneTo] can't send get request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[oneTo] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), result); err != nil {
		return fmt.Errorf("[oneTo] can't unmarshal response, err %w", err)
	}

	return nil
}

// List retrieves a paginated list of records from the specified collection.
func (c *Client) List(collection string, params ParamsList) (ResponseList[map[string]any], error) {
	var response ResponseList[map[string]any]

	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", collection)

	if params.Page > 0 {
		request.SetQueryParam("page", convertor.ToString(params.Page))
	}
	if params.Size > 0 {
		request.SetQueryParam("perPage", convertor.ToString(params.Size))
	}
	if params.Filters != "" {
		request.SetQueryParam("filter", params.Filters)
	}
	if params.Sort != "" {
		request.SetQueryParam("sort", params.Sort)
	}
	if params.Expand != "" {
		request.SetQueryParam("expand", params.Expand)
	}
	if params.Fields != "" {
		request.SetQueryParam("fields", params.Fields)
	}

	resp, err := request.Get(c.url + "/api/collections/{collection}/records")
	if err != nil {
		return response, fmt.Errorf("[list] can't send update request to pocketbase, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[list] pocketbase returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	var responseRef any = &response
	if params.hackResponseRef != nil {
		responseRef = params.hackResponseRef
	}
	if err := json.Unmarshal(resp.Body(), responseRef); err != nil {
		return response, fmt.Errorf("[list] can't unmarshal response, err %w", err)
	}
	return response, nil
}

// FullList retrieves all records from the specified collection without pagination.
func (c *Client) FullList(collection string, params ParamsList) (ResponseList[map[string]any], error) {
	var response ResponseList[map[string]any]
	params.Page = 1
	params.Size = 500

	if err := c.Authorize(); err != nil {
		return response, err
	}

	r, e := c.List(collection, params)
	if e != nil {
		return response, e
	}
	response.Items = append(response.Items, r.Items...)
	response.Page = r.Page
	response.PerPage = r.PerPage
	response.TotalItems = r.TotalItems
	response.TotalPages = r.TotalPages

	for i := 2; i <= r.TotalPages; i++ { // Start from page 2 because first page is already fetched
		params.Page = i
		r, e := c.List(collection, params)
		if e != nil {
			return response, e
		}
		response.Items = append(response.Items, r.Items...)
	}

	return response, nil
}

// AuthStore returns the client's authentication store.
func (c *Client) AuthStore() authStore {
	return c.authorizer
}

// Backup returns a Backup instance for managing backup operations.
func (c *Client) Backup() Backup {
	return Backup{
		Client: c,
	}
}

// Files returns a Files instance for managing file operations.
func (c *Client) Files() Files {
	return Files{
		Client: c,
	}
}
