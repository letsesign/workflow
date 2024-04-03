package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

const GITHUB_BASE_PATH = "https://api.github.com"
const GITHUB_GET_REPO_API = GITHUB_BASE_PATH + "/repos/%s/%s"
const GITHUB_GET_ISSUE_API = GITHUB_BASE_PATH + "/repositories/%s/issues/%s"

type githubAPIClient struct {
	token string
}

func newGithubAPIClient() (*githubAPIClient, error) {
	var client githubAPIClient

	githubToken, err := getGithubTokenAPI()
	if err != nil {
		fmt.Println(err.Error())
		return nil, fmt.Errorf("failed to get github token")
	}

	client.token = githubToken

	return &client, nil
}

func _doAPICalling(request *resty.Request, url string) (*map[string]any, error) {
	resp, err := request.Get(url)

	if err != nil {
		fmt.Println(err.Error())
		return nil, errors.New("error occurred while calling Github API")
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("invalid status code of Github API: %d", resp.StatusCode())
	}

	var respData map[string]any
	err = json.Unmarshal(resp.Body(), &respData)
	if err != nil {
		fmt.Println(err.Error())
		return nil, fmt.Errorf("failed to parse response of Github API")
	}

	return &respData, nil
}

func (c githubAPIClient) getRepository(repoOwner string, repoName string) (*map[string]any, error) {
	url := fmt.Sprintf(GITHUB_GET_REPO_API, repoOwner, repoName)

	// setup request instance
	request := resty.New().R().
		SetHeader("Accept", "application/vnd.github+json").
		SetHeader("Authorization", "Bearer "+c.token).
		SetHeader("X-GitHub-Api-Version", "2022-11-28")

	return _doAPICalling(request, url)
}

func (c githubAPIClient) getIssue(repoID string, issueNum string) (*map[string]any, error) {
	url := fmt.Sprintf(GITHUB_GET_ISSUE_API, repoID, issueNum)

	// setup request instance
	request := resty.New().R().
		SetHeader("Accept", "application/vnd.github+json").
		SetHeader("Authorization", "Bearer "+c.token).
		SetHeader("X-GitHub-Api-Version", "2022-11-28")

	return _doAPICalling(request, url)
}
