package editor

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/knackwurstking/pg-press/internal/db"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/knackwurstking/pg-press/internal/urlb"
	"github.com/knackwurstking/pg-press/internal/utils"
	"github.com/labstack/echo/v4"
)

func Save(c echo.Context) *echo.HTTPError {
	var (
		editorType = c.FormValue("type")
		idParam    = c.FormValue("id")
	)

	log.Info("Save editor content with type %s and ID %s", editorType, idParam)

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

	// Process existing attachments removal
	var existingAttachmentsToRemove []string
	if vExistingAttachmentsRemoval := c.FormValue("existing_attachments_removal"); vExistingAttachmentsRemoval != "" {
		existingAttachmentsToRemove = strings.Split(vExistingAttachmentsRemoval, ",")
	}

	v, _ := c.FormParams()
	log.Debug("Form values: existingAttachmentsToRemove=%#v, %#v", existingAttachmentsToRemove, v)

	// Process new file uploads
	attachments, err := processAttachments(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to process attachments: "+err.Error())
	}

	var id int64
	if idParam != "" {
		var err error
		if id, err = strconv.ParseInt(idParam, 10, 64); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid ID parameter")
		}
	}

	switch editorType {
	case "troublereport":
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

		// Handle attachments for trouble reports
		if len(attachments) > 0 {
			for _, attachment := range attachments {
				// TODO: Store attachment locally at SERVER_PATH_IMAGES
				// For now, just log the attachment info
				log.Info("Processing attachment: %s", attachment)

				// TODO: Implement local file storage
				// attachmentPath := fmt.Sprintf("%s/attachment_%d.%s",
				// 	env.ServerPathImages, time.Now().Unix(), getFileExtension(attachment.MimeType))
				// err := os.WriteFile(attachmentPath, attachment.Data, 0644)
				// if err != nil {
				// 	log.Error("Failed to save attachment: %v", err)
				// 	continue
				// }

				// TODO: Store attachment path in database instead of binary data
				// tr.LinkedAttachments = append(tr.LinkedAttachments, attachmentPath)
			}
		}

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
	case "troublereport":
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

func processAttachments(c echo.Context) ([]string, error) {
	var attachments []string

	// Check if this is a multipart form
	if c.Request().MultipartForm == nil {
		err := c.Request().ParseMultipartForm(10 * 1024 * 1024) // 10MB max
		if err != nil {
			return nil, fmt.Errorf("parse multipart form: %w", err)
		}
	}

	// Get uploaded files
	files := c.FormValue("attachments")
	if files == "" {
		return attachments, nil
	}

	fileHeaders, err := c.FormFile("attachments")
	if err != nil && err != http.ErrMissingFile {
		return nil, fmt.Errorf("get file headers: %w", err)
	}

	if err == http.ErrMissingFile {
		return attachments, nil
	}

	log.Debug("processAttachments: attachments=%#v, files=%#v, fileHeaders=%#v", attachments, files, fileHeaders)

	// Process each file
	//for _, fileHeader := range fileHeaders {
	//	if fileHeader.Size == 0 {
	//		continue
	//	}

	//	// Validate file size (max 10MB)
	//	if fileHeader.Size > 10*1024*1024 {
	//		return nil, fmt.Errorf("file %s is too large (max 10MB)", fileHeader.Filename)
	//	}

	//	// Validate file type (images only)
	//	contentType := fileHeader.Header.Get("Content-Type")
	//	if !strings.HasPrefix(contentType, "image/") {
	//		return nil, fmt.Errorf("file %s is not an image", fileHeader.Filename)
	//	}

	//	// Process the file
	//	file, err := fileHeader.Open()
	//	if err != nil {
	//		return nil, fmt.Errorf("open file %s: %w", fileHeader.Filename, err)
	//	}
	//	defer file.Close()

	//	// Read file data
	//	data := make([]byte, fileHeader.Size)
	//	_, err = file.Read(data)
	//	if err != nil {
	//		return nil, fmt.Errorf("read file %s: %w", fileHeader.Filename, err)
	//	}

	//	// Create attachment
	//	attachment := &shared.Attachment{
	//		MimeType: contentType,
	//		Data:     data,
	//	}

	//	attachments = append(attachments, attachment)
	//}

	return attachments, nil
}

// getFileExtension returns a file extension based on MIME type
func getFileExtension(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "image/svg+xml":
		return "svg"
	default:
		return "bin"
	}
}
