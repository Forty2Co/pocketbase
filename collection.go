package pocketbase

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/Forty2Co/pocketbase/realtime"
)

// Collection represents a type-safe wrapper around a PocketBase collection.
type Collection[T any] struct {
	*Client
	Name               string
	BaseCollectionPath string
}

// CollectionSet creates a new type-safe collection wrapper for the specified collection.
func CollectionSet[T any](client *Client, collection string) *Collection[T] {
	return &Collection[T]{
		Client:             client,
		Name:               collection,
		BaseCollectionPath: client.url + "/api/collections/" + url.QueryEscape(collection),
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
	params.hackResponseRef = &response

	_, err := c.Client.List(c.Name, params)
	return response, err
}

// FullList retrieves all records from the collection without pagination.
func (c *Collection[T]) FullList(params ParamsList) (ResponseList[T], error) {
	var response ResponseList[T]
	params.hackResponseRef = &response

	_, err := c.Client.FullList(c.Name, params)
	return response, err
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