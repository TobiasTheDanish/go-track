package web

import (
	"go-track/internal/auth"
	"log"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func (h *Handler) SignInHandler(c echo.Context) error {
	authUrl := h.gh.GetAuthUrl()

	return SignIn(templ.SafeURL(authUrl)).Render(c.Request().Context(), c.Response().Writer)
}

func (h *Handler) GithubAuthCallbackHandler(c echo.Context) error {
	code := c.QueryParam("code")

	log.Printf("Github oAuth code: %v\n", code)

	authUser, err := h.gh.AuthUserByCode(code)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	log.Printf("AuthUserRes: %v\n", authUser)

	user, err := h.gh.GetAuthorizedUser(authUser)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	log.Printf("AuthorizedUser: %v\n", user)

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
