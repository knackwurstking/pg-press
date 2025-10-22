package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/knackwurstking/pgpress/components"
	"github.com/knackwurstking/pgpress/logger"
	"github.com/knackwurstking/pgpress/models"
	"github.com/knackwurstking/pgpress/services"
	"github.com/knackwurstking/pgpress/utils"
	"github.com/labstack/echo/v4"
)

type Editor struct {
	*Base
}

func NewEditor(db *services.Registry) *Editor {
	return &Editor{
		Base: NewBase(db, logger.NewComponentLogger("Editor")),
	}
}

func (h *Editor) RegisterRoutes(e *echo.Echo) {
	utils.RegisterEchoRoutes(e, []*utils.EchoRoute{
		utils.NewEchoRoute(http.MethodGet, "/editor", h.GetEditorPage),
		utils.NewEchoRoute(http.MethodPost, "/editor/save", h.PostSaveContent),
	})
}

func (h *Editor) GetEditorPage(c echo.Context) error {
	// Parse query parameters
	editorType := c.QueryParam("type")
	idParam := c.QueryParam("id")
	returnURL := c.QueryParam("return_url")

	if editorType == "" {
		return HandleBadRequest(nil, "editor type is required")
	}

	var id int64
	var err error
	if idParam != "" {
		id, err = strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			return HandleBadRequest(err, "invalid ID parameter")
		}
	}

	options := &components.PageEditorOptions{
		Type:      editorType,
		ID:        id,
		ReturnURL: returnURL,
	}

	// Load existing content based on type
	if id > 0 {
		err := h.loadExistingContent(options)
		if err != nil {
			return HandleError(err, "failed to load existing content")
		}
	}

	// Render the editor page
	page := components.PageEditor(options)
	if err := page.Render(c.Request().Context(), c.Response()); err != nil {
		return HandleError(err, "failed to render editor page")
	}

	return nil
}

func (h *Editor) PostSaveContent(c echo.Context) error {
	// Get user from context
	user, err := GetUserFromContext(c)
	if err != nil {
		return HandleError(err, "failed to get user from context")
	}

	// Parse form data
	editorType := c.FormValue("type")
	idParam := c.FormValue("id")
	title := strings.TrimSpace(c.FormValue("title"))
	content := strings.TrimSpace(c.FormValue("content"))
	useMarkdownStr := c.FormValue("use_markdown")
	returnURL := c.FormValue("return_url")

	if editorType == "" {
		return HandleBadRequest(nil, "editor type is required")
	}

	if title == "" || content == "" {
		return HandleBadRequest(nil, "title and content are required")
	}

	useMarkdown := useMarkdownStr == "on"

	var id int64
	if idParam != "" {
		id, err = strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			return HandleBadRequest(err, "invalid ID parameter")
		}
	}

	// Handle attachments
	attachments, err := h.processAttachments(c)
	if err != nil {
		return HandleBadRequest(err, "failed to process attachments")
	}

	// Save content based on type
	err = h.saveContent(editorType, id, title, content, useMarkdown, attachments, user)
	if err != nil {
		return HandleError(err, "failed to save content")
	}

	// Redirect back to return URL or appropriate page
	if returnURL != "" {
		return c.Redirect(http.StatusSeeOther, returnURL)
	}

	// Default redirects based on type
	switch editorType {
	case "troublereport":
		return c.Redirect(http.StatusSeeOther, "/trouble-reports")
	default:
		return c.Redirect(http.StatusSeeOther, "/")
	}
}

// loadExistingContent loads existing content based on type and ID
func (h *Editor) loadExistingContent(options *components.PageEditorOptions) error {
	switch options.Type {
	case "troublereport":
		tr, err := h.Registry.TroubleReports.Get(options.ID)
		if err != nil {
			return fmt.Errorf("failed to get trouble report: %w", err)
		}
		options.Title = tr.Title
		options.Content = tr.Content
		options.UseMarkdown = tr.UseMarkdown

		// Load attachments
		loadedAttachments, err := h.Registry.TroubleReports.LoadAttachments(tr)
		if err == nil {
			options.Attachments = loadedAttachments
		} else {
			h.Log.Error("Failed to load attachments for trouble report %d: %v", options.ID, err)
		}

	// Note: Notes are not supported in the editor as they have a different structure
	// (Level-based rather than title/content based)

	default:
		return fmt.Errorf("unsupported editor type: %s (only 'troublereport' is supported)", options.Type)
	}

	return nil
}

// saveContent saves content based on type
func (h *Editor) saveContent(editorType string, id int64, title, content string, useMarkdown bool, attachments []*models.Attachment, user *models.User) error {
	switch editorType {
	case "troublereport":
		if id > 0 {
			// Update existing trouble report
			tr, err := h.Registry.TroubleReports.Get(id)
			if err != nil {
				return fmt.Errorf("failed to get trouble report: %w", err)
			}

			// Filter out existing and new attachments
			var existingAttachmentIDs []int64
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

			err = h.Registry.TroubleReports.UpdateWithAttachments(id, tr, user, newAttachments...)
			if err != nil {
				return fmt.Errorf("failed to update trouble report: %w", err)
			}

			// Create feed entry
			feedTitle := "Problembericht aktualisiert"
			feedContent := fmt.Sprintf("Titel: %s", tr.Title)
			totalAttachments := len(existingAttachmentIDs) + len(newAttachments)
			if totalAttachments > 0 {
				feedContent += fmt.Sprintf("\nAnhänge: %d", totalAttachments)
			}
			feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
			if err := h.Registry.Feeds.Add(feed); err != nil {
				h.Log.Error("Failed to create feed for trouble report update: %v", err)
			}

		} else {
			// Create new trouble report
			tr := models.NewTroubleReport(title, content)
			tr.UseMarkdown = useMarkdown

			err := h.Registry.TroubleReports.AddWithAttachments(tr, user, attachments...)
			if err != nil {
				return fmt.Errorf("failed to add trouble report: %w", err)
			}

			// Create feed entry
			feedTitle := "Neuer Problembericht erstellt"
			feedContent := fmt.Sprintf("Titel: %s", tr.Title)
			if len(attachments) > 0 {
				feedContent += fmt.Sprintf("\nAnhänge: %d", len(attachments))
			}
			feed := models.NewFeed(feedTitle, feedContent, user.TelegramID)
			if err := h.Registry.Feeds.Add(feed); err != nil {
				h.Log.Error("Failed to create feed for trouble report creation: %v", err)
			}
		}

	// Note: Notes are not supported in the editor

	default:
		return fmt.Errorf("unsupported editor type: %s (only 'troublereport' is supported)", editorType)
	}

	return nil
}

// processAttachments handles file uploads and existing attachments
func (h *Editor) processAttachments(c echo.Context) ([]*models.Attachment, error) {
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
			return nil, fmt.Errorf("failed to process file %s: %w", fileHeader.Filename, err)
		}

		attachments = append(attachments, attachment)
	}

	return attachments, nil
}

// processFileUpload processes a single file upload
func (h *Editor) processFileUpload(fileHeader *multipart.FileHeader) (*models.Attachment, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read file data
	data := make([]byte, fileHeader.Size)
	_, err = file.Read(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
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
func (h *Editor) getMimeTypeFromFilename(filename string) string {
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
