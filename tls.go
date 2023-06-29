package rx

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewReverseProxyWithTLS(target *url.URL) *httputil.ReverseProxy {
	var proxy = httputil.NewSingleHostReverseProxy(target)
	var director = proxy.Director
	proxy.Transport = &http.Transport{DialTLSContext: dialTLS}
	proxy.Director = func(req *http.Request) {
		director(req)
		req.Host = req.URL.Host
	}
	return proxy
}

func dialTLS(ctx context.Context, network, addr string) (net.Conn, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	cfg := &tls.Config{ServerName: host}
	tlsConn := tls.Client(conn, cfg)
	if err = tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	cert := tlsConn.ConnectionState().PeerCertificates[0]
	if err = cert.VerifyHostname(host); err != nil {
		return nil, err
	}
	return tlsConn, nil
}
