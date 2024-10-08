package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

type EnvironmentID string

const EnvIdKey EnvironmentID = "environment_id"

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
	var errMsg strings.Builder
	errMsg.WriteString(fmt.Sprintf("Something went wrong. Error Code: %d\n", e.Code))

	if e.APIError.ErrorModel.Code != "" {
		errMsg.WriteString(fmt.Sprintf("API Error: %s\n", e.APIError.ErrorModel.Code))
		errMsg.WriteString(fmt.Sprintf("Message: %s\n", e.APIError.ErrorModel.Message))
	} else {
		errMsg.WriteString(fmt.Sprintf("Message: %s\n", e.ErrorMessage))
	}
	if e.APIError.ErrorModel.Detail != "" {
		errMsg.WriteString(fmt.Sprintf("Details: %s\n", e.APIError.ErrorModel.Detail))
	}
	if e.Err != nil {
		errMsg.WriteString(fmt.Sprintf("Error: %s\n", e.Err.Error()))
	}
	return errMsg.String()
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
	return c, nil
}

func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return c.doRequest(ctx, "POST", path, strings.NewReader(string(b)))
}

func (c *Client) Get(ctx context.Context, path string, queryParams map[string]string) ([]byte, error) {
	if queryParams != nil {
		path += "?" + buildQueryParams(queryParams)
	}
	return c.doRequest(ctx, "GET", path, nil)
}

func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, "DELETE", path, nil)
}

func (c *Client) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return c.doRequest(ctx, "PUT", path, strings.NewReader(string(b)))
}

func (c *Client) Patch(ctx context.Context, path string, body interface{}) ([]byte, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return c.doRequest(ctx, "PATCH", path, strings.NewReader(string(b)))
}

func buildQueryParams(queryParams map[string]string) string {
	query := ""
	for key, value := range queryParams {
		query += key + "=" + value + "&"
	}
	return query
}

func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, c.HostURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en")
	environmentID, ok := ctx.Value(EnvIdKey).(string)
	if ok {
		req.Header.Set("X-automq-environment-id", environmentID)
	} else {
		return nil, &ErrorResponse{Code: 0, ErrorMessage: "Error getting environment ID from context"}
	}

	var seeker io.ReadSeeker
	if sr, ok := body.(io.ReadSeeker); ok {
		seeker = sr
	} else if rc, ok := body.(io.ReadCloser); ok {
		data, err := io.ReadAll(rc)
		if err != nil {
			return nil, &ErrorResponse{Code: 0, ErrorMessage: "Error reading request body", Err: err}
		}
		seeker = bytes.NewReader(data)
	}
	_, err = c.Signer.Sign(req, seeker, "cmp", "private", time.Now())
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

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, &ErrorResponse{Code: res.StatusCode, ErrorMessage: "Error reading response body", Err: err}
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		apiError := APIError{}
		err = json.Unmarshal(data, &apiError)
		if err != nil {
			return nil, &ErrorResponse{Code: res.StatusCode, ErrorMessage: string(data), Err: err}
		}
		return nil, &ErrorResponse{Code: res.StatusCode, APIError: apiError}
	}
	return data, nil
}
