package model

type Project struct {
	Id      int
	Name    string
	Columns []Column
}

type Column struct {
	Id    int
	Name  string
	Items []Item
}

type Item struct {
	Id       int
	Name     string
	ColumnID int
}
