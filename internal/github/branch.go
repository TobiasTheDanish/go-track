package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type BranchDTO struct {
	Name   string `json:"name"`
	Commit struct {
		Sha string `json:"sha"`
	} `json:"commit"`
}

func (gh *githubService) GetBranches(owner string, repo string) ([]BranchDTO, error) {
	installation, err := gh.GetUserInstallation(owner)
	if err != nil {
		return nil, err
	}

	access, err := gh.GetInstallationAccessToken(installation)
	if err != nil {
		return nil, err
	}

	reqUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches", owner, repo)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", access.Token))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Creating issue for repo: '%s/%s', failed with body: %s", owner, repo, resBody))
	}

	branches := make([]BranchDTO, 0, 0)
	err = json.Unmarshal(resBody, &branches)
	if err != nil {
		return nil, err
	}

	return branches, nil
}

func (gh *githubService) GetBranch(owner, repo, name string) (BranchDTO, error) {
	installation, err := gh.GetUserInstallation(owner)
	if err != nil {
		return BranchDTO{}, err
	}

	access, err := gh.GetInstallationAccessToken(installation)
	if err != nil {
		return BranchDTO{}, err
	}

	reqUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/branches/%s", owner, repo, name)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return BranchDTO{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", access.Token))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return BranchDTO{}, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return BranchDTO{}, err
	}

	if res.StatusCode != 200 {
		return BranchDTO{}, errors.New(fmt.Sprintf("Creating issue for repo: '%s/%s', failed with body: %s", owner, repo, resBody))
	}

	var branch BranchDTO
	err = json.Unmarshal(resBody, &branch)
	if err != nil {
		return BranchDTO{}, err
	}

	return branch, nil
}

type githubReference struct {
	Ref string `json:"ref"`
	Sha string `json:"sha"`
}

func (gh *githubService) CreateBranch(owner, repo, name, fromSha string) (BranchDTO, error) {
	installation, err := gh.GetUserInstallation(owner)
	if err != nil {
		return BranchDTO{}, err
	}

	access, err := gh.GetInstallationAccessToken(installation)
	if err != nil {
		return BranchDTO{}, err
	}

	ref := githubReference{
		Ref: fmt.Sprintf("refs/heads/%s", name),
		Sha: fromSha,
	}

	reqBody, err := json.Marshal(ref)
	if err != nil {
		return BranchDTO{}, err
	}

	bodyReader := bytes.NewReader(reqBody)
	reqUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs", owner, repo)
	req, err := http.NewRequest(http.MethodPost, reqUrl, bodyReader)
	if err != nil {
		return BranchDTO{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", access.Token))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return BranchDTO{}, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return BranchDTO{}, err
	}

	if res.StatusCode != 201 {
		return BranchDTO{}, errors.New(fmt.Sprintf("Creating branch for repo: '%s/%s', failed with body: %s", owner, repo, resBody))
	}

	return BranchDTO{
		Name: name,
		Commit: struct {
			Sha string `json:"sha"`
		}{
			Sha: fromSha,
		},
	}, nil
}
