package main

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/package-url/packageurl-go"
)

func validateIssue(issuePageInfos *IssuePageInfos) (*RepoInfos, *PKGInfos, error) {
	// check owner ID
	if issuePageInfos.ownerID == "" {
		return nil, nil, errors.New("missing owner ID")
	}

	// check repo ID
	if issuePageInfos.repoID == "" {
		return nil, nil, errors.New("missing repo ID")
	}

	// check issue state
	if !issuePageInfos.isClosedAsCompleted {
		return nil, nil, errors.New("issue is not been closed as completed")
	}

	// check title state
	if issuePageInfos.hasTitleChanged {
		return nil, nil, errors.New("the title has been changed")
	}

	// check title content
	regex := regexp.MustCompile(`\((.*?)\)`)
	matches := regex.FindAllStringSubmatch(issuePageInfos.title, -1)

	if len(matches) != 1 {
		return nil, nil, errors.New("invalid number of matched text within parentheses")
	}

	pkgURLInfo, err := packageurl.FromString(matches[0][1])
	if err != nil {
		return nil, nil, err
	}

	registryType := ""
	if pkgURLInfo.Type == "cargo" {
		registryType = "crates.io"
	} else if pkgURLInfo.Type == "pypi" {
		registryType = "PyPI"
	} else {
		return nil, nil, errors.New("unsupported package type")
	}

	if issuePageInfos.title != fmt.Sprintf("eSign your %s package (%s)", registryType, matches[0][1]) {
		return nil, nil, errors.New("not an expected title content")
	}

	repoInfos := RepoInfos{issuePageInfos.repoID, issuePageInfos.repoName, issuePageInfos.ownerID, issuePageInfos.repoOwner}
	pkgInfos, err := initPKGInfos(pkgURLInfo.Type, pkgURLInfo.Name, pkgURLInfo.Version)
	if err != nil {
		return nil, nil, err
	}

	return &repoInfos, pkgInfos, nil
}
