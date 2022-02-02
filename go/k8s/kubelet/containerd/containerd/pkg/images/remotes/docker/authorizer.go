package docker

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"

	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/images/remotes/docker/auth"
)

type Authorizer struct {
	credentials func(string) (string, string, error)

	client *http.Client
	header http.Header

	// indexed by host name
	handlers map[string]*authHandler
}

func NewAuthorizer() *Authorizer {
	if ao.client == nil {
		ao.client = http.DefaultClient
	}

	return &Authorizer{
		credentials:         ao.credentials,
		client:              ao.client,
		header:              ao.header,
		handlers:            make(map[string]*authHandler),
		onFetchRefreshToken: ao.onFetchRefreshToken,
	}
}

func (authorizer *Authorizer) Authorize(ctx context.Context, req *http.Request) error {
	// skip if there is no auth handler
	ah := authorizer.getAuthHandler(req.URL.Host)
	if ah == nil {
		return nil
	}

	auth, refreshToken, err := ah.authorize(ctx)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", auth)

	if refreshToken != "" {
		onFetchRefreshToken := authorizer.onFetchRefreshToken
		if onFetchRefreshToken != nil {
			onFetchRefreshToken(ctx, refreshToken, req)
		}
	}
	return nil
}

func (authorizer *Authorizer) getAuthHandler(host string) *authHandler {
	return authorizer.handlers[host]
}

// authHandler is used to handle auth request per registry server.
type authHandler struct {
	sync.Mutex

	header http.Header

	client *http.Client

	// only support basic and bearer schemes
	scheme auth.AuthenticationScheme

	// common contains common challenge answer
	common auth.TokenOptions

	// scopedTokens caches token indexed by scopes, which used in
	// bearer auth case
	//scopedTokens map[string]*authResult
}

func (ah *authHandler) authorize(ctx context.Context) (string, string, error) {
	switch ah.scheme {
	case auth.BasicAuth:
		return ah.doBasicAuth(ctx)
	case auth.BearerAuth:
		return ah.doBearerAuth(ctx)
	default:
		return "", "", fmt.Errorf("failed to find supported auth scheme: %s: not implemented", string(ah.scheme))
	}
}

func (ah *authHandler) doBasicAuth(ctx context.Context) (string, string, error) {
	username, secret := ah.common.Username, ah.common.Secret

	if username == "" || secret == "" {
		return "", "", fmt.Errorf("failed to handle basic auth because missing username or secret")
	}

	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(username+":"+secret))), "", nil
}

func (ah *authHandler) doBearerAuth(ctx context.Context) (token, refreshToken string, err error) {
	// copy common tokenOptions
	to := ah.common
	// Docs: https://docs.docker.com/registry/spec/auth/scope
	//scoped := strings.Join(to.Scopes, " ")

	// fetch token for the resource scope
	if to.Secret != "" {
		defer func() {
			if err != nil {
				err = fmt.Errorf("failed to fetch oauth token: %w", err)
			}
		}()
		resp, err := auth.FetchTokenWithOAuth(ctx, ah.client, ah.header, "containerd-client", to)
		if err != nil {
			return "", "", err
		}
		return resp.AccessToken, resp.RefreshToken, nil
	}
	// do request anonymously
	resp, err := auth.FetchToken(ctx, ah.client, ah.header, to)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch anonymous token: %w", err)
	}
	return resp.Token, resp.RefreshToken, nil
}
