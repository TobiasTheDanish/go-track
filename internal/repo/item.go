package repo

import (
	"errors"
	"fmt"
	view "go-track/cmd/web/view"
	"go-track/internal/db"
	"go-track/internal/github"
	"go-track/internal/model"
	"strings"
)

type ItemRepository interface {
	Move(projID, itemID int, dir string) (model.Item, error)
	Get(itemID int) (model.Item, error)
	CreateIssue(owner, repo string, item model.Item) (model.Item, error)
}

type itemRepo struct {
	db db.DatabaseFacade
	gh github.GithubService
}

func NewItemRepo(db db.DatabaseFacade, gh github.GithubService) ItemRepository {
	return &itemRepo{
		db: db,
		gh: gh,
	}
}

func (r *itemRepo) Get(itemID int) (model.Item, error) {
	return r.db.GetItem(itemID)
}

func (r *itemRepo) CreateIssue(owner, repo string, item model.Item) (model.Item, error) {
	issue, err := r.gh.CreateIssue(owner, repo, item.Name)
	if err != nil {
		return model.Item{}, err
	}

	item.IssueID = issue.Id
	item.IssueNumber = issue.Number
	item.IssueUrl = issue.HtmlUrl

	return r.db.UpdateItem(item.Id, item)
}

func (r *itemRepo) Move(projID, itemID int, dir string) (model.Item, error) {
	item, err := r.db.GetItem(itemID)
	if err != nil {
		return model.Item{}, err
	}

	switch strings.ToLower(dir) {
	case "left":
		return r.moveItemLeft(projID, item)
	case "right":
		return r.moveItemRight(projID, item)
	case "up":
		return r.moveItemUp(item)
	case "down":
		return r.moveItemDown(item)

	default:
		return model.Item{}, errors.New("Invalid move direction")
	}
}

func (r *itemRepo) itemEnter(projID, colID int, item model.Item) (view.ModalState, error) {
	col, err := r.db.GetColumn(colID)
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
			issue, err := r.gh.CreateIssue("TobiasTheDanish", "go-track", item.Name)
			if err != nil {
				return view.ModalState{}, err
			}

			item.IssueID = issue.Id
			item.IssueNumber = issue.Number
			item.IssueUrl = issue.HtmlUrl

			_, err = r.db.UpdateItem(item.Id, item)
			if err != nil {
				return view.ModalState{}, err
			}
		}
		modalState = view.ModalState{Show: false}
		break
	case "in progress":
		// create branch for issue
		if item.BranchName == "" {
			branches, err := r.gh.GetBranches("TobiasTheDanish", "go-track")
			if err != nil {
				return view.ModalState{}, err
			}

			dropdownItems := make([]view.DropdownItem, len(branches), len(branches))

			for i, branch := range branches {
				dropdownItems[i] = view.DropdownItem{
					Value: branch.Commit.Sha,
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
			branches, err := r.gh.GetBranches("TobiasTheDanish", "go-track")
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

func (h *itemRepo) moveItemDown(item model.Item) (model.Item, error) {
	col, err := h.db.GetColumn(item.ColumnID)
	if err != nil {
		return model.Item{}, err
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

	newItem, err := h.db.UpdateItem(item.Id, item)
	if err != nil {
		return model.Item{}, err
	}
	_, err = h.db.UpdateItem(itemToSwap.Id, itemToSwap)
	if err != nil {
		return model.Item{}, err
	}
	return newItem, nil
}

func (h *itemRepo) moveItemUp(item model.Item) (model.Item, error) {
	col, err := h.db.GetColumn(item.ColumnID)
	if err != nil {
		return model.Item{}, err
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

	newItem, err := h.db.UpdateItem(item.Id, item)
	if err != nil {
		return model.Item{}, err
	}
	_, err = h.db.UpdateItem(itemToSwap.Id, itemToSwap)
	if err != nil {
		return model.Item{}, err
	}
	return newItem, nil
}

func (h *itemRepo) moveItemRight(projID int, item model.Item) (model.Item, error) {
	proj, err := h.db.GetProject(projID)
	if err != nil {
		return model.Item{}, err
	}

	colIndex := len(proj.Columns)
	for i, col := range proj.Columns {
		if col.Id == item.ColumnID {
			colIndex = i
			break
		}
	}

	if colIndex >= len(proj.Columns)+1 {
		return model.Item{}, errors.New("Could not move item right")
	}

	newCol := proj.Columns[colIndex+1]

	colOrder, err := h.db.GetNextItemColumnOrder(newCol.Id)
	if err != nil {
		return model.Item{}, err
	}

	item.ColumnID = newCol.Id
	item.ColumnOrder = colOrder

	return h.db.UpdateItem(item.Id, item)
}

func (h *itemRepo) moveItemLeft(projID int, item model.Item) (model.Item, error) {
	proj, err := h.db.GetProject(projID)
	if err != nil {
		return model.Item{}, err
	}

	colIndex := -1
	for i, col := range proj.Columns {
		if col.Id == item.ColumnID {
			colIndex = i
			break
		}
	}

	if colIndex <= 0 {
		return model.Item{}, errors.New("Could not move item left")
	}

	newCol := proj.Columns[colIndex-1]

	colOrder, err := h.db.GetNextItemColumnOrder(newCol.Id)
	if err != nil {
		return model.Item{}, err
	}

	item.ColumnID = newCol.Id
	item.ColumnOrder = colOrder

	return h.db.UpdateItem(item.Id, item)
}
