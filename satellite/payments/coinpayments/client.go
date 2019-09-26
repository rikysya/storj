package coinpayments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type Credentials struct {
	PublicKey  string
	PrivateKey string
}

type Client struct {
	creds Credentials
	http  http.Client
}

func NewClient(creds Credentials) *Client {
	client := &Client{
		creds: creds,
		http: http.Client{
			Timeout: 0,
		},
	}
	return client
}

func (c *Client) Transactions() Transactions {
	return Transactions{client: c}
}

func (c *Client) hmac(payload []byte) string {
	mac := hmac.New(sha512.New, []byte(c.creds.PrivateKey))
	mac.Write(payload)
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func (c *Client) do(ctx context.Context, cmd string, values url.Values) (json.RawMessage, error) {
	values.Set("version", "1")
	values.Set("format", "json")
	values.Set("key", c.creds.PublicKey)
	values.Set("cmd", cmd)

	encoded := values.Encode()

	buff := bytes.NewBufferString(encoded)

	req, err := http.NewRequest(http.MethodPost, "https://www.coinpayments.net/api.php", buff)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HMAC", c.hmac([]byte(encoded)))

	resp, err := c.http.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("internal server error")
	}

	var data struct {
		Error  string          `json:"error"`
		Result json.RawMessage `json:"result"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.Error != "ok" {
		return nil, errors.New(data.Error)
	}

	return data.Result, nil
}
