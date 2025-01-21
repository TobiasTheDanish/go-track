package repo

import (
	"go-track/internal/github"
	"go-track/internal/model"
)

type AuthRepository interface {
	GetAuthUrl() string
	AuthorizeUser(code string) (model.AuthorizedUser, error)
}

type authRepo struct {
	gh github.GithubService
}

func NewAuthRepo(gh github.GithubService) AuthRepository {
	return &authRepo{
		gh: gh,
	}
}

func (r *authRepo) GetAuthUrl() string {
	return r.gh.GetAuthUrl()
}
func (r *authRepo) AuthorizeUser(code string) (model.AuthorizedUser, error) {

	authUser, err := r.gh.AuthUserByCode(code)
	if err != nil {
		return model.AuthorizedUser{}, err
	}

	return r.gh.GetAuthorizedUser(authUser)
}
