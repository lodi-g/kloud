package nextcloud

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
)

// Client is a NextCloud client for Kloud, it wraps http.Client
type Client struct {
	http    http.Client
	server  string
	shareID string
}

// NewClient creates a new NextCloud client with the configured TLS settings
func NewClient(cacert []byte, server, shareID string) (Client, error) {
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(cacert)
	if ok == false {
		panic(ok)
	}

	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	return Client{http: httpClient, server: server, shareID: shareID}, nil
}
