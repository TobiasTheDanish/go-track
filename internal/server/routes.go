package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go-track/cmd/web"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	fileServer := http.FileServer(http.FS(web.Files))
	e.GET("/assets/*", echo.WrapHandler(fileServer))

	e.GET("/:id", s.webHandler.ProjectPageHandler)

	e.POST("/columns/items", s.webHandler.ProjectItemHandler)
	e.POST("/project/:id/items/:itemID/move", s.webHandler.MoveProjectItemHandler)

	e.DELETE("/columns/:colID/items/:itemID", s.webHandler.DeleteProjectItemHandler)

	e.GET("/api", s.HelloWorldHandler)

	return e
}

func (s *Server) HelloWorldHandler(c echo.Context) error {
	resp := map[string]string{
		"message": "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}
