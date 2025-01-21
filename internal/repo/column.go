package repo

import (
	"fmt"
	"go-track/internal/db"
	"go-track/internal/model"
	"strings"
)

type ColumnRepository interface {
	GetColumnsForProject(projectID int) ([]model.Column, error)
	GetColumn(id int) (model.Column, error)
	AddItemToColumn(name string, columnID int) (model.Item, error)
}

type columnRepo struct {
	db db.DatabaseFacade
}

func NewColumnRepo(db db.DatabaseFacade) ColumnRepository {
	return &columnRepo{
		db: db,
	}
}

func (r *columnRepo) GetColumnsForProject(projectID int) ([]model.Column, error) {
	return r.db.GetColumnsForProject(projectID)
}

func (r *columnRepo) GetColumn(id int) (model.Column, error) {
	return r.db.GetColumn(id)
}

func (r *columnRepo) AddItemToColumn(name string, columnID int) (model.Item, error) {
	return r.db.AddItemToColumn(name, columnID)
}
