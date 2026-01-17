package editor

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func Save(c echo.Context) *echo.HTTPError {
	var (
		editorType  = shared.EditorType(c.FormValue("type"))
		vID         = c.FormValue("id")
		title       = strings.TrimSpace(c.FormValue("title"))
		content     = strings.TrimSpace(c.FormValue("content"))
		useMarkdown = c.FormValue("use_markdown") == "on"
	)

	log.Info("Save editor content with type %s and ID %s", editorType, vID)

	if editorType == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "editor type is required")
	}

	var id int64
	if vID != "" {
		var err error
		if id, err = strconv.ParseInt(vID, 10, 64); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid ID parameter")
		}
	}

	if title == "" || content == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title and content are required")
	}

	// Process existing attachments removal
	var existingAttachmentsToRemove []string
	if vExistingAttachmentsRemoval := c.FormValue("existing_attachments_removal"); vExistingAttachmentsRemoval != "" {
		existingAttachmentsToRemove = strings.Split(vExistingAttachmentsRemoval, ",")
	}

	// Process new file uploads
	attachments, err := processAttachments(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to process attachments: "+err.Error())
	}

	v, _ := c.FormParams()
	log.Debug("Form values: existingAttachmentsToRemove=%#v", existingAttachmentsToRemove)
	log.Debug("Form values: attachments=%#v", attachments)
	log.Debug("Form values: v=%#v", v)

	switch editorType {
	case shared.EditorTypeTroubleReport:
		tr, merr := db.GetTroubleReport(shared.EntityID(id))
		if merr != nil && !merr.IsNotFoundError() {
			return merr.Echo()
		}
		if tr == nil {
			tr = &shared.TroubleReport{
				Title:       title,
				Content:     content,
				UseMarkdown: useMarkdown,
			}
		} else {
			tr.Title = title
			tr.Content = content
			tr.UseMarkdown = useMarkdown
		}

		tr.LinkedAttachments = attachments

		if merr != nil && merr.IsNotFoundError() {
			if merr = db.AddTroubleReport(tr); merr != nil {
				return merr.Echo()
			}
		} else {
			if merr = db.UpdateTroubleReport(tr); merr != nil {
				return merr.Echo()
			}
		}
	}

	return handleRedirect(c, editorType)
}

func processAttachments(c echo.Context) ([]string, error) {
	var attachments []string

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
			return nil, errors.NewValidationError(
				"file %s is too large (max 10MB)", fileHeader.Filename,
			).HTTPError()
		}

		// Validate file type (images only)
		if !strings.HasPrefix(fileHeader.Header.Get("Content-Type"), "image/") {
			return nil, errors.NewValidationError(
				"file %s is not an image", fileHeader.Filename,
			).HTTPError()
		}

		file, err := fileHeader.Open()
		if err != nil {
			return nil, errors.NewValidationError("open file: %v", err).HTTPError()
		}
		defer file.Close()

		// Read file data
		data := make([]byte, fileHeader.Size)
		_, err = file.Read(data)
		if err != nil {
			return nil, errors.NewValidationError("read file: %v", err).HTTPError()
		}

		// Generate a unique filename for this and add to attachments
		fileName := fmt.Sprintf("%s%d%s",
			time.Now().Format("20060102150405"),
			uuid.New().ID(),
			strings.ToLower(filepath.Ext(fileHeader.Filename)))

		// TODO: Implement local file storage
		// attachmentPath := fmt.Sprintf("%s/attachment_%d.%s",
		// 	env.ServerPathImages, time.Now().Unix(), getFileExtension(attachment.MimeType))
		// err := os.WriteFile(attachmentPath, attachment.Data, 0644)
		// if err != nil {
		// 	log.Error("Failed to save attachment: %v", err)
		// 	continue
		// }

		attachments = append(attachments, fileName)
	}

	return attachments, nil
}

func handleRedirect(c echo.Context, editorType shared.EditorType) *echo.HTTPError {
	// Redirect back to return URL or appropriate page
	returnURL := c.FormValue("return_url")
	if returnURL != "" {
		url := templ.SafeURL(returnURL)
		if merr := utils.RedirectTo(c, url); merr != nil {
			return merr.WrapEcho("redirect to %#v", url)
		}
		return nil
	}

	// Default redirects based on type
	switch editorType {
	case shared.EditorTypeTroubleReport:
		url := urlb.TroubleReports()
		if merr := utils.RedirectTo(c, url); merr != nil {
			return merr.WrapEcho("redirect to %#v", url)
		}
		return nil

	default:
		url := urlb.Home()
		if merr := utils.RedirectTo(c, url); merr != nil {
			return merr.WrapEcho("redirect to %#v", url)
		}
		return nil
	}
}
