package web

import (
	"go-track/internal/db"
	"go-track/internal/github"
	"go-track/internal/repo"
)

type Handler struct {
	projectRepo repo.ProjectRepository
	columnRepo  repo.ColumnRepository
	itemRepo    repo.ItemRepository
	branchRepo  repo.BranchRepository
	authRepo    repo.AuthRepository
}

func NewHandler(db db.DatabaseFacade, gh github.GithubService) *Handler {
	return &Handler{
		projectRepo: repo.NewProjectRepo(db),
		columnRepo:  repo.NewColumnRepo(db),
		itemRepo:    repo.NewItemRepo(db, gh),
		branchRepo:  repo.NewBranchRepo(gh),
		authRepo:    repo.NewAuthRepo(gh),
	}
}
