package client

import (
	"encoding/json"
	"net/http"
	"strings"
)

// CreateUser creates a new user
func (c *Client) CreateKafkaUser(instanceId string, user InstanceUserCreateParam) (*KafkaUserVO, error) {
	userRequest, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	var localVarPath = c.HostURL + "/api/v1/instances/" + instanceId + "/users"
	req, err := http.NewRequest("POST", localVarPath, strings.NewReader(string(userRequest)))
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	newuser := KafkaUserVO{}
	err = json.Unmarshal(body, &newuser)
	if err != nil {
		return nil, err
	}
	return &newuser, nil
}

func (c *Client) DeleteKafkaUser(instanceId string, userName string) error {
	var localVarPath = c.HostURL + "/api/v1/instances/" + instanceId + "/users/" + userName
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

func (c *Client) GetKafkaUser(instanceId string, userName string) (*KafkaUserVO, error) {
	var localVarPath = c.HostURL + "/api/v1/instances/" + instanceId + "/users"
	localVarPath = localVarPath + "?userNames=" + userName
	req, err := http.NewRequest("GET", localVarPath, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	userPage := PageNumResultKafkaUserVO{}
	err = json.Unmarshal(body, &userPage)
	if err != nil {
		return nil, err
	}
	if len(userPage.List) == 0 {
		return nil, nil
	}
	user := userPage.List[0]
	return &user, nil
}
