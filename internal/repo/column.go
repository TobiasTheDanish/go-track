package repo

import (
	"go-track/internal/db"
	"go-track/internal/model"
)

type ColumnRepository interface {
	GetForProject(projectID int) ([]model.Column, error)
	Get(id int) (model.Column, error)
	AddItem(name string, columnID int) (model.Item, error)
	RemoveItem(itemID, columnID int) (model.Column, error)
}

type columnRepo struct {
	db db.DatabaseFacade
}

func NewColumnRepo(db db.DatabaseFacade) ColumnRepository {
	return &columnRepo{
		db: db,
	}
}

func (r *columnRepo) GetForProject(projectID int) ([]model.Column, error) {
	return r.db.GetColumnsForProject(projectID)
}

func (r *columnRepo) Get(id int) (model.Column, error) {
	return r.db.GetColumn(id)
}

func (r *columnRepo) AddItem(name string, columnID int) (model.Item, error) {
	return r.db.AddItemToColumn(name, columnID)
}

func (r *columnRepo) RemoveItem(itemID, columnID int) (model.Column, error) {
	err := r.db.DeleteItem(itemID)
	if err != nil {
		return model.Column{}, err
	}

	return r.db.GetColumn(columnID)
}
