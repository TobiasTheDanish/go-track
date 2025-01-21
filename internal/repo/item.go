package repo

import (
	"errors"
	"go-track/internal/db"
	"go-track/internal/github"
	"go-track/internal/model"
	"strings"
)

type ItemRepository interface {
	Move(projID, itemID int, dir string) (model.Item, error)
	Get(itemID int) (model.Item, error)
	CreateIssue(owner, repo string, item model.Item) (model.Item, error)
	CreateBranch(owner, repo, branchName, branchSha string, itemID int) (model.Item, error)
	CreatePullRequest(owner, repo, headBranch, baseBranch string, itemID int) (model.Item, error)
	MergePullRequest(owner, repo, title, message string, pullNumber int, deleteBranch bool, itemID int) (model.Item, error)
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

func (r *itemRepo) CreateBranch(owner, repo, branchName, branchSha string, itemID int) (model.Item, error) {

	item, err := r.Get(itemID)
	if err != nil {
		return model.Item{}, err
	}

	branch, err := r.gh.CreateBranch("TobiasTheDanish", "go-track", branchName, branchSha)
	if err != nil {
		return model.Item{}, err
	}

	item.BranchName = branch.Name

	return r.db.UpdateItem(itemID, item)
}

func (r *itemRepo) CreatePullRequest(owner, repo, headBranch, baseBranch string, itemID int) (model.Item, error) {
	item, err := r.Get(itemID)
	if err != nil {
		return model.Item{}, err
	}

	pr, err := r.gh.CreatePullRequest(owner, repo, headBranch, baseBranch, item.IssueNumber)
	if err != nil {
		return model.Item{}, err
	}

	item.IssueID = -1
	item.IssueNumber = -1
	item.IssueUrl = ""
	item.PullRequestID = pr.Id
	item.PullRequestNumber = pr.Number

	return r.db.UpdateItem(itemID, item)
}

func (r *itemRepo) MergePullRequest(owner, repo, title, message string, pullNumber int, deleteBranch bool, itemID int) (model.Item, error) {
	item, err := r.Get(itemID)

	_, err = r.gh.MergePullRequest("TobiasTheDanish", "go-track", title, message, pullNumber)
	if err != nil {
		return model.Item{}, err
	}

	if deleteBranch {
		err = r.gh.DeleteBranch("TobiasTheDanish", "go-track", item.BranchName)
		if err != nil {
			return model.Item{}, err
		}
		item.BranchName = ""
	}

	item.PullRequestID = -1
	item.PullRequestNumber = -1

	return r.db.UpdateItem(itemID, item)
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
