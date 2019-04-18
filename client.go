package pirategopher

import (
	"net/http"
	"time"
)

type Client struct {
	PublicKey []byte
	HttpClient *http.Client
}

func NewClient(timeout time.Duration) *Client {
	return &Client {
		HttpClient: &http.Client{
			Timeout: timeout,
		},
	}
}
