package docker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"k8s-lx1036/k8s/kubelet/containerd/containerd/pkg/images/remotes/docker/reference"

	"golang.org/x/net/context/ctxhttp"
	"k8s.io/klog/v2"
)

type dockerBase struct {
	refspec    reference.Spec
	repository string
	hosts      []RegistryHost
	header     http.Header
}

func (r *dockerBase) filterHosts(caps HostCapabilities) (hosts []RegistryHost) {
	for _, host := range r.hosts {
		if host.Capabilities.Has(caps) {
			hosts = append(hosts, host)
		}
	}
	return
}

type request struct {
	method string
	path   string
	header http.Header
	host   RegistryHost
	body   func() (io.ReadCloser, error)
	size   int64
}

func (r *dockerBase) request(host RegistryHost, method string, ps ...string) *request {
	header := r.header.Clone()
	if header == nil {
		header = http.Header{}
	}

	for key, value := range host.Header {
		header[key] = append(header[key], value...)
	}
	parts := append([]string{"/", host.Path, r.repository}, ps...)
	p := path.Join(parts...)
	// Join strips trailing slash, re-add ending "/" if included
	if len(parts) > 0 && strings.HasSuffix(parts[len(parts)-1], "/") {
		p = p + "/"
	}
	return &request{
		method: method,
		path:   p,
		header: header,
		host:   host,
	}
}

func (r *request) addNamespace(ns string) (err error) {
	if !r.host.isProxy(ns) {
		return nil
	}
	var q url.Values
	// Parse query
	if i := strings.IndexByte(r.path, '?'); i > 0 {
		r.path = r.path[:i+1]
		q, err = url.ParseQuery(r.path[i+1:])
		if err != nil {
			return
		}
	} else {
		r.path = r.path + "?"
		q = url.Values{}
	}
	q.Add("ns", ns)

	r.path = r.path + q.Encode()

	return
}

func (r *request) doWithRetries(ctx context.Context, responses []*http.Response) (*http.Response, error) {
	resp, err := r.do(ctx)
	if err != nil {
		return nil, err
	}

	responses = append(responses, resp)
	retry, err := r.retryRequest(ctx, responses)
	if err != nil {
		resp.Body.Close()
		return nil, err
	}
	if retry {
		resp.Body.Close()
		return r.doWithRetries(ctx, responses)
	}
	return resp, err
}

func (r *request) do(ctx context.Context) (*http.Response, error) {
	u := r.host.Scheme + "://" + r.host.Host + r.path
	req, err := http.NewRequest(r.method, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header = http.Header{} // headers need to be copied to avoid concurrent map access
	for k, v := range r.header {
		req.Header[k] = v
	}
	if r.body != nil {
		body, err := r.body()
		if err != nil {
			return nil, err
		}
		req.Body = body
		req.GetBody = r.body
		if r.size > 0 {
			req.ContentLength = r.size
		}
	}

	klog.Infof(fmt.Sprintf("do request url:%s", u))
	if err := r.authorize(ctx, req); err != nil {
		return nil, fmt.Errorf("failed to authorize: %w", err)
	}

	var client = &http.Client{}
	if r.host.Client != nil {
		*client = *r.host.Client
	}
	if client.CheckRedirect == nil {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			if err := r.authorize(ctx, req); err != nil {
				return fmt.Errorf("failed to authorize redirect: %w", err)
			}
			return nil
		}
	}

	resp, err := ctxhttp.Do(ctx, client, req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	klog.Infof(fmt.Sprintf("fetch response received"))
	return resp, nil
}

func (r *request) authorize(ctx context.Context, req *http.Request) error {
	// Check if has header for host
	if err := r.host.Authorizer.Authorize(ctx, req); err != nil {
		return err
	}

	return nil
}

type Fetcher struct {
	*dockerBase
}
