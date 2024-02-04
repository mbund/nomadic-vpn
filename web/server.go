package web

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func Run(port uint16, accessToken string) {
	e := echo.New()
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

	// listen for api calls with the access token as bearer token
	// e.Use(middleware.JWTWithConfig(middleware.JWTConfig{
	// 	SigningKey: []byte(accessToken),
	// }))
	// e.GET("/api", func(c echo.Context) error {
	// 	return c.String(http.StatusOK, "api")
	// })

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}
