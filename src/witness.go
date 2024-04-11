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

	// load issue page
	issuePageInfos, err := parseIssuePage(inputs.repoOwner, inputs.repoName, inputs.issueNum)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("Failed to parse issue page")
		os.Exit(1)
	}

	// validate issue
	repoInfos, pkgInfos, err := validateIssue(issuePageInfos)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Printf("Failed to validate issue")
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
