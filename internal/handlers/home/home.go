package home

import (
	"net/http"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/home/templates"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	DB *common.DB
	//Logger *log.Logger
}

func NewHandler(db *common.DB) *Handler {
	return &Handler{
		DB: db,
		//Logger: log.New(os.Stderr, "home-handler: ", log.LstdFlags),
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		ui.NewEchoRoute(http.MethodGet, path, h.GetHomePage),
	})
}

func (h *Handler) GetHomePage(c echo.Context) *echo.HTTPError {
	t := templates.HomePage()
	err := t.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, "Home Page")
	}
	return nil
}
