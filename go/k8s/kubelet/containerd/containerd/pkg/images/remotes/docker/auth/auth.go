package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/context/ctxhttp"
)

// AuthenticationScheme defines scheme of the authentication method
type AuthenticationScheme byte

const (
	// BasicAuth is scheme for Basic HTTP Authentication RFC 7617
	BasicAuth AuthenticationScheme = 1 << iota
	// DigestAuth is scheme for HTTP Digest Access Authentication RFC 7616
	DigestAuth
	// BearerAuth is scheme for OAuth 2.0 Bearer Tokens RFC 6750
	BearerAuth
)

var (
	ErrNoToken = errors.New("authorization server did not include a token in the response")
)

// TokenOptions are options for requesting a token
type TokenOptions struct {
	Realm    string
	Service  string
	Scopes   []string
	Username string
	Secret   string

	// FetchRefreshToken enables fetching a refresh token (aka "identity token", "offline token") along with the bearer token.
	//
	// For HTTP GET mode (FetchToken), FetchRefreshToken sets `offline_token=true` in the request.
	// https://docs.docker.com/registry/spec/auth/token/#requesting-a-token
	//
	// For HTTP POST mode (FetchTokenWithOAuth), FetchRefreshToken sets `access_type=offline` in the request.
	// https://docs.docker.com/registry/spec/auth/oauth/#getting-a-token
	FetchRefreshToken bool
}

// OAuthTokenResponse is response from fetching token with a OAuth POST request
type OAuthTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	IssuedAt     time.Time `json:"issued_at"`
	Scope        string    `json:"scope"`
}

// FetchTokenWithOAuth fetches a token using a POST request
func FetchTokenWithOAuth(ctx context.Context, client *http.Client, headers http.Header, clientID string, to TokenOptions) (*OAuthTokenResponse, error) {
	form := url.Values{}
	if len(to.Scopes) > 0 {
		form.Set("scope", strings.Join(to.Scopes, " "))
	}
	form.Set("service", to.Service)
	form.Set("client_id", clientID)

	if to.Username == "" {
		form.Set("grant_type", "refresh_token")
		form.Set("refresh_token", to.Secret)
	} else {
		form.Set("grant_type", "password")
		form.Set("username", to.Username)
		form.Set("password", to.Secret)
	}
	if to.FetchRefreshToken {
		form.Set("access_type", "offline")
	}

	req, err := http.NewRequest("POST", to.Realm, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	for k, v := range headers {
		req.Header[k] = append(req.Header[k], v...)
	}
	if len(req.Header.Get("User-Agent")) == 0 {
		req.Header.Set("User-Agent", "containerd/v1.6.0")
	}

	resp, err := ctxhttp.Do(ctx, client, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)

	var tr OAuthTokenResponse
	if err = decoder.Decode(&tr); err != nil {
		return nil, fmt.Errorf("unable to decode token response: %w", err)
	}

	if tr.AccessToken == "" {
		return nil, ErrNoToken
	}

	return &tr, nil
}

// FetchTokenResponse is response from fetching token with GET request
type FetchTokenResponse struct {
	Token        string    `json:"token"`
	AccessToken  string    `json:"access_token"`
	ExpiresIn    int       `json:"expires_in"`
	IssuedAt     time.Time `json:"issued_at"`
	RefreshToken string    `json:"refresh_token"`
}

// FetchToken fetches a token using a GET request
func FetchToken(ctx context.Context, client *http.Client, headers http.Header, to TokenOptions) (*FetchTokenResponse, error) {
	req, err := http.NewRequest("GET", to.Realm, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header[k] = append(req.Header[k], v...)
	}
	if len(req.Header.Get("User-Agent")) == 0 {
		req.Header.Set("User-Agent", "containerd/v1.6.0")
	}

	reqParams := req.URL.Query()

	if to.Service != "" {
		reqParams.Add("service", to.Service)
	}

	for _, scope := range to.Scopes {
		reqParams.Add("scope", scope)
	}

	if to.Secret != "" {
		req.SetBasicAuth(to.Username, to.Secret)
	}

	if to.FetchRefreshToken {
		reqParams.Add("offline_token", "true")
	}

	req.URL.RawQuery = reqParams.Encode()

	resp, err := ctxhttp.Do(ctx, client, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)

	var tr FetchTokenResponse
	if err = decoder.Decode(&tr); err != nil {
		return nil, fmt.Errorf("unable to decode token response: %w", err)
	}

	// `access_token` is equivalent to `token` and if both are specified
	// the choice is undefined.  Canonicalize `access_token` by sticking
	// things in `token`.
	if tr.AccessToken != "" {
		tr.Token = tr.AccessToken
	}

	if tr.Token == "" {
		return nil, ErrNoToken
	}

	return &tr, nil
}
