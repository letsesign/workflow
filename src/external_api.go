package main

import (
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/digitorus/timestamp"
	"github.com/go-resty/resty/v2"
)

const TSA_URL = "http://timestamp.digicert.com"

func getGithubTokenAPI() (string, error) {
	oidcToken, err := reqOIDCToken("letsesign")
	if err != nil {
		fmt.Println(err.Error())
		return "", errors.New("failed to request Github OIDC token")
	}

	resp, err := resty.New().R().
		SetQueryString("auth=" + oidcToken).
		Get("https://meis5lxrn2x7ctmljmb4iwiupy0onycc.lambda-url.us-east-1.on.aws")

	if err != nil {
		fmt.Println(err.Error())
		return "", errors.New("error occurred while getting Github token")
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("invalid status code of get Github token API %d", resp.StatusCode())
	}

	var respData map[string]any
	err = json.Unmarshal(resp.Body(), &respData)
	if err != nil {
		fmt.Println(err.Error())
		return "", fmt.Errorf("failed to parse response of get Github token API")
	}

	return respData["token"].(string), nil
}

func exportAttestationAPI(attestation string) error {
	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(attestation).
		Post("https://bewig66i2wmwtom5llgvbbfs2q0dgsux.lambda-url.us-east-1.on.aws")

	if err != nil {
		fmt.Println(err.Error())
		return errors.New("error occurred while exporting attestation")
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("invalid status code of export attestation API: %d", resp.StatusCode())
	}

	return nil
}

func getTimestampAPI(data string) ([]byte, error) {
	// create timestamp request
	tsq, err := timestamp.CreateRequest(strings.NewReader(data), &timestamp.RequestOptions{
		Hash:         crypto.SHA256,
		Certificates: true,
	})
	if err != nil {
		return nil, errors.New("failed to create timestamp request")
	}

	// request timestamp from TSA
	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/timestamp-query").
		SetBody(tsq).
		Post(TSA_URL)
	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("error occurred while requesting timestamp")
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("invalid timestamp status code (%d)", resp.StatusCode())
	}

	return resp.Body(), nil
}
