package web

import (
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

	proj, err := h.projectRepo.GetProject(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	modalState := view.ModalState{
		Show: false,
	}

	return view.ProjectPage(proj, modalState).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) ProjectColumnsHandler(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	cols, err := h.columnRepo.GetForProject(id)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	modalState := view.ModalState{
		Show: false,
	}

	return view.ProjectColumns(cols, modalState).Render(c.Request().Context(), c.Response().Writer)
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

	oldItem, err := h.itemRepo.Get(itemID)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	dir := c.QueryParam("dir")

	movedItem, err := h.itemRepo.Move(id, itemID, dir)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	modalState := view.ModalState{Show: false}
	if movedItem.ColumnID != oldItem.ColumnID {
		modalState, err = h.itemEnter(id, movedItem.ColumnID, movedItem)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	cols, err := h.columnRepo.GetForProject(id)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return view.ProjectColumns(cols, modalState).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) itemEnter(projID, colID int, item model.Item) (view.ModalState, error) {
	col, err := h.columnRepo.Get(colID)
	if err != nil {
		return view.ModalState{}, err
	}

	var modalState view.ModalState
	switch strings.ToLower(col.Name) {
	case "backlog":
		modalState = view.ModalState{Show: false}
		break
	case "todo":
		// create new issue
		if item.IssueID == -1 {
			_, err := h.itemRepo.CreateIssue("TobiasTheDanish", "go-track", item)
			if err != nil {
				return view.ModalState{}, err
			}
		}
		modalState = view.ModalState{Show: false}
		break
	case "in progress":
		// create branch for issue
		if item.BranchName == "" {
			branches, err := h.branchRepo.GetAll("TobiasTheDanish", "go-track")
			if err != nil {
				return view.ModalState{}, err
			}

			dropdownItems := make([]view.DropdownItem, len(branches), len(branches))

			for i, branch := range branches {
				dropdownItems[i] = view.DropdownItem{
					Value: branch.Sha,
					Name:  branch.Name,
				}
			}

			modalState = view.ModalState{
				Show:            true,
				Title:           fmt.Sprintf("Create branch for '%s'", item.Name),
				Body:            view.CreateBranchModalBody(dropdownItems...),
				Endpoint:        fmt.Sprintf("/project/%d/items/%d/branch", projID, item.Id),
				TargetElementID: "columns-container",
			}
		} else {
			modalState = view.ModalState{Show: false}
		}
		break
	case "ready for pull request":
		// create pr for branch
		if item.BranchName != "" {
			branches, err := h.branchRepo.GetAll("TobiasTheDanish", "go-track")
			if err != nil {
				return view.ModalState{}, err
			}

			dropdownItems := make([]view.DropdownItem, len(branches), len(branches))

			for i, branch := range branches {
				dropdownItems[i] = view.DropdownItem{
					Value: branch.Name,
					Name:  branch.Name,
				}
			}

			modalState = view.ModalState{
				Show:            true,
				Title:           fmt.Sprintf("Create pull request for branch '%s'", item.BranchName),
				Body:            view.CreatePRModalBody(item.BranchName, dropdownItems...),
				Endpoint:        fmt.Sprintf("/project/%d/items/%d/pr", projID, item.Id),
				TargetElementID: "columns-container",
			}

		} else {
			modalState = view.ModalState{Show: false}
		}
		break
	case "done":
		// close pr and issue
		if item.PullRequestNumber != -1 {
			title := fmt.Sprintf("Merge pull request #%d from TobiasTheDanish/%s", item.PullRequestNumber, item.BranchName)

			modalState = view.ModalState{
				Show:            true,
				Title:           fmt.Sprintf("Merge pull request for branch '%s'", item.BranchName),
				Body:            view.MergePRModalBody(title, item.BranchName, item.PullRequestNumber),
				Endpoint:        fmt.Sprintf("/project/%d/items/%d/merge", projID, item.Id),
				TargetElementID: "columns-container",
			}
		} else {
			modalState = view.ModalState{Show: false}
		}
		break
	}

	return modalState, nil
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

	item, err := h.columnRepo.AddItem(name, columnID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	col, err := h.columnRepo.Get(columnID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return view.ProjectItem(col.ProjectID, item).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) CreateBranchHandler(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	name := c.FormValue("branch-name")
	sha := c.FormValue("branch-sha")

	_, err = h.itemRepo.CreateBranch("TobiasTheDanish", "go-track", name, sha, itemID)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%d/columns", id))
}

func (h *Handler) CreatePRHandler(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	head := c.FormValue("head-branch")
	base := c.FormValue("base-branch")

	_, err = h.itemRepo.CreatePullRequest("TobiasTheDanish", "go-track", head, base, itemID)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%d/columns", id))
}

func (h *Handler) MergePRHandler(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	pullNumber, err := strconv.Atoi(c.FormValue("pull-number"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	title := c.FormValue("commit-title")
	message := c.FormValue("commit-message")
	deleteBranch := c.FormValue("delete-branch")

	_, err = h.itemRepo.MergePullRequest("TobiasTheDanish", "go-track", title, message, pullNumber, deleteBranch == "on", itemID)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/%d/columns", id))
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

	column, err := h.columnRepo.RemoveItem(itemID, columnID)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Could not delete item: %s", err.Error()))
	}

	return view.ProjectColumn(column).Render(c.Request().Context(), c.Response().Writer)
}
