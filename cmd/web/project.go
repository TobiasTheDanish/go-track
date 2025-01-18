package web

import (
	"errors"
	"fmt"
	view "go-track/cmd/web/view"
	"go-track/internal/model"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *Handler) ProjectPageHandler(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	proj, err := h.db.GetProject(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	log.Printf("%s, %v\n", proj.Name, proj.Columns)

	return view.ProjectPage(proj).Render(c.Request().Context(), c.Response().Writer)
}

type moveItemRequest struct {
	Direction string `query:"dir"`
}

func (h *Handler) MoveProjectItemHandler(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	dir := c.QueryParam("dir")

	item, err := h.db.GetItem(itemID)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	newColId := -1
	switch strings.ToLower(dir) {
	case "left":
		newColId, err = h.moveItemLeft(id, item)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		break
	case "right":
		newColId, err = h.moveItemRight(id, item)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		break
	case "up":
		err = h.moveItemUp(item)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		break
	case "down":
		err = h.moveItemDown(item)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		break
	}

	if newColId != -1 {
		item.ColumnID = newColId
		err = h.itemEnter(newColId, item)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	cols, err := h.db.GetColumnsForProject(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return view.ProjectColumns(cols, view.ModalState{Show: false}).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) itemEnter(colID int, item model.Item) error {
	col, err := h.db.GetColumn(colID)
	if err != nil {
		return err
	}

	switch strings.ToLower(col.Name) {
	case "backlog":
		break
	case "todo":
		// create new issue
		if item.IssueID == -1 {
			issue, err := h.gh.CreateIssue("TobiasTheDanish", "go-track", item.Name)
			if err != nil {
				return err
			}

			item.IssueID = issue.Id
			item.IssueNumber = issue.Number
			item.IssueUrl = issue.HtmlUrl

			_, err = h.db.UpdateItem(item.Id, item)
			if err != nil {
				return err
			}
		}
		break
	case "in progress":
		// create branch for issue
		break
	case "ready for pull request":
		// create pr for branch
		break
	case "done":
		// close pr and issue
		break
	}

	return nil
}

func (h *Handler) moveItemDown(item model.Item) error {
	col, err := h.db.GetColumn(item.ColumnID)
	if err != nil {
		return err
	}

	itemToSwap := model.Item{ColumnOrder: -1}
	for _, i := range col.Items {
		if i.ColumnOrder > item.ColumnOrder {
			if itemToSwap.ColumnOrder == -1 || i.ColumnOrder < itemToSwap.ColumnOrder {
				itemToSwap = i
			}
		}
	}

	temp := item.ColumnOrder
	item.ColumnOrder = itemToSwap.ColumnOrder
	itemToSwap.ColumnOrder = temp

	_, err = h.db.UpdateItem(item.Id, item)
	if err != nil {
		return err
	}
	_, err = h.db.UpdateItem(itemToSwap.Id, itemToSwap)
	return err
}

func (h *Handler) moveItemUp(item model.Item) error {
	col, err := h.db.GetColumn(item.ColumnID)
	if err != nil {
		return err
	}

	itemToSwap := model.Item{ColumnOrder: -1}
	for _, i := range col.Items {
		if i.ColumnOrder < item.ColumnOrder {
			if i.ColumnOrder > itemToSwap.ColumnOrder {
				itemToSwap = i
			}
		}
	}

	temp := item.ColumnOrder
	item.ColumnOrder = itemToSwap.ColumnOrder
	itemToSwap.ColumnOrder = temp

	_, err = h.db.UpdateItem(item.Id, item)
	if err != nil {
		return err
	}
	_, err = h.db.UpdateItem(itemToSwap.Id, itemToSwap)
	return err
}

func (h *Handler) moveItemRight(projID int, item model.Item) (int, error) {
	proj, err := h.db.GetProject(projID)
	if err != nil {
		return -1, err
	}

	colIndex := len(proj.Columns)
	for i, col := range proj.Columns {
		if col.Id == item.ColumnID {
			colIndex = i
			break
		}
	}

	if colIndex >= len(proj.Columns)+1 {
		return -1, errors.New("Could not move item right")
	}

	newCol := proj.Columns[colIndex+1]

	colOrder, err := h.db.GetNextItemColumnOrder(newCol.Id)
	if err != nil {
		return -1, err
	}

	item.ColumnID = newCol.Id
	item.ColumnOrder = colOrder
	log.Printf("new item %v\n", item)

	_, err = h.db.UpdateItem(item.Id, item)
	if err != nil {
		return -1, err
	}

	return newCol.Id, nil
}

func (h *Handler) moveItemLeft(projID int, item model.Item) (int, error) {
	proj, err := h.db.GetProject(projID)
	if err != nil {
		return -1, err
	}

	colIndex := -1
	for i, col := range proj.Columns {
		if col.Id == item.ColumnID {
			colIndex = i
			break
		}
	}

	if colIndex <= 0 {
		return -1, errors.New("Could not move item left")
	}

	newCol := proj.Columns[colIndex-1]

	colOrder, err := h.db.GetNextItemColumnOrder(newCol.Id)
	if err != nil {
		return -1, err
	}

	item.ColumnID = newCol.Id
	item.ColumnOrder = colOrder

	_, err = h.db.UpdateItem(item.Id, item)
	if err != nil {
		return -1, err
	}
	return newCol.Id, err
}

func (h *Handler) ProjectItemHandler(c echo.Context) error {
	name := c.FormValue("name")
	if len(name) == 0 {
		return c.String(http.StatusOK, "")
	}
	columnID, err := strconv.Atoi(c.FormValue("column"))
	if err != nil {
		log.Printf("Error parsing columnID in ProjectItemHandler: %e", err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	item, err := h.db.AddItemToColumn(name, columnID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	col, err := h.db.GetColumn(columnID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return view.ProjectItem(col.ProjectID, item).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) DeleteProjectItemHandler(c echo.Context) error {
	columnID, err := strconv.Atoi(c.Param("colID"))
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid columnID: %s", err.Error()))
	}
	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid itemID: %s", err.Error()))
	}

	log.Printf("ColumnID: %d, ItemID: %d\n", columnID, itemID)

	err = h.db.DeleteItem(itemID)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Could not delete item: %s", err.Error()))
	}

	column, err := h.db.GetColumn(columnID)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Could not delete item: %s", err.Error()))
	}

	return view.ProjectColumn(column).Render(c.Request().Context(), c.Response().Writer)
}
