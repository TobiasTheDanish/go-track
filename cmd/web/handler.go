package web

import (
	"go-track/internal/db"
	"go-track/internal/github"
)

type Handler struct {
	db db.DatabaseFacade
	gh github.GithubService
}

func NewHandler(db db.DatabaseFacade, gh github.GithubService) *Handler {
	return &Handler{
		db: db,
		gh: gh,
	}
}
