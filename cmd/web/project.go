package web

import (
	"errors"
	"fmt"
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

	return ProjectPage(proj).Render(c.Request().Context(), c.Response().Writer)
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

	switch strings.ToLower(dir) {
	case "left":
		err = h.moveItemLeft(id, item)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		break
	case "right":
		err = h.moveItemRight(id, item)
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

	cols, err := h.db.GetColumnsForProject(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return ProjectColumns(cols).Render(c.Request().Context(), c.Response().Writer)
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

func (h *Handler) moveItemRight(projID int, item model.Item) error {
	proj, err := h.db.GetProject(projID)
	if err != nil {
		return err
	}

	colIndex := len(proj.Columns)
	for i, col := range proj.Columns {
		if col.Id == item.ColumnID {
			colIndex = i
			break
		}
	}

	if colIndex >= len(proj.Columns)+1 {
		return errors.New("Could not move item right")
	}

	newCol := proj.Columns[colIndex+1]

	colOrder, err := h.db.GetNextItemColumnOrder(newCol.Id)
	if err != nil {
		return err
	}

	item.ColumnID = newCol.Id
	item.ColumnOrder = colOrder
	log.Printf("new item %v\n", item)

	_, err = h.db.UpdateItem(item.Id, item)
	return err
}

func (h *Handler) moveItemLeft(projID int, item model.Item) error {
	proj, err := h.db.GetProject(projID)
	if err != nil {
		return err
	}

	colIndex := -1
	for i, col := range proj.Columns {
		if col.Id == item.ColumnID {
			colIndex = i
			break
		}
	}

	if colIndex <= 0 {
		return errors.New("Could not move item left")
	}

	newCol := proj.Columns[colIndex-1]

	colOrder, err := h.db.GetNextItemColumnOrder(newCol.Id)
	if err != nil {
		return err
	}

	item.ColumnID = newCol.Id
	item.ColumnOrder = colOrder

	_, err = h.db.UpdateItem(item.Id, item)
	return err
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

	return ProjectItem(col.ProjectID, item).Render(c.Request().Context(), c.Response().Writer)
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

	return ProjectColumn(column).Render(c.Request().Context(), c.Response().Writer)
}
