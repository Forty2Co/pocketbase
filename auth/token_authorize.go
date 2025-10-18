package auth

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/singleflight"
)

// TokenAuth handles token-based authentication
type TokenAuth struct {
	client      *resty.Client
	url         string
	token       string
	tokenValid  time.Time
	tokenSingle singleflight.Group
}

// NewTokenAuth creates a new token-based authenticator
func NewTokenAuth(c *resty.Client, url string, token string) Store {
	c.SetHeader("Authorization", token)
	return &TokenAuth{
		client:      c,
		url:         url,
		token:       token,
		tokenSingle: singleflight.Group{},
	}
}

func (a *TokenAuth) Authorize() error {
	type authResponse struct {
		Token string `json:"token"`
	}
	_, err, _ := a.tokenSingle.Do("auth-refresh", func() (interface{}, error) {
		if time.Now().Before(a.tokenValid) {
			return nil, nil
		}
		resp, err := a.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Authorization", a.token).
			SetResult(&authResponse{}).
			Post(a.url)
		if err != nil {
			return nil, fmt.Errorf("[auth-refresh] can't send request to pocketbase %w", err)
		}
		if resp.IsError() {
			return nil, fmt.Errorf("[auth-refresh] pocketbase returned status: %d, msg: %s, err %w",
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

func (a *TokenAuth) IsValid() bool {
	return time.Now().Before(a.tokenValid)
}

func (a *TokenAuth) Token() string {
	return a.token
}