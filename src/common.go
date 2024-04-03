package main

import (
	"errors"
	"strconv"
	"strings"
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

type RepoInfos struct {
	id    string
	name  string
	oID   string
	owner string
}

func initRepoInfos(repoOwner string, repoName string, client *githubAPIClient) (*RepoInfos, error) {
	repoInfos := RepoInfos{}

	// get repository data
	repoData, err := client.getRepository(repoOwner, repoName)
	if err != nil {
		return nil, err
	}

	// fill RepoInfos struct
	repoInfos.id = strconv.Itoa(int((*repoData)["id"].(float64)))
	repoInfos.name = repoName
	repoInfos.oID = strconv.Itoa(int((*repoData)["owner"].(map[string]any)["id"].(float64)))
	repoInfos.owner = repoOwner

	return &repoInfos, nil
}

type IssueInfos struct {
	title       string
	body        string
	comments    int
	locked      bool
	lockReason  string
	state       string
	stateReason string
}

func initIssueInfos(repoInfos *RepoInfos, issueNum string, client *githubAPIClient) (*IssueInfos, error) {
	issueInfos := IssueInfos{}

	// get issue data
	issueData, err := client.getIssue(repoInfos.id, issueNum)
	if err != nil {
		return nil, err
	}

	// fill IssueInfos struct
	issueInfos.title = (*issueData)["title"].(string)
	issueInfos.body = (*issueData)["body"].(string)
	issueInfos.comments = int((*issueData)["comments"].(float64))
	issueInfos.locked = (*issueData)["locked"].(bool)
	tmpLockReason, hasLockReason := (*issueData)["active_lock_reason"].(string)
	if hasLockReason {
		issueInfos.lockReason = tmpLockReason
	}
	issueInfos.state = (*issueData)["state"].(string)
	tmpStateReason, hasStateReason := (*issueData)["state_reason"].(string)
	if hasStateReason {
		issueInfos.stateReason = tmpStateReason
	}

	return &issueInfos, nil
}

type PKGInfos struct {
	Type    string
	Name    string
	Version string
	Hashes  []string
}
