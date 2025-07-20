package nav

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/knackwurstking/pg-vis/routes/constants"
	"github.com/knackwurstking/pg-vis/routes/internal/utils"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	db               *pgvis.DB
	serverPathPrefix string
	templates        fs.FS
}

func NewHandler(db *pgvis.DB, serverPathPrefix string, templates fs.FS) *Handler {
	return &Handler{
		db:               db,
		serverPathPrefix: serverPathPrefix,
		templates:        templates,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET(h.serverPathPrefix+"/nav/feed-counter", h.handleGetFeedCounter)
}

func (h *Handler) handleGetFeedCounter(c echo.Context) error {
	data := &FeedCounterTemplateData{}

	feeds, err := h.db.Feeds.ListRange(0, 100)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusInternalServerError,
			fmt.Errorf("list feeds: %w", err),
		)
	}

	user, herr := utils.GetUserFromContext(c)
	if herr != nil {
		return herr
	}

	for _, feed := range feeds {
		if feed.ID > user.LastFeed {
			data.Count++
		} else {
			break
		}
	}

	return utils.HandleTemplate(c, data,
		h.templates,
		[]string{
			constants.FeedCounterComponentTemplatePath,
		},
	)
}

type FeedCounterTemplateData struct {
	Count int
}
