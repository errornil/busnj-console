package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Client responsible for forwarding requests to targetAddress
type Client struct {
	target *url.URL
	proxy  *httputil.ReverseProxy
}

// NewClient creates new Client
func NewClient(targetAddress string) (*Client, error) {
	target, err := url.Parse(targetAddress)
	if err != nil {
		return nil, err
	}

	return &Client{
		target: target,
		proxy:  httputil.NewSingleHostReverseProxy(target),
	}, nil
}

// Serve uses httputil.ReverseProxy to forward request
func (c *Client) Serve(w http.ResponseWriter, r *http.Request) {
	req := r
	// Update the headers to allow for SSL redirection
	req.URL.Host = c.target.Host
	req.URL.Scheme = c.target.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = c.target.Host

	// Note that ServeHttp is non blocking and uses a go routine under the hood
	c.proxy.ServeHTTP(w, req)
}
