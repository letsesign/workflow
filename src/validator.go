package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/go-resty/resty/v2"
	"gopkg.in/yaml.v3"
)

const ISSUE_MSG = "If the following package info is correct, please lock conversation as resolved and close this issue as compeleted to generate an attestation."

func validateIssue(issueInfos *IssueInfos) (*PKGInfos, error) {
	var pkgInfos PKGInfos

	// check issue lock status
	if !issueInfos.locked || issueInfos.lockReason != "resolved" {
		return nil, errors.New("invalid lock status")
	}

	// check issue state
	if issueInfos.state != "closed" || issueInfos.stateReason != "completed" {
		return nil, errors.New("invalid issue state")
	}

	// check issue body
	if !strings.HasPrefix(issueInfos.body, ISSUE_MSG) {
		return nil, errors.New("invalid issue body")
	}

	// check issue comments
	if issueInfos.comments != 0 {
		return nil, errors.New("unexpected number of issue comments")
	}

	// extract pkgInfos from issue body
	err := yaml.Unmarshal([]byte(strings.Replace(issueInfos.body, ISSUE_MSG, "", -1)), &pkgInfos)
	if err != nil {
		return nil, errors.New("failed to parse package info from issue")
	}

	return &pkgInfos, nil
}

func validatePKG(pkgInfos *PKGInfos) error {
	var collectedHashes []string
	var err error

	// collect binary hashes from registry
	if pkgInfos.Type == "cargo" {
		collectedHashes, err = getCargoHashes(pkgInfos.Name, pkgInfos.Version)
		if err != nil {
			fmt.Println(err.Error())
			return fmt.Errorf("failed to get cargo package hashes")
		}
	} else if pkgInfos.Type == "pypi" {
		collectedHashes, err = getPypiHashes(pkgInfos.Name, pkgInfos.Version)
		if err != nil {
			fmt.Println(err.Error())
			return fmt.Errorf("failed to get pypi package hashes")
		}
	} else {
		return errors.New("unsupported package type")
	}

	// check binary hashes
	for _, hash := range pkgInfos.Hashes {
		if !slices.Contains(collectedHashes, hash) {
			return fmt.Errorf("hash %s is not found on registry", hash)
		}
	}

	return nil
}

func getCargoHashes(name string, version string) ([]string, error) {
	var hashes []string

	url := fmt.Sprintf("https://crates.io/api/v1/crates/%s", name)

	// get package metadata from registry
	resp, err := resty.New().R().Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("error occurred while calling crates.io API")
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("invalid status code of crates.io API: %d", resp.StatusCode())
	}

	var respData map[string]any
	err = json.Unmarshal(resp.Body(), &respData)
	if err != nil {
		fmt.Println(err.Error())
		return nil, fmt.Errorf("failed to parse response of crates.io API")
	}

	// extract binary hashes
	for _, item := range respData["versions"].([]interface{}) {
		pkgData := item.(map[string]any)

		if pkgData["num"].(string) == version {
			hashes = append(hashes, "sha256:"+pkgData["checksum"].(string))
			break
		}
	}

	if len(hashes) == 0 {
		return nil, fmt.Errorf("no package version found on registry")
	}

	return hashes, nil
}

func getPypiHashes(name string, version string) ([]string, error) {
	var hashes []string

	url := fmt.Sprintf("https://pypi.org/pypi/%s/%s/json", name, version)

	// get package metadata from registry
	resp, err := resty.New().R().Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("error occurred while calling pypi.org API")
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("invalid status code of pypi.org API: %d", resp.StatusCode())
	}

	var respData map[string]any
	err = json.Unmarshal(resp.Body(), &respData)
	if err != nil {
		fmt.Println(err.Error())
		return nil, fmt.Errorf("failed to parse response of pypi.org API")
	}

	// extract binary hashes
	for _, item := range respData["urls"].([]interface{}) {
		binaryData := item.(map[string]any)
		digestData := binaryData["digests"].(map[string]any)

		hashes = append(hashes, "sha256:"+digestData["sha256"].(string))
	}

	if len(hashes) == 0 {
		return nil, fmt.Errorf("no package version found on registry")
	}

	return hashes, nil
}
