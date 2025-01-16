package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type createIssueBody struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Assignees []string `json:"assignees"`
	Milestone *string  `json:"milestone"`
	Labels    []string `json:"labels"`
}

func emptyIssue(title string) createIssueBody {
	return createIssueBody{
		Title:     title,
		Body:      "",
		Assignees: make([]string, 0),
		Labels:    make([]string, 0),
		Milestone: nil,
	}
}

type CreateIssueRes struct {
	Id      int64  `json:"id"`
	HtmlUrl string `json:"html_url"`
	Number  int    `json:"number"`
}

func (gh *githubService) CreateIssue(owner string, repo string, title string) (CreateIssueRes, error) {
	issue := emptyIssue(title)

	installation, err := gh.GetUserInstallation(owner)
	if err != nil {
		return CreateIssueRes{}, err
	}

	access, err := gh.GetInstallationAccessToken(installation)
	if err != nil {
		return CreateIssueRes{}, err
	}

	reqBody, err := json.Marshal(issue)
	if err != nil {
		return CreateIssueRes{}, err
	}

	bodyReader := bytes.NewReader(reqBody)
	reqUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", owner, repo)
	req, err := http.NewRequest(http.MethodPost, reqUrl, bodyReader)
	if err != nil {
		return CreateIssueRes{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", access.Token))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return CreateIssueRes{}, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return CreateIssueRes{}, err
	}

	if res.StatusCode != 201 {
		return CreateIssueRes{}, errors.New(fmt.Sprintf("Creating issue for repo: '%s/%s', failed with body: %s", owner, repo, resBody))
	}

	var issueRes CreateIssueRes
	err = json.Unmarshal(resBody, &issueRes)
	if err != nil {
		return CreateIssueRes{}, err
	}

	return issueRes, nil
}

type Installation interface {
	GetId() int
}

type userInstallationRes struct {
	Id int `json:"id"`
}

func (i userInstallationRes) GetId() int { return i.Id }

func (s *githubService) GetUserInstallation(username string) (Installation, error) {
	token, err := s.getJWT()
	if err != nil {
		return nil, err
	}

	reqUrl := fmt.Sprintf("https://api.github.com/users/%s/installation", username)
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
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
		return nil, errors.New(fmt.Sprintf("Getting installation for user: '%s', failed with body: %s", username, resBody))
	}

	var installationRes userInstallationRes
	err = json.Unmarshal(resBody, &installationRes)
	if err != nil {
		return nil, err
	}

	return installationRes, nil
}

type installationAccess struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

func (s *githubService) GetInstallationAccessToken(i Installation) (*installationAccess, error) {
	token, err := s.getJWT()
	if err != nil {
		return nil, err
	}

	reqUrl := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", i.GetId())
	req, err := http.NewRequest(http.MethodPost, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 201 {
		return nil, errors.New(fmt.Sprintf("Creating installation access token for id: %d, failed with body: %s", i.GetId(), resBody))
	}

	var access installationAccess
	err = json.Unmarshal(resBody, &access)
	if err != nil {
		return nil, err
	}

	return &access, nil
}
