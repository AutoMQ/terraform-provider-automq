package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const HostURL string = ""

type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
	Auth       AuthStruct
}

// AuthStruct -
type AuthStruct struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

type ErrorResponse struct {
	Code         int                 `json:"code"`
	ErrorMessage string              `json:"error_message"`
	APIError     GenericOpenAPIError `json:"api_error"`
	Err          error               `json:"error"`
}

type GenericOpenAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Errorf("code: %d, message: %s, details: %v", e.Code, e.ErrorMessage, e.Err).Error()
}

func NewClient(host, token *string) (*Client, error) {
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
		apiError := GenericOpenAPIError{}
		err = json.Unmarshal(body, &apiError)
		if err != nil {
			return nil, &ErrorResponse{Code: res.StatusCode, ErrorMessage: "Error unmarshaling response body", Err: err}
		}
		return nil, &ErrorResponse{Code: res.StatusCode, APIError: apiError, Err: nil}
	}

	return body, nil
}
