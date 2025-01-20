package model

type Project struct {
	Id      int
	Name    string
	Columns []Column
}

type Column struct {
	Id        int
	Name      string
	ProjectID int
	Items     []Item
}

type Item struct {
	Id          int
	Name        string
	ColumnID    int
	ColumnOrder int
	IssueID     int64
	IssueNumber int
	IssueUrl    string
	BranchName  string
}
