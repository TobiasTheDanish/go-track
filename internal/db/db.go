package db

import (
	"database/sql"
	"errors"
	"fmt"
	"go-track/internal/model"
	"log"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type DatabaseFacade interface {
	GetProject(id int) (model.Project, error)
	AddItemToColumn(name string, columnID int) (model.Item, error)
}

type database struct {
	db *sql.DB
}

func New() (DatabaseFacade, error) {
	dbUrl := os.Getenv("TURSO_DB_URL")
	dbToken := os.Getenv("TURSO_TOKEN")

	db, err := sql.Open("libsql", fmt.Sprintf("%s?authToken=%s", dbUrl, dbToken))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not open database connection: %s\n", err.Error()))
	}

	return &database{
		db: db,
	}, nil
}

func (db *database) GetProject(id int) (model.Project, error) {
	row := db.db.QueryRow("SELECT id, name FROM `gt_project` WHERE id=?", id)

	var proj model.Project
	if err := row.Scan(&proj.Id, &proj.Name); err != nil {
		return model.Project{}, err
	}

	log.Printf("Project: %v\n", proj)

	cols, err := db.GetColumnsForProject(id)
	if err != nil {
		return model.Project{}, err
	}

	proj.Columns = cols

	return proj, nil
}

func (db *database) GetColumn(id int) (model.Column, error) {
	row := db.db.QueryRow("SELECT id, name FROM `gt_project_column` WHERE id=?", id)

	var col model.Column
	if err := row.Scan(&col.Id, &col.Name); err != nil {
		return model.Column{}, err
	}

	items, err := db.GetItemsForColumn(id)
	if err != nil {
		return model.Column{}, errors.Join(errors.New("Could not fetch items for column"), err)
	}

	col.Items = items
	return col, nil
}

func (db *database) GetColumnsForProject(projectID int) ([]model.Column, error) {
	rows, err := db.db.Query("SELECT id, name FROM `gt_project_column` WHERE project_id=?", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := make([]model.Column, 0)

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		items, err := db.GetItemsForColumn(id)
		if err != nil {
			return nil, errors.Join(errors.New("Could not fetch items for column"), err)
		}

		cols = append(cols, model.Column{
			Id:    id,
			Name:  name,
			Items: items,
		})
	}

	log.Printf("Columns for %d: %v\n", projectID, cols)

	return cols, err
}

func (db *database) GetItemsForColumn(columnID int) ([]model.Item, error) {
	rows, err := db.db.Query("SELECT * FROM `gt_project_column_item` WHERE column_id=?", columnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Item, 0)

	for rows.Next() {
		var id int
		var name string
		var columnID int
		if err := rows.Scan(&id, &name, &columnID); err != nil {
			return nil, err
		}
		items = append(items, model.Item{
			Id:       id,
			Name:     name,
			ColumnID: columnID,
		})
	}

	return items, nil
}

func (db *database) AddItemToColumn(name string, columnID int) (model.Item, error) {
	res, err := db.db.Exec("INSERT INTO `gt_project_column_item` (name, columnID) values (?, ?)", name, columnID)
	if err != nil {
		return model.Item{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return model.Item{}, err
	}

	return model.Item{
		Id:       int(id),
		Name:     name,
		ColumnID: columnID,
	}, nil
}
