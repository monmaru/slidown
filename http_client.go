package slidown

import (
	"fmt"
	"net"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine/socket"
	"google.golang.org/appengine/urlfetch"
)

// DefaultHTTPClient ...
func DefaultHTTPClient(ctx context.Context) *http.Client {
	return urlfetch.Client(ctx)
}

// CustomHTTPClient ...
func CustomHTTPClient(ctx context.Context) *http.Client {
	tr := &http.Transport{
		Dial: func(network, hostPort string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(hostPort)
			if err != nil {
				return nil, err
			}

			addrs, err := socket.LookupIP(ctx, host)
			if err != nil {
				return nil, err
			}

			firstIP := addrs[0]
			var conn *socket.Conn
			if firstIP.To4() != nil {
				conn, err = socket.Dial(ctx, network, fmt.Sprintf("%s:%s", addrs[0], port))
			} else {
				// brackets for IPv6 addrs
				conn, err = socket.Dial(ctx, network, fmt.Sprintf("[%s]:%s", addrs[0], port))
			}

			if err != nil {
				return nil, err
			}

			conn.SetContext(ctx)

			return conn, nil
		},
	}

	return &http.Client{Transport: tr}
}
