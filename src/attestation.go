package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

type Statement struct {
	Repo struct {
		Path    string `json:"path"`
		OwnerID string `json:"ownerID"`
		RepoID  string `json:"repoID"`
	} `json:"repository"`

	Pkg struct {
		Type    string   `json:"type"`
		Name    string   `json:"name"`
		Version string   `json:"version"`
		Hashes  []string `json:"hashes"`
	} `json:"package"`
}

type Attestation struct {
	Statement string `json:"statement"`
	Evidence  string `json:"evidence"`
	Timestamp string `json:"timestamp"`
}

func genAttestation(repoInfos *RepoInfos, pkgInfos *PKGInfos) (string, error) {
	// generate statement
	statement := Statement{}
	statement.Repo.Path = repoInfos.owner + "/" + repoInfos.name
	statement.Repo.OwnerID = repoInfos.oID
	statement.Repo.RepoID = repoInfos.id
	statement.Pkg.Type = pkgInfos.Type
	statement.Pkg.Name = pkgInfos.Name
	statement.Pkg.Version = pkgInfos.Version
	statement.Pkg.Hashes = pkgInfos.Hashes
	statementBytes, err := json.Marshal(statement)
	if err != nil {
		return "", errors.New("failed to encode statement as JSON string")
	}

	// generate evidence
	hasher := sha256.New()
	_, err = hasher.Write(statementBytes)
	if err != nil {
		return "", errors.New("failed to calculate hash of statement")
	}
	statementHash := hex.EncodeToString(hasher.Sum(nil))
	evidence, err := reqOIDCToken("letsesign:" + statementHash)
	if err != nil {
		fmt.Println(err.Error())
		return "", errors.New("failed to generate evidence")
	}

	// generate timestamp
	timestampBytes, err := getTimestampAPI(evidence)
	if err != nil {
		fmt.Println(err.Error())
		return "", errors.New("failed to generate timestamp")
	}

	// generate attestation
	attestation := Attestation{}
	attestation.Statement = base64.StdEncoding.EncodeToString(statementBytes)
	attestation.Evidence = evidence
	attestation.Timestamp = base64.StdEncoding.EncodeToString(timestampBytes)
	attestationBytes, err := json.Marshal(attestation)
	if err != nil {
		return "", errors.New("failed to encode attestation to JSON string")
	}

	return string(attestationBytes), nil
}
