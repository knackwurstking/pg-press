package editor

import (
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/env"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/handlers/editor/templates"
	"github.com/knackwurstking/pg-press/internal/shared"

	ui "github.com/knackwurstking/ui/ui-templ"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *common.DB
}

func NewHandler(r *common.DB) *Handler {
	return &Handler{
		registry: r,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	ui.RegisterEchoRoutes(e, env.ServerPathPrefix, []*ui.EchoRoute{
		ui.NewEchoRoute(http.MethodGet, path, h.GetEditorPage),
		ui.NewEchoRoute(http.MethodPost, path+"/save", h.PostSaveContent),
	})
}

func (h *Handler) GetEditorPage(c echo.Context) *echo.HTTPError {
	// Parse query parameters
	editorType := models.EditorType(c.QueryParam("type"))
	if editorType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "editor type is required")
	}

	var id int64
	if idParam := c.QueryParam("id"); idParam != "" {
		var err error
		id, err = strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid ID parameter")
		}
	}

	props := &templates.PageProps{
		Type:      editorType,
		ID:        int64(id),
		ReturnURL: templ.SafeURL(c.QueryParam("return_url")),
	}

	// Load existing content based on type
	if id > 0 {
		merr := h.loadExistingContent(props)
		if merr != nil {
			return merr.WrapEcho("load existing content")
		}
	}

	// Render the editor page
	page := templates.Page(props)
	err := page.Render(c.Request().Context(), c.Response())
	if err != nil {
		return errors.NewRenderError(err, fmt.Sprintf("Editor Page: %s", editorType))
	}

	return nil
}

func (h *Handler) PostSaveContent(c echo.Context) *echo.HTTPError {
	var (
		editorType = c.FormValue("type")
		idParam    = c.FormValue("id")
	)

	slog.Info("Save editor content", "type", editorType, "id", idParam)

	// Get user from context
	user, merr := shared.GetUserFromContext(c)
	if merr != nil {
		return merr.Echo()
	}

	// Parse form data
	var (
		title       = strings.TrimSpace(c.FormValue("title"))
		content     = strings.TrimSpace(c.FormValue("content"))
		useMarkdown = c.FormValue("use_markdown") == "on"
	)

	if editorType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "editor type is required")
	}

	if title == "" || content == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title and content are required")
	}

	var id int64
	if idParam != "" {
		var err error
		id, err = strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid ID parameter")
		}
	}

	// Handle attachments
	attachments, merr := h.processAttachments(c)
	if merr != nil {
		return merr.WrapEcho("process attachments")
	}

	// Save content based on type
	merr = h.saveContent(editorType, id, title, content, useMarkdown, attachments, user)
	if merr != nil {
		return merr.WrapEcho("save")
	}

	// Redirect back to return URL or appropriate page
	returnURL := c.FormValue("return_url")
	if returnURL != "" {
		url := templ.SafeURL(returnURL)
		merr = utils.RedirectTo(c, url)
		if merr != nil {
			return merr.WrapEcho("redirect to %#v", url)
		}
		return nil
	}

	// Default redirects based on type
	switch editorType {
	case "troublereport":
		url := utils.UrlTroubleReports(0, 0, 0).Page
		merr = utils.RedirectTo(c, url)
		if merr != nil {
			return merr.WrapEcho("redirect to %#v", url)
		}
		return nil
	default:
		url := utils.UrlHome().Page
		merr = utils.RedirectTo(c, url)
		if merr != nil {
			return merr.WrapEcho("redirect to %#v", url)
		}
		return nil
	}
}

// loadExistingContent loads existing content based on type and ID
func (h *Handler) loadExistingContent(props *templates.PageProps) *errors.MasterError {
	switch props.Type {
	case "troublereport":
		tr, merr := h.registry.TroubleReports.Get(models.TroubleReportID(props.ID))
		if merr != nil {
			return merr
		}
		props.Title = tr.Title
		props.Content = tr.Content
		props.UseMarkdown = tr.UseMarkdown

		// Load attachments
		loadedAttachments, merr := h.registry.TroubleReports.LoadAttachments(tr)
		if merr == nil {
			props.Attachments = loadedAttachments
		} else {
			slog.Error("Failed to load attachments for trouble report",
				"trouble_report_id", props.ID, "error", merr)
		}

	default:
		return errors.NewMasterError(
			fmt.Errorf("unsupported editor type: %s (only 'troublereport' is supported)", props.Type),
			http.StatusBadRequest,
		)
	}

	return nil
}

// saveContent saves content based on type
func (h *Handler) saveContent(editorType string, id int64, title, content string, useMarkdown bool, attachments []*models.Attachment, user *models.User) *errors.MasterError {
	switch editorType {
	case "troublereport":
		if id > 0 {
			trID := models.TroubleReportID(id)

			// Update existing trouble report
			tr, merr := h.registry.TroubleReports.Get(trID)
			if merr != nil {
				return merr
			}

			// Filter out existing and new attachments
			var existingAttachmentIDs []models.AttachmentID
			var newAttachments []*models.Attachment
			for _, a := range attachments {
				if a.GetID() > 0 {
					existingAttachmentIDs = append(existingAttachmentIDs, a.GetID())
				} else {
					newAttachments = append(newAttachments, a)
				}
			}

			tr.Title = title
			tr.Content = content
			tr.UseMarkdown = useMarkdown
			tr.LinkedAttachments = existingAttachmentIDs

			merr = h.registry.TroubleReports.UpdateWithAttachments(trID, tr, user, newAttachments...)
			if merr != nil {
				return merr
			}

			// Create feed entry
			feedTitle := "Problembericht aktualisiert"
			feedContent := fmt.Sprintf("Titel: %s", tr.Title)

			totalAttachments := len(existingAttachmentIDs) + len(newAttachments)
			if totalAttachments > 0 {
				feedContent += fmt.Sprintf("\nAnhänge: %d", len(attachments))
			}

			merr = h.registry.Feeds.Add(feedTitle, feedContent, user.TelegramID)
			if merr != nil {
				slog.Warn("Failed to create feed for trouble report update", "error", merr)
			}

		} else {
			// Create new trouble report
			tr := models.NewTroubleReport(title, content)
			tr.UseMarkdown = useMarkdown

			merr := h.registry.TroubleReports.AddWithAttachments(tr, user, attachments...)
			if merr != nil {
				return merr
			}

			// Create feed entry
			feedTitle := "Neuer Problembericht erstellt"
			feedContent := fmt.Sprintf("Titel: %s", tr.Title)

			if len(attachments) > 0 {
				feedContent += fmt.Sprintf("\nAnhänge: %d", len(attachments))
			}

			merr = h.registry.Feeds.Add(feedTitle, feedContent, user.TelegramID)
			if merr != nil {
				slog.Warn("Failed to create feed for trouble report creation", "error", merr)
			}
		}

	// Note: Notes are not supported in the editor

	default:
		return errors.NewMasterError(
			fmt.Errorf("unsupported editor type: %s (only 'troublereport' is supported)", editorType),
			http.StatusBadRequest,
		)
	}

	return nil
}

// processAttachments handles file uploads and existing attachments
func (h *Handler) processAttachments(c echo.Context) ([]*models.Attachment, *errors.MasterError) {
	var attachments []*models.Attachment

	// Handle new file uploads
	form, err := c.MultipartForm()
	if err != nil {
		// No multipart form is okay, just return empty attachments
		return attachments, nil
	}

	files := form.File["attachments"]
	for _, fileHeader := range files {
		if fileHeader.Size == 0 {
			continue
		}

		// Validate file size (max 10MB)
		if fileHeader.Size > 10*1024*1024 {
			return nil, errors.NewMasterError(
				fmt.Errorf("file %s is too large (max 10MB)", fileHeader.Filename),
				http.StatusBadRequest,
			)
		}

		// Validate file type (images only)
		if !strings.HasPrefix(fileHeader.Header.Get("Content-Type"), "image/") {
			return nil, errors.NewMasterError(
				fmt.Errorf("file %s is not an image", fileHeader.Filename),
				http.StatusBadRequest,
			)
		}

		attachment, merr := h.processFileUpload(fileHeader)
		if merr != nil {
			return nil, merr.Wrap("process file %s", fileHeader.Filename)
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

// processFileUpload processes a single file upload
func (h *Handler) processFileUpload(fileHeader *multipart.FileHeader) (*models.Attachment, *errors.MasterError) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.NewMasterError(
			fmt.Errorf("open file: %v", err), http.StatusInternalServerError,
		)
	}
	defer file.Close()

	// Read file data
	data := make([]byte, fileHeader.Size)
	_, err = file.Read(data)
	if err != nil {
		return nil, errors.NewMasterError(
			fmt.Errorf("read file: %v", err), http.StatusInternalServerError,
		)
	}

	// Get MIME type
	mimeType := h.getMimeTypeFromFilename(fileHeader.Filename)

	// Create attachment - need to check the actual constructor for Attachment model
	attachment := &models.Attachment{
		Data:     data,
		MimeType: mimeType,
	}

	return attachment, nil
}

// getMimeTypeFromFilename determines MIME type from filename
func (h *Handler) getMimeTypeFromFilename(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	case strings.HasSuffix(lower, ".svg"):
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
