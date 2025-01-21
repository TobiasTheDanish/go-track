package repo

import (
	"go-track/internal/db"
	"go-track/internal/model"
)

type ProjectRepository interface {
	GetProject(id int) (model.Project, error)
}

type projectRepo struct {
	db db.DatabaseFacade
}

func NewProjectRepo(db db.DatabaseFacade) ProjectRepository {
	return &projectRepo{
		db: db,
	}
}

func (r *projectRepo) GetProject(id int) (model.Project, error) {
	return r.db.GetProject(id)
}
