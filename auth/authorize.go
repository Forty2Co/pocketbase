package auth

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/singleflight"
)

// Store represents an authentication store that manages tokens and validation
type Store interface {
	Authorizer
	IsValid() bool
	Token() string
}

// Authorizer handles the authorization process
type Authorizer interface {
	Authorize() error
}

// NoOpAuth is a no-operation authenticator that always returns empty/invalid values
type NoOpAuth struct{}

func (a NoOpAuth) Authorize() error {
	return nil
}

func (a NoOpAuth) IsValid() bool {
	return false
}

func (a NoOpAuth) Token() string {
	return ""
}

// EmailPasswordAuth handles email/password authentication
type EmailPasswordAuth struct {
	email       string
	password    string
	token       string
	tokenValid  time.Time
	client      *resty.Client
	url         string
	tokenSingle singleflight.Group
}

// NewEmailPasswordAuth creates a new email/password authenticator
func NewEmailPasswordAuth(c *resty.Client, url string, email string, password string) Store {
	return &EmailPasswordAuth{
		client:      c,
		email:       email,
		password:    password,
		url:         url,
		tokenSingle: singleflight.Group{},
	}
}

func (a *EmailPasswordAuth) Authorize() error {
	type authResponse struct {
		Token string `json:"token"`
	}

	_, err, _ := a.tokenSingle.Do("auth", func() (interface{}, error) {
		if time.Now().Before(a.tokenValid) {
			return nil, nil
		}

		resp, err := a.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(map[string]interface{}{
				"identity": a.email,
				"password": a.password,
			}).
			SetResult(&authResponse{}).
			SetHeader("Authorization", "").
			Post(a.url)

		if err != nil {
			return nil, fmt.Errorf("[auth] can't send request to pocketbase %w", err)
		}

		if resp.IsError() {
			return nil, fmt.Errorf("[auth] pocketbase returned status: %d, msg: %s, err %w",
				resp.StatusCode(),
				resp.String(),
				ErrInvalidResponse,
			)
		}

		auth := *resp.Result().(*authResponse)
		a.token = auth.Token
		a.client.SetHeader("Authorization", auth.Token)
		a.tokenValid = time.Now().Add(60 * time.Minute)

		return nil, nil
	})
	return err
}

func (a *EmailPasswordAuth) IsValid() bool {
	return time.Now().Before(a.tokenValid)
}

func (a *EmailPasswordAuth) Token() string {
	return a.token
}