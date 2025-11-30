package editor

import (
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/errors"
	"github.com/knackwurstking/pg-press/handlers/editor/templates"
	"github.com/knackwurstking/pg-press/models"
	"github.com/knackwurstking/pg-press/services"
	"github.com/knackwurstking/pg-press/utils"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	registry *services.Registry
}

func NewHandler(r *services.Registry) *Handler {
	return &Handler{
		registry: r,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo, path string) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		utils.NewEchoRoute(http.MethodGet, path, h.GetEditorPage),
		utils.NewEchoRoute(http.MethodPost, path+"/save", h.PostSaveContent),
	})
}

func (h *Handler) GetEditorPage(c echo.Context) error {
	// Parse query parameters
	editorType := c.QueryParam("type")
	idParam := c.QueryParam("id")
	returnURL := templ.SafeURL(c.QueryParam("return_url"))

	if editorType == "" {
		return errors.BadRequest(nil, "editor type is required")
	}

	var id int64
	if idParam != "" {
		var err error
		id, err = strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			return errors.BadRequest(err, "invalid ID parameter")
		}
	}

	options := &templates.EditorPageParams{
		Type:      editorType,
		ID:        int64(id),
		ReturnURL: returnURL,
	}

	// Load existing content based on type
	if id > 0 {
		err := h.loadExistingContent(options)
		if err != nil {
			return errors.Handler(err, "load existing content")
		}
	}

	// Render the editor page
	page := templates.EditorPage(options)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return errors.Handler(err, "render editor page")
	}

	return nil
}

func (h *Handler) PostSaveContent(c echo.Context) error {
	// Get user from context
	user, eerr := utils.GetUserFromContext(c)
	if eerr != nil {
		return eerr
	}

	// Parse form data
	editorType := c.FormValue("type")
	idParam := c.FormValue("id")
	title := strings.TrimSpace(c.FormValue("title"))
	content := strings.TrimSpace(c.FormValue("content"))
	useMarkdownStr := c.FormValue("use_markdown")
	returnURL := c.FormValue("return_url")

	if editorType == "" {
		return errors.BadRequest(nil, "editor type is required")
	}

	if title == "" || content == "" {
		return errors.BadRequest(nil, "title and content are required")
	}

	useMarkdown := useMarkdownStr == "on"

	var id int64
	if idParam != "" {
		var err error
		id, err = strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			return errors.BadRequest(err, "invalid ID parameter")
		}
	}

	// Handle attachments
	attachments, err := h.processAttachments(c)
	if err != nil {
		return errors.BadRequest(err, "process attachments")
	}

	// Save content based on type
	err = h.saveContent(editorType, id, title, content, useMarkdown, attachments, user)
	if err != nil {
		return errors.Handler(err, "save content")
	}

	// Redirect back to return URL or appropriate page
	if returnURL != "" {
		return utils.RedirectTo(c, templ.SafeURL(returnURL))
	}

	// Default redirects based on type
	switch editorType {
	case "troublereport":
		return utils.RedirectTo(c, utils.UrlTroubleReports(0, 0, 0).Page)
	default:
		return utils.RedirectTo(c, utils.UrlHome().Page)
	}
}

// loadExistingContent loads existing content based on type and ID
func (h *Handler) loadExistingContent(options *templates.EditorPageParams) error {
	switch options.Type {
	case "troublereport":
		tr, err := h.registry.TroubleReports.Get(models.TroubleReportID(options.ID))
		if err != nil {
			return fmt.Errorf("get trouble report: %v", err)
		}
		options.Title = tr.Title
		options.Content = tr.Content
		options.UseMarkdown = tr.UseMarkdown

		// Load attachments
		loadedAttachments, err := h.registry.TroubleReports.LoadAttachments(tr)
		if err == nil {
			options.Attachments = loadedAttachments
		} else {
			slog.Error("Failed to load attachments for trouble report",
				"trouble_report_id", options.ID, "error", err)
		}

	// Note: Notes are not supported in the editor as they have a different structure
	// (Level-based rather than title/content based)

	default:
		return fmt.Errorf("unsupported editor type: %s (only 'troublereport' is supported)", options.Type)
	}

	return nil
}

// saveContent saves content based on type
func (h *Handler) saveContent(editorType string, id int64, title, content string, useMarkdown bool, attachments []*models.Attachment, user *models.User) error {
	switch editorType {
	case "troublereport":
		if id > 0 {
			trID := models.TroubleReportID(id)

			// Update existing trouble report
			tr, err := h.registry.TroubleReports.Get(trID)
			if err != nil {
				return fmt.Errorf("get trouble report: %v", err)
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

			err = h.registry.TroubleReports.UpdateWithAttachments(trID, tr, user, newAttachments...)
			if err != nil {
				return fmt.Errorf("update trouble report: %v", err)
			}

			// Create feed entry
			feedTitle := "Problembericht aktualisiert"
			feedContent := fmt.Sprintf("Titel: %s", tr.Title)
			totalAttachments := len(existingAttachmentIDs) + len(newAttachments)
			if totalAttachments > 0 {
				feedContent += fmt.Sprintf("\nAnhänge: %d", totalAttachments)
			}
			feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
			if err := h.registry.Feeds.Add(feed); err != nil {
				slog.Error("Failed to create feed for trouble report update", "error", err)
			}

		} else {
			// Create new trouble report
			tr := models.NewTroubleReport(title, content)
			tr.UseMarkdown = useMarkdown

			err := h.registry.TroubleReports.AddWithAttachments(tr, user, attachments...)
			if err != nil {
				return fmt.Errorf("add trouble report: %v", err)
			}

			// Create feed entry
			feedTitle := "Neuer Problembericht erstellt"
			feedContent := fmt.Sprintf("Titel: %s", tr.Title)
			if len(attachments) > 0 {
				feedContent += fmt.Sprintf("\nAnhänge: %d", len(attachments))
			}
			feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
			if err := h.registry.Feeds.Add(feed); err != nil {
				slog.Error("Failed to create feed for trouble report creation", "error", err)
			}
		}

	// Note: Notes are not supported in the editor

	default:
		return fmt.Errorf("unsupported editor type: %s (only 'troublereport' is supported)", editorType)
	}

	return nil
}

// processAttachments handles file uploads and existing attachments
func (h *Handler) processAttachments(c echo.Context) ([]*models.Attachment, error) {
	var attachments []*models.Attachment

	// Handle existing attachments (for updates)
	// existingAttachmentsRemoval := c.FormValue("existing_attachments_removal")
	// This would need to be implemented based on the specific logic for handling existing attachments

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
			return nil, fmt.Errorf("file %s is too large (max 10MB)", fileHeader.Filename)
		}

		// Validate file type (images only)
		if !strings.HasPrefix(fileHeader.Header.Get("Content-Type"), "image/") {
			return nil, fmt.Errorf("file %s is not an image", fileHeader.Filename)
		}

		attachment, err := h.processFileUpload(fileHeader)
		if err != nil {
			return nil, fmt.Errorf("process file %s: %v", fileHeader.Filename, err)
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

// processFileUpload processes a single file upload
func (h *Handler) processFileUpload(fileHeader *multipart.FileHeader) (*models.Attachment, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("open file: %v", err)
	}
	defer file.Close()

	// Read file data
	data := make([]byte, fileHeader.Size)
	_, err = file.Read(data)
	if err != nil {
		return nil, fmt.Errorf("read file: %v", err)
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
