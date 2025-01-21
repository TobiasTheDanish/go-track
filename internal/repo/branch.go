package repo

import (
	"go-track/internal/github"
	"go-track/internal/model"
)

type BranchRepository interface {
	GetAll(owner, repo string) ([]model.Branch, error)
	Get(owner, repo, name string) (model.Branch, error)
}

type branchRepo struct {
	gh github.GithubService
}

func NewBranchRepo(gh github.GithubService) BranchRepository {
	return &branchRepo{
		gh: gh,
	}
}

func (r *branchRepo) GetAll(owner, repo string) ([]model.Branch, error) {
	branchDTOs, err := r.gh.GetBranches(owner, repo)
	if err != nil {
		return nil, err
	}

	branches := make([]model.Branch, len(branchDTOs), len(branchDTOs))
	for i, b := range branchDTOs {
		branches[i] = model.Branch{
			Name: b.Name,
			Sha:  b.Commit.Sha,
		}
	}

	return branches, nil
}
func (r *branchRepo) Get(owner, repo, name string) (model.Branch, error) {
	b, err := r.gh.GetBranch(owner, repo, name)
	if err != nil {
		return model.Branch{}, err
	}

	return model.Branch{
		Name: b.Name,
		Sha:  b.Commit.Sha,
	}, nil
}
