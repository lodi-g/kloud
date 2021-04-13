package nextcloud

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
)

// Client is a NextCloud client for Kloud, it wraps http.Client
type Client struct {
	http    http.Client
	server  string
	shareID string
}

// NewClient creates a new NextCloud client with the configured TLS settings
func NewClient(tlsFilePath, server, shareID string) (Client, error) {
	caCert, err := os.ReadFile(tlsFilePath)
	if err != nil {
		return Client{}, err
	}

	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
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
