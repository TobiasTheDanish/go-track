package web

import "go-track/internal/db"

type Handler struct {
	db db.DatabaseFacade
}

func NewHandler(db db.DatabaseFacade) *Handler {
	return &Handler{
		db: db,
	}
}
