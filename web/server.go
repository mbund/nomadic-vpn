package web

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/acme/autocert"
)

func Run(port uint16, domain string, accessToken string) {
	e := echo.New()

	e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(domain)
	e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")

	e.GET("/", func(c echo.Context) error {
		return Render(c, http.StatusOK, Home("there!"))
	})
	e.GET("/healthz", func(c echo.Context) error {
		var result struct {
			Status string `json:"status"`
		}
		result.Status = "ok"
		return c.JSON(http.StatusOK, result)
	})

	e.Logger.Fatal(e.StartAutoTLS(fmt.Sprintf(":%d", port)))
}

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}
