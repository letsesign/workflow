package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
)

func isStringInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

type InputParams struct {
	repoOwner string
	repoName  string
	issueNum  string
}

func (inputs *InputParams) parse(inputRepo string, inputIssueNum string) error {
	// split repository path
	splitedRepoPath := strings.Split(inputRepo, "/")
	if len(splitedRepoPath) != 2 {
		return errors.New("invalid repository format")
	}

	// parse issue number
	if !isStringInt(inputIssueNum) {
		return errors.New("invalid issue number format")
	}

	inputs.repoOwner = splitedRepoPath[0]
	inputs.repoName = splitedRepoPath[1]
	inputs.issueNum = inputIssueNum

	return nil
}

type IssuePageInfos struct {
	title               string
	repoOwner           string
	repoName            string
	ownerID             string
	repoID              string
	hasTitleChanged     bool
	isClosedAsCompleted bool
}

type RepoInfos struct {
	id    string
	name  string
	oID   string
	owner string
}

type PKGInfos struct {
	Type    string
	Name    string
	Version string
	Hashes  []string
}

func initPKGInfos(pkgType string, pkgName string, pkgVer string) (*PKGInfos, error) {
	var hashes []string
	var err error

	// collect binary hashes from registry
	if pkgType == "cargo" {
		hashes, err = getCargoHashes(pkgName, pkgVer)
		if err != nil {
			fmt.Println(err.Error())
			return nil, fmt.Errorf("failed to get cargo package hashes")
		}
	} else if pkgType == "pypi" {
		hashes, err = getPypiHashes(pkgName, pkgVer)
		if err != nil {
			fmt.Println(err.Error())
			return nil, fmt.Errorf("failed to get pypi package hashes")
		}
	} else {
		return nil, errors.New("unsupported package type")
	}

	pkgInfos := PKGInfos{pkgType, pkgName, pkgVer, hashes}

	return &pkgInfos, nil
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
