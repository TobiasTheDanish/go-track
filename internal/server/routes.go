package server

import (
	"log"
	"net/http"

	"go-track/cmd/web"
	"go-track/internal/auth"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	fileServer := http.FileServer(http.FS(web.Files))
	e.GET("/assets/*", echo.WrapHandler(fileServer))

	e.GET("/", func(c echo.Context) error {
		jwtCookie, err := c.Request().Cookie("authSession")
		if err != nil {
			log.Printf("Could not get cookie with name authSession: %s\n", err)
			return c.Redirect(http.StatusTemporaryRedirect, "/sign-in?error=Authorization%20failed")
		}

		_, err = auth.ParseAuthJWT(jwtCookie.Value)
		if err != nil {
			log.Printf("Error when parsing jwt: %s\n", err)
			return c.Redirect(http.StatusTemporaryRedirect, "/sign-in")
		}

		return c.Redirect(http.StatusTemporaryRedirect, "/1")
	})

	e.GET("/sign-in", s.webHandler.SignInHandler)
	e.GET("/auth/callback", s.webHandler.GithubAuthCallbackHandler)

	e.GET("/:id", s.webHandler.ProjectPageHandler)
	e.GET("/:id/columns", s.webHandler.ProjectColumnsHandler)

	e.POST("/columns/items", s.webHandler.ProjectItemHandler)
	e.POST("/project/:id/items/:itemID/move", s.webHandler.MoveProjectItemHandler)
	e.POST("/project/:id/items/:itemID/branch", s.webHandler.CreateBranchHandler)

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
