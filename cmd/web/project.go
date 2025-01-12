package web

import (
	"fmt"
	"go-track/internal/model"
	"log"
	"net/http"
	"slices"
	"strconv"

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

func (h *Handler) ProjectItemHandler(c echo.Context) error {
	name := c.FormValue("name")
	columnID, err := strconv.Atoi(c.FormValue("column"))
	if err != nil {
		log.Printf("Error parsing columnID in ProjectItemHandler: %e", err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	item, err := h.db.AddItemToColumn(name, columnID)
	component := ProjectItem(item)
	if len(name) == 0 {
		return c.String(http.StatusOK, "")
	} else {
		return component.Render(c.Request().Context(), c.Response().Writer)
	}
}

func DeleteProjectItemHandler(c echo.Context) error {
	columnID, err := strconv.Atoi(c.Param("colID"))
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid columnID: %s", err.Error()))
	}
	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Invalid itemID: %s", err.Error()))
	}

	log.Printf("ColumnID: %d, ItemID: %d\n", columnID, itemID)

	return ProjectColumn(columns[colIndex]).Render(c.Request().Context(), c.Response().Writer)
}
