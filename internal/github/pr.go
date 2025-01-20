package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type PullRequestDTO struct {
	Id     int    `json:"id"`
	Number int    `json:"number"`
	Url    string `json:"html_url"`
}

type createPullRequestDTO struct {
	Head  string `json:"head"`
	Base  string `json:"base"`
	Issue int    `json:"issue"`
}

func (gh *githubService) CreatePullRequest(owner string, repo string, head string, base string, issueNumber int) (PullRequestDTO, error) {
	installation, err := gh.GetUserInstallation(owner)
	if err != nil {
		return PullRequestDTO{}, err
	}

	access, err := gh.GetInstallationAccessToken(installation)
	if err != nil {
		return PullRequestDTO{}, err
	}

	pr := createPullRequestDTO{
		Head:  head,
		Base:  base,
		Issue: issueNumber,
	}

	reqBody, err := json.Marshal(pr)
	if err != nil {
		return PullRequestDTO{}, err
	}

	bodyReader := bytes.NewReader(reqBody)

	reqUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", owner, repo)
	req, err := http.NewRequest(http.MethodPost, reqUrl, bodyReader)
	if err != nil {
		return PullRequestDTO{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", access.Token))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return PullRequestDTO{}, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return PullRequestDTO{}, err
	}

	if res.StatusCode != 201 {
		return PullRequestDTO{}, errors.New(fmt.Sprintf("Creating pr for repo: '%s/%s', failed with body: %s", owner, repo, resBody))
	}

	var dto PullRequestDTO
	err = json.Unmarshal(resBody, &dto)
	if err != nil {
		return PullRequestDTO{}, err
	}

	return dto, nil
}
