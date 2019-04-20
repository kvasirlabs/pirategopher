package pirategopher

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	ServerUrl	string
	PublicKey 	[]byte
	HttpClient 	*http.Client
}

func newClient(serverUrl string, publicKey []byte) *Client {
	return &Client {
		ServerUrl: serverUrl,
		PublicKey: publicKey,
		HttpClient: &http.Client{},
	}
}

func (c *Client) doRequest(method, endpoint string, body io.Reader,
	headers map[string]string) (*http.Response, error) {
		req, err := http.NewRequest(method, c.ServerUrl + endpoint, body)
		if err != nil {
			return &http.Response{}, err
		}
		for k, header := range headers {
			req.Header.Set(k, header)
		}
		res, err := c.HttpClient.Do(req)
		if err != nil {
			return &http.Response{}, err
		}
		return res, nil
}

// SendEncryptedPayload send an encrypted payload to server
func (c *Client) sendEncryptedPayload(endpoint, payload string,
	customHeaders map[string]string) (*http.Response, error) {
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	for k, v := range customHeaders {
		headers[k] = v
	}

	ciphertext, err := encrypt(c.PublicKey, []byte(payload))
	if err != nil {
		return &http.Response{}, err
	}

	data := url.Values{}
	data.Add("payload", string(ciphertext))
	return c.doRequest("POST", endpoint, strings.NewReader(data.Encode()), headers)
}

// AddNewKeyPair persist a new keypair on server
func (c *Client) addNewKeyPair(id, encKey string) (*http.Response, error) {
	payload := fmt.Sprintf(`{"id": "%s", "enckey": "%s"}`, id, encKey)
	return c.sendEncryptedPayload("/api/keys/add", payload, map[string]string{})
}