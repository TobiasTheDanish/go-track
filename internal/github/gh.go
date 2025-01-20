package github

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"go-track/internal/model"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type GithubService interface {
	GetAuthUrl() string
	AuthUserByCode(code string) (model.AuthUserRes, error)
	GetAuthorizedUser(auth model.AuthUserRes) (model.AuthorizedUser, error)
	CreateIssue(owner string, repo string, title string) (CreateIssueRes, error)

	GetBranches(owner string, repo string) ([]BranchDTO, error)
	GetBranch(owner string, repo string, name string) (BranchDTO, error)
	CreateBranch(owner string, repo string, name string, sha string) (BranchDTO, error)

	CreatePullRequest(owner string, repo string, head string, base string, issueNumber int) (PullRequestDTO, error)
}

type githubService struct {
	appId        string
	clientId     string
	clientSecret string
	privateKey   *rsa.PrivateKey
}

func New() (GithubService, error) {
	data := os.Getenv("GITHUB_PRIVATE_KEY")

	private, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(data))
	if err != nil {
		return nil, err
	}

	return &githubService{
		appId:        os.Getenv("GITHUB_APP_ID"),
		clientId:     os.Getenv("GITHUB_CLIENT_ID"),
		clientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		privateKey:   private,
	}, nil
}

func (s *githubService) GetAuthUrl() string {
	return fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s", s.clientId)
}

func (s *githubService) AuthUserByCode(code string) (model.AuthUserRes, error) {
	reqUrl := "https://github.com/login/oauth/access_token?"
	reqVals, err := url.ParseQuery(fmt.Sprintf("code=%s&client_id=%s&client_secret=%s", code, s.clientId, s.clientSecret))
	if err != nil {
		return model.AuthUserRes{}, errors.Join(errors.New("Parsing request query parameters failed"), err)
	}
	reqBody := bytes.NewReader([]byte(reqVals.Encode()))

	req, err := http.NewRequest(http.MethodPost, reqUrl, reqBody)
	if err != nil {
		return model.AuthUserRes{}, errors.Join(errors.New("Creating request for authentication failed."), err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.AuthUserRes{}, errors.Join(errors.New("Getting from oauth url failed."), err)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return model.AuthUserRes{}, errors.Join(errors.New("Reading response body failed."), err)
	}

	if res.StatusCode != 200 {
		return model.AuthUserRes{}, errors.New(fmt.Sprintf("Authentication with code '%s', failed with body: %s", code, resBody))
	}

	var authRes model.AuthUserRes
	err = json.Unmarshal(resBody, &authRes)
	if err != nil {
		return model.AuthUserRes{}, errors.Join(errors.New("Unmarshalling response body failed."), err)
	}

	return authRes, nil
}

func (s *githubService) GetAuthorizedUser(auth model.AuthUserRes) (model.AuthorizedUser, error) {
	reqUrl := "https://api.github.com/user"
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return model.AuthorizedUser{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.AuthorizedUser{}, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return model.AuthorizedUser{}, err
	}

	var user model.AuthorizedUser
	if err := json.Unmarshal(resBody, &user); err != nil {
		return model.AuthorizedUser{}, err
	}

	return user, nil
}

func (s *githubService) getJWT() (string, error) {
	now := time.Now()
	issuedAt := now.Unix() - 60
	expires := now.Unix() + 5*60

	signer := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": issuedAt,
		"exp": expires,
		"iss": s.clientId,
		"alg": "RS256",
	})

	return signer.SignedString(s.privateKey)
}
