package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
)

func reqOIDCToken(audience string) (string, error) {
	oidcReqURL := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	oidcReqToken := os.Getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")

	oidcReqURL = oidcReqURL + "&audience=" + audience

	request := resty.New().R().
		SetHeader("Authorization", "Bearer "+oidcReqToken)
	resp, err := request.Get(oidcReqURL)

	if err != nil {
		fmt.Println(err.Error())
		return "", errors.New("error occurred while requesting Github OIDC token")
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("invalid status code of request OIDC token API: %d", resp.StatusCode())
	}

	var respData map[string]any
	err = json.Unmarshal(resp.Body(), &respData)
	if err != nil {
		fmt.Println(err.Error())
		return "", fmt.Errorf("failed to parse response of Github OIDC token API")
	}

	return respData["value"].(string), nil
}
