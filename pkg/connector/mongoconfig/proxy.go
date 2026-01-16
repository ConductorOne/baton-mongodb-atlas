package mongoconfig

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
)

type MongoProxy struct {
	Host string
	Port int
}

func (p *MongoProxy) Enabled() bool {
	return p.Host != "" && p.Port != 0
}

func (p *MongoProxy) Address() string {
	return fmt.Sprintf("%s:%d", p.Host, p.Port)
}

// Dialer returns a SOCKS5 proxy dialer that routes all connections including
// DNS resolution through the proxy.
func (p *MongoProxy) Dialer() (proxy.ContextDialer, error) {
	if p == nil || !p.Enabled() {
		return nil, nil
	}

	base := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	socksDialer, err := proxy.SOCKS5("tcp", p.Address(), nil, base)
	if err != nil {
		return nil, fmt.Errorf("failed to create socks5 dialer: %w", err)
	}

	if cd, ok := socksDialer.(proxy.ContextDialer); ok {
		return cd, nil
	}

	// Wrap non-context dialer
	return contextDialerFunc(func(ctx context.Context, network, address string) (net.Conn, error) {
		return socksDialer.Dial(network, address)
	}), nil
}

// HTTPTransport returns an http.Transport configured to use the SOCKS5 proxy
// for all connections. If the proxy is not enabled, returns nil.
func (p *MongoProxy) HTTPTransport() (*http.Transport, error) {
	if p == nil || !p.Enabled() {
		return nil, nil
	}

	dialer, err := p.Dialer()
	if err != nil {
		return nil, err
	}

	return &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}, nil
}

// HTTPClient returns an http.Client configured to use the SOCKS5 proxy.
// If the proxy is not enabled, returns http.DefaultClient.
func (p *MongoProxy) HTTPClient() (*http.Client, error) {
	transport, err := p.HTTPTransport()
	if err != nil {
		return nil, err
	}

	if transport == nil {
		return http.DefaultClient, nil
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}, nil
}

type contextDialerFunc func(ctx context.Context, network, address string) (net.Conn, error)

func (f contextDialerFunc) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return f(ctx, network, address)
}
