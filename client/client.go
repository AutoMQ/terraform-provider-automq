package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const HostURL string = ""

type Client struct {
	HostURL     string
	HTTPClient  *http.Client
	Token       string
	Credentials AuthCredentials
}

type AuthCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
}

type ErrorResponse struct {
	Code         int      `json:"code"`
	ErrorMessage string   `json:"error_message"`
	APIError     APIError `json:"api_error"`
	Err          error    `json:"error"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Errorf("code: %d, message: %s, details: %v", e.Code, e.ErrorMessage, e.Err).Error()
}

func NewClient(ctx context.Context, host, token *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		// Default Hashicups URL
		HostURL: HostURL,
	}

	if host != nil {
		c.HostURL = *host
	}

	if token == nil {
		return &c, nil
	}

	c.Token = *token

	err := c.checkAuth()
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Client) doRequest(req *http.Request, authToken *string) ([]byte, error) {
	token := c.Token

	if authToken != nil {
		token = *authToken
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en")

	// signer := signer.NewSigner(signer.Credentials{AccessKeyID: c.Credentials.AccessKeyID, SecretAccessKey: c.Credentials.SecretAccessKey})
	// var seeker io.ReadSeeker
	// if sr, ok := req.Body.(io.ReadSeeker); ok {
	// 	seeker = sr
	// }
	// signer.Sign(req, seeker, "cmp", "private", time.Now())

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &ErrorResponse{Code: 0, ErrorMessage: "Error sending request", Err: err}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, &ErrorResponse{Code: res.StatusCode, ErrorMessage: "Error reading response body", Err: err}
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		apiError := APIError{}
		err = json.Unmarshal(body, &apiError)
		if err != nil {
			return nil, &ErrorResponse{Code: res.StatusCode, ErrorMessage: "Error unmarshaling response body", Err: err}
		}
		return nil, &ErrorResponse{Code: res.StatusCode, APIError: apiError, Err: nil}
	}

	return body, nil
}
