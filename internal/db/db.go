package db

import (
	"database/sql"
	"errors"
	"fmt"
	"go-track/internal/model"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type DatabaseFacade interface {
	GetProject(id int) (model.Project, error)
	GetColumnsForProject(projectID int) ([]model.Column, error)
	GetColumn(id int) (model.Column, error)
	AddItemToColumn(name string, columnID int) (model.Item, error)
	GetNextItemColumnOrder(columnID int) (int, error)

	GetItem(id int) (model.Item, error)
	UpdateItem(id int, item model.Item) (model.Item, error)
	DeleteItem(itemID int) error
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

	cols, err := db.GetColumnsForProject(id)
	if err != nil {
		return model.Project{}, err
	}

	proj.Columns = cols

	return proj, nil
}

func (db *database) GetColumn(id int) (model.Column, error) {
	row := db.db.QueryRow("SELECT id, name, project_id FROM `gt_project_column` WHERE id=?", id)

	var col model.Column
	if err := row.Scan(&col.Id, &col.Name, &col.ProjectID); err != nil {
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
	rows, err := db.db.Query("SELECT id, name, project_id FROM `gt_project_column` WHERE project_id=?", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := make([]model.Column, 0)

	for rows.Next() {
		var id int
		var name string
		var projectID int
		if err := rows.Scan(&id, &name, &projectID); err != nil {
			return nil, err
		}
		items, err := db.GetItemsForColumn(id)
		if err != nil {
			return nil, errors.Join(errors.New("Could not fetch items for column"), err)
		}

		cols = append(cols, model.Column{
			Id:        id,
			Name:      name,
			ProjectID: projectID,
			Items:     items,
		})
	}

	return cols, err
}

func (db *database) GetItemsForColumn(columnID int) ([]model.Item, error) {
	rows, err := db.db.Query("SELECT * FROM `gt_project_column_item` WHERE column_id=? ORDER BY column_order", columnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Item, 0)

	for rows.Next() {
		var item model.Item
		if err := rows.Scan(
			&item.Id,
			&item.Name,
			&item.ColumnID,
			&item.ColumnOrder,
			&item.IssueID,
			&item.IssueNumber,
			&item.IssueUrl,
			&item.BranchName,
			&item.PullRequestID,
			&item.PullRequestNumber,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (db *database) AddItemToColumn(name string, columnID int) (model.Item, error) {
	colOrder, err := db.GetNextItemColumnOrder(columnID)
	if err != nil {
		return model.Item{}, err
	}

	res, err := db.db.Exec("INSERT INTO `gt_project_column_item` (name, column_id, column_order) values (?, ?, ?)", name, columnID, colOrder)
	if err != nil {
		return model.Item{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return model.Item{}, err
	}

	return model.Item{
		Id:                int(id),
		Name:              name,
		ColumnID:          columnID,
		ColumnOrder:       colOrder,
		IssueID:           -1,
		IssueNumber:       -1,
		IssueUrl:          "",
		BranchName:        "",
		PullRequestID:     -1,
		PullRequestNumber: -1,
	}, nil
}

func (db *database) GetNextItemColumnOrder(columnID int) (int, error) {
	res := db.db.QueryRow("SELECT column_order FROM `gt_project_column_item` WHERE column_id=? ORDER BY column_order DESC", columnID)

	var colOrder int
	if err := res.Scan(&colOrder); err != nil {
		if err == sql.ErrNoRows {
			return 1, nil
		}
		return -1, err
	}

	return colOrder + 1, nil
}

func (db *database) GetItem(itemID int) (model.Item, error) {
	res := db.db.QueryRow("SELECT * FROM `gt_project_column_item` WHERE id=?", itemID)

	var item model.Item
	if err := res.Scan(
		&item.Id,
		&item.Name,
		&item.ColumnID,
		&item.ColumnOrder,
		&item.IssueID,
		&item.IssueNumber,
		&item.IssueUrl,
		&item.BranchName,
		&item.PullRequestID,
		&item.PullRequestNumber,
	); err != nil {
		return model.Item{}, err
	}

	return item, nil
}

func (db *database) UpdateItem(id int, itemData model.Item) (model.Item, error) {
	res := db.db.QueryRow("UPDATE `gt_project_column_item` SET name=?, column_id=?, column_order=?, gh_issue_no=?, gh_issue_id=?, gh_issue_url=?, gh_branch_name=? WHERE id=? RETURNING *", itemData.Name, itemData.ColumnID, itemData.ColumnOrder, itemData.IssueNumber, itemData.IssueID, itemData.IssueUrl, itemData.BranchName, itemData.Id)

	var item model.Item
	if err := res.Scan(
		&item.Id,
		&item.Name,
		&item.ColumnID,
		&item.ColumnOrder,
		&item.IssueID,
		&item.IssueNumber,
		&item.IssueUrl,
		&item.BranchName,
		&item.PullRequestID,
		&item.PullRequestNumber,
	); err != nil {
		return model.Item{}, err
	}

	return item, nil
}

func (db *database) DeleteItem(itemID int) error {
	_, err := db.db.Exec("DELETE FROM `gt_project_column_item` WHERE id=?", itemID)
	if err != nil {
		return err
	}

	return nil
}
