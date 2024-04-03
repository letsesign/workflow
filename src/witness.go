package main

import (
	"fmt"
	"os"
)

func main() {
	// parse input parameters
	var inputs InputParams
	err := inputs.parse(os.Getenv("INPUT_REPO"), os.Getenv("INPUT_ISSUE_NUM"))
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Invalid input parameters")
		os.Exit(1)
	}

	// new a githubAPIClient
	githubAPIClient, err := newGithubAPIClient()
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Failed to new a github API client")
		os.Exit(1)
	}

	// initialize a repoInfos
	repoInfos, err := initRepoInfos(inputs.repoOwner, inputs.repoName, githubAPIClient)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Failed to initialize a repoInfos")
		os.Exit(1)
	}

	// initialize an issueInfos
	issueInfos, err := initIssueInfos(repoInfos, inputs.issueNum, githubAPIClient)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Printf("Failed to initialize an issueInfos")
		os.Exit(1)
	}

	// validate an issue
	pkgInfos, err := validateIssue(issueInfos)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Printf("Failed to validate issue")
		os.Exit(1)
	}

	// validate a package
	err = validatePKG(pkgInfos)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Printf("Failed to validate package")
		os.Exit(1)
	}

	// generate an attestation
	attestation, err := genAttestation(repoInfos, pkgInfos)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Printf("Failed to generate attestation")
		os.Exit(1)
	}

	// export attestation
	err = exportAttestationAPI(attestation)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Printf("Failed to export attestation")
		os.Exit(1)
	}
}
