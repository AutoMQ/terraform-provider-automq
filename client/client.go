package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"terraform-provider-automq/client/signer"
	"time"
)

type Client struct {
	HostURL     string
	HTTPClient  *http.Client
	Token       string
	Credentials AuthCredentials
	Signer      *signer.Signer
}

type AuthCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
}

type ErrorResponse struct {
	Code         int      `json:"code"`
	ErrorMessage string   `json:"error_message"`
	APIError     APIError `json:"api_error"`
	Err          error
}

type APIError struct {
	ErrorModel ErrorModel `json:"error"`
}

type ErrorModel struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

func (e *ErrorResponse) Error() string {
	if e.APIError.ErrorModel.Code != "" {
		return fmt.Sprintf("Error %d: %s: %s", e.Code, e.APIError.ErrorModel.Code, e.APIError.ErrorModel.Message)
	}
	return fmt.Sprintf("Error %d: %s, detail: %s", e.Code, e.ErrorMessage, e.Err.Error())
}

func NewClient(ctx context.Context, host string, credentials AuthCredentials) (*Client, error) {
	c := &Client{
		HTTPClient:  &http.Client{Timeout: 0 * time.Second},
		HostURL:     host,
		Credentials: credentials,
		Signer: signer.NewSigner(signer.Credentials{
			AccessKeyID:     credentials.AccessKeyID,
			SecretAccessKey: credentials.SecretAccessKey,
		}),
	}
	if err := c.checkAuth(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en")
	var seeker io.ReadSeeker
	if sr, ok := req.Body.(io.ReadSeeker); ok {
		seeker = sr
	} else if rc, ok := req.Body.(io.Reader); ok {
		data, err := io.ReadAll(rc)
		if err != nil {
			return nil, &ErrorResponse{Code: 0, ErrorMessage: "Error reading request body", Err: err}
		}
		seeker = bytes.NewReader(data)
	}
	_, err := c.Signer.Sign(req, seeker, "cmp", "private", time.Now())
	if err != nil {
		return nil, &ErrorResponse{Code: 0, ErrorMessage: "Error signing request", Err: err}
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &ErrorResponse{Code: 0, ErrorMessage: "Error sending request", Err: err}
	}
	defer func() {
		closeErr := res.Body.Close()
		if closeErr != nil {
			fmt.Printf("Error closing response body: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, &ErrorResponse{Code: res.StatusCode, ErrorMessage: "Error reading response body", Err: err}
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		apiError := APIError{}
		err = json.Unmarshal(body, &apiError)
		if err != nil {
			return nil, &ErrorResponse{Code: res.StatusCode, ErrorMessage: string(body), Err: err}
		}
		return nil, &ErrorResponse{Code: res.StatusCode, APIError: apiError}
	}

	return body, nil
}
