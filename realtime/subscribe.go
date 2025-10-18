package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/SierraSoftworks/multicast/v2"
	"github.com/cenkalti/backoff/v4"
	"github.com/donovanhide/eventsource"
	"github.com/go-resty/resty/v2"
)

// Event represents a real-time event from PocketBase with action, record data, and optional error.
type Event[T any] struct {
	Action string `json:"action"`
	Record T      `json:"record"`
	Error  error  `json:"-"`
}

// Authorizer defines the interface for authentication operations.
type Authorizer interface {
	Authorize() error
}

// HTTPClient defines the interface for HTTP client operations.
type HTTPClient interface {
	R() *resty.Request
}

// Collection defines the interface for collection operations needed by realtime subscriptions.
type Collection[T any] interface {
	GetName() string
	GetURL() string
	GetClient() HTTPClient
	GetAuthorizer() Authorizer
	IsSSEDebugEnabled() bool
}

// Subscribe creates a real-time subscription to the collection with default options.
func Subscribe[T any](c Collection[T], targets ...string) (*Stream[T], error) {
	opts := SubscribeOptions{
		ReconnectStrategy: &backoff.ZeroBackOff{},
	}
	return SubscribeWith[T](c, opts, targets...)
}

// SubscribeOptions configures real-time subscription behavior including reconnection strategy.
type SubscribeOptions struct {
	ReconnectStrategy backoff.BackOff
}

// SubscribeWith creates a real-time subscription with custom options and target collections.
func SubscribeWith[T any](c Collection[T], opts SubscribeOptions, targets ...string) (*Stream[T], error) {
	if err := c.GetAuthorizer().Authorize(); err != nil {
		return nil, err
	}

	if len(targets) == 0 {
		targets = []string{c.GetName()}
	}

	stream := newStream[T]()
	ctx, cancel := context.WithCancel(context.Background())
	stream.unsubscribe = func() { cancel() }

	handleSSEEvent := func(ev eventsource.Event) {
		var e Event[T]
		if c.IsSSEDebugEnabled() {
			log.Printf("SSE event: %+v", ev)
		}
		e.Error = json.Unmarshal([]byte(ev.Data()), &e)
		stream.channel.C <- e
	}

	once := &sync.Once{}
	stream.ready.Lock()
	startStream := func(check bool) func() error {
		return func() (err error) {
			req := c.GetClient().R().SetContext(ctx).SetDoNotParseResponse(true)
			resp, err := req.Get(c.GetURL() + "/api/realtime")
			if err != nil {
				return
			}
			defer func() {
				if closeErr := resp.RawBody().Close(); closeErr != nil {
					if c.IsSSEDebugEnabled() {
						log.Printf("Failed to close response body: %v", closeErr)
					}
				}
			}()

			d := eventsource.NewDecoder(resp.RawBody())

			ev, err := d.Decode()
			if err != nil {
				return err
			}
			if event := ev.Event(); event != "PB_CONNECT" {
				return fmt.Errorf("first event must be PB_CONNECT, but got %s", event)
			}

			if err := authSubscribeStream[T](c, []byte(ev.Data()), targets); err != nil {
				return err
			}

			if !check {
				once.Do(func() {
					stream.ready.Unlock()
				})
				for {
					ev, err := d.Decode()
					if err != nil {
						return err
					}
					go handleSSEEvent(ev)
				}
			}

			return nil
		}
	}

	if err := startStream(true)(); err != nil {
		return nil, err
	}

	go func() {
		if err := backoff.Retry(startStream(false), backoff.WithContext(opts.ReconnectStrategy, ctx)); err != nil {
			log.Print(err)
		}
	}()

	return stream, nil
}

// SubscriptionsSet represents the subscription configuration sent to PocketBase.
type SubscriptionsSet struct {
	ClientID      string   `json:"clientId"`
	Subscriptions []string `json:"subscriptions"`
}

func authSubscribeStream[T any](c Collection[T], data []byte, targets []string) (err error) {
	var s SubscriptionsSet
	if err = json.Unmarshal(data, &s); err != nil {
		return
	}
	s.Subscriptions = targets
	resp, err := c.GetClient().R().SetBody(s).Post(c.GetURL() + "/api/realtime")
	if err != nil {
		return
	}
	if code := resp.StatusCode(); code != http.StatusNoContent {
		return fmt.Errorf("auth subscribe stream failed. resp status code is %v", code)
	}
	return
}

// Stream represents a real-time event stream with subscription management capabilities.
type Stream[T any] struct {
	channel     *multicast.Channel[Event[T]]
	unsubscribe func()

	ready       *sync.RWMutex
	onceCleanup *sync.Once
}

func newStream[T any]() *Stream[T] {
	return &Stream[T]{
		channel:     multicast.New[Event[T]](),
		ready:       &sync.RWMutex{},
		onceCleanup: &sync.Once{},
	}
}

// Events returns a channel that receives real-time events from the stream.
func (s *Stream[T]) Events() <-chan Event[T] {
	return s.channel.Listen().C
}

// Unsubscribe closes the stream and cleans up resources.
func (s *Stream[T]) Unsubscribe() {
	s.onceCleanup.Do(func() {
		s.unsubscribe()
		s.channel.Close()
	})
}

// WaitAuthReady waits for the stream to be ready for authentication.
// Deprecated: use <-stream.Ready() instead.
func (s *Stream[T]) WaitAuthReady() error {
	s.ready.RLock()
	defer s.ready.RUnlock()
	return nil
}

// Ready returns a channel that closes when the stream is ready to receive events.
func (s *Stream[T]) Ready() <-chan struct{} {
	readyCh := make(chan struct{})
	go func() {
		s.ready.RLock()
		defer s.ready.RUnlock()
		close(readyCh)
	}()
	return readyCh
}

// CollectionSubscriber provides subscription methods for collections that implement the Collection interface.
type CollectionSubscriber[T any] struct {
	collection Collection[T]
}

// NewCollectionSubscriber creates a new subscriber for the given collection.
func NewCollectionSubscriber[T any](c Collection[T]) *CollectionSubscriber[T] {
	return &CollectionSubscriber[T]{collection: c}
}

// Subscribe creates a real-time subscription to the collection with default options.
func (cs *CollectionSubscriber[T]) Subscribe(targets ...string) (*Stream[T], error) {
	return Subscribe[T](cs.collection, targets...)
}

// SubscribeWith creates a real-time subscription with custom options and target collections.
func (cs *CollectionSubscriber[T]) SubscribeWith(opts SubscribeOptions, targets ...string) (*Stream[T], error) {
	return SubscribeWith[T](cs.collection, opts, targets...)
}