package web

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/mbund/nomadic-vpn/core"
	"github.com/mbund/nomadic-vpn/db"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

func Run(port uint16, domain string, accessToken string) {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		x := Render(c, http.StatusOK, Connections())
		fmt.Println(x)
		return x
	})

	e.GET("/settings", func(c echo.Context) error {
		return Render(c, http.StatusOK, Settings())
	})

	e.POST("/api/initialize", func(c echo.Context) error {
		fmt.Println("Initializing")
		fmt.Println(c)
		fmt.Println(c.Request())
		fmt.Println(c.Request().Form)
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader != fmt.Sprintf("Bearer %s", accessToken) {
			return c.JSON(http.StatusUnauthorized, nil)
		}

		db.InitDb()
		db.SetVultrApiKey(c.Request().FormValue("vultrApiKey"))
		db.SetDuckDnsToken(c.Request().FormValue("duckDnsToken"))
		db.SetDuckDnsDomain(c.Request().FormValue("duckDnsDomain"))

		clientConfig, err := core.Bootstrap(domain)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}

		return c.JSON(http.StatusOK, db.ApiInitializeResponse{
			WireguardConf: clientConfig,
		})
	})

	e.GET("/healthz", func(c echo.Context) error {
		var result struct {
			Status string `json:"status"`
		}
		result.Status = "ok"
		return c.JSON(http.StatusOK, result)
	})

	if os.Getenv("NOMADIC_VPN_DEBUG") == "true" {
		e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
	} else {
		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(domain)
		e.AutoTLSManager.Client = &acme.Client{
			DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory",
		}
		e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
		e.Logger.Fatal(e.StartAutoTLS(fmt.Sprintf(":%d", port)))
	}
}

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}
