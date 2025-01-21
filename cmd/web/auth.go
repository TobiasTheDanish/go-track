package web

import (
	view "go-track/cmd/web/view"
	"go-track/internal/auth"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func (h *Handler) SignInHandler(c echo.Context) error {
	authUrl := h.authRepo.GetAuthUrl()

	return view.SignIn(templ.SafeURL(authUrl)).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) GithubAuthCallbackHandler(c echo.Context) error {
	code := c.QueryParam("code")

	user, err := h.authRepo.AuthorizeUser(code)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	jwtString, err := auth.SignAuthSession(auth.AuthSession{
		Username: user.Username,
		OrgUrl:   user.OrgUrl,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	http.SetCookie(c.Response().Writer, &http.Cookie{
		Name:   "authSession",
		Value:  jwtString,
		MaxAge: int(24 * time.Hour),
		Path:   "/",
	})

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
