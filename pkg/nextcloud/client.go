package nextcloud

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
)

// Errors returned by the NewClient function
var (
	ErrUnableToAppendCerts = errors.New("unable to append certificates to pool")
)

// Client is a NextCloud client for Kloud and wraps http.Client
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
		return Client{}, ErrUnableToAppendCerts
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
