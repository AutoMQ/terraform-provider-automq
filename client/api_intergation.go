package client

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (c *Client) CreateIntergration(param IntegrationParam) (*IntegrationVO, error) {
	integrationRequest, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	localVarPath := c.HostURL + "/api/v1/integrations"
	req, err := http.NewRequest("POST", localVarPath, strings.NewReader(string(integrationRequest)))
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	integration := IntegrationVO{}
	err = json.Unmarshal(body, &integration)
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

func (c *Client) GetIntergration(integrationId string) (*IntegrationVO, error) {
	localVarPath := c.HostURL + "/api/v1/integrations/" + integrationId
	req, err := http.NewRequest("GET", localVarPath, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	integration := IntegrationVO{}
	err = json.Unmarshal(body, &integration)
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

func (c *Client) UpdateIntergration(integrationId string, param *IntegrationParam) (*IntegrationVO, error) {
	integrationRequest, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	localVarPath := c.HostURL + "/api/v1/integrations/" + integrationId
	req, err := http.NewRequest("PATCH", localVarPath, strings.NewReader(string(integrationRequest)))
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	integration := IntegrationVO{}
	err = json.Unmarshal(body, &integration)
	if err != nil {
		return nil, err
	}
	return &integration, nil
}

func (c *Client) DeleteIntergration(integrationId string) error {
	localVarPath := c.HostURL + "/api/v1/integrations/" + integrationId
	req, err := http.NewRequest("DELETE", localVarPath, nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ListInstanceIntegrations(instanceId string) ([]IntegrationVO, error) {
	localVarPath := c.HostURL + "/api/v1/instances/" + instanceId + "/integrations"
	req, err := http.NewRequest("GET", localVarPath, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var integrations PageNumResultIntegrationVO
	err = json.Unmarshal(body, &integrations)
	if err != nil {
		return nil, err
	}
	return integrations.List, nil
}
