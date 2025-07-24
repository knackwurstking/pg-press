package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/knackwurstking/pg-vis/pgvis"
	"github.com/labstack/echo/v4"
)

// Test data for various file types
var testFiles = map[string]struct {
	filename string
	content  []byte
	mimeType string
}{
	"image": {
		filename: "test_image.png",
		content:  []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG header
		mimeType: "image/png",
	},
	"text": {
		filename: "test_log.txt",
		content:  []byte("This is a test log file\nLine 2\nLine 3"),
		mimeType: "text/plain",
	},
	"pdf": {
		filename: "document.pdf",
		content:  []byte("%PDF-1.4\n1 0 obj\n<<\n/Type /Catalog"),
		mimeType: "application/pdf",
	},
}

// TestAttachmentValidation tests the pgvis.Attachment validation
func TestAttachmentValidation(t *testing.T) {
	tests := []struct {
		name        string
		attachment  *pgvis.Attachment
		expectError bool
	}{
		{
			name: "valid attachment",
			attachment: &pgvis.Attachment{
				ID:       "test_file_123",
				MimeType: "text/plain",
				Data:     []byte("test content"),
			},
			expectError: false,
		},
		{
			name: "empty ID",
			attachment: &pgvis.Attachment{
				ID:       "",
				MimeType: "text/plain",
				Data:     []byte("test content"),
			},
			expectError: true,
		},
		{
			name: "empty mime type",
			attachment: &pgvis.Attachment{
				ID:       "test_file_123",
				MimeType: "",
				Data:     []byte("test content"),
			},
			expectError: true,
		},
		{
			name: "nil data",
			attachment: &pgvis.Attachment{
				ID:       "test_file_123",
				MimeType: "text/plain",
				Data:     nil,
			},
			expectError: true,
		},
		{
			name: "oversized attachment",
			attachment: &pgvis.Attachment{
				ID:       "large_file_123",
				MimeType: "application/octet-stream",
				Data:     make([]byte, pgvis.MaxAttachmentDataSize+1),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.attachment.Validate()
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestMimeTypeDetection tests MIME type detection from file content and extensions
func TestMimeTypeDetection(t *testing.T) {
	tests := []struct {
		filename     string
		content      []byte
		expectedType string
	}{
		{
			filename:     "image.png",
			content:      []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			expectedType: "image/png",
		},
		{
			filename:     "document.pdf",
			content:      []byte("%PDF-1.4"),
			expectedType: "application/pdf",
		},
		{
			filename:     "text.txt",
			content:      []byte("Hello world"),
			expectedType: "text/plain",
		},
		{
			filename:     "archive.zip",
			content:      []byte("PK\x03\x04"),
			expectedType: "application/zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			// Test Go's built-in detection
			detectedType := http.DetectContentType(tt.content)
			if strings.HasPrefix(tt.expectedType, "image/") && !strings.HasPrefix(detectedType, "image/") {
				t.Errorf("Expected image type for %s, got %s", tt.filename, detectedType)
			}
		})
	}
}

// TestFileUploadProcessing tests the file upload processing logic
func TestFileUploadProcessing(t *testing.T) {
	// Create a multipart form with test files
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add a test file
	testFile := testFiles["text"]
	part, err := writer.CreateFormFile("attachments", testFile.filename)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = part.Write(testFile.content)
	if err != nil {
		t.Fatalf("Failed to write file content: %v", err)
	}

	// Add form fields
	writer.WriteField("title", "Test Report")
	writer.WriteField("content", "Test content with attachment")

	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	// Create echo context
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/trouble-reports/dialog-edit", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test parsing multipart form
	form, err := c.MultipartForm()
	if err != nil {
		t.Fatalf("Failed to parse multipart form: %v", err)
	}

	files := form.File["attachments"]
	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}

	fileHeader := files[0]
	if fileHeader.Filename != testFile.filename {
		t.Errorf("Expected filename %s, got %s", testFile.filename, fileHeader.Filename)
	}

	// Test file reading
	file, err := fileHeader.Open()
	if err != nil {
		t.Fatalf("Failed to open uploaded file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read file content: %v", err)
	}

	if !bytes.Equal(content, testFile.content) {
		t.Errorf("File content mismatch")
	}
}

// TestAttachmentSizeValidation tests file size validation
func TestAttachmentSizeValidation(t *testing.T) {
	tests := []struct {
		name        string
		size        int64
		expectError bool
	}{
		{
			name:        "valid small file",
			size:        1024, // 1KB
			expectError: false,
		},
		{
			name:        "valid large file",
			size:        5 * 1024 * 1024, // 5MB
			expectError: false,
		},
		{
			name:        "oversized file",
			size:        11 * 1024 * 1024, // 11MB
			expectError: true,
		},
		{
			name:        "maximum allowed size",
			size:        pgvis.MaxAttachmentDataSize,
			expectError: false,
		},
		{
			name:        "just over limit",
			size:        pgvis.MaxAttachmentDataSize + 1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.size)
			attachment := &pgvis.Attachment{
				ID:       fmt.Sprintf("test_file_%d", tt.size),
				MimeType: "application/octet-stream",
				Data:     data,
			}

			err := attachment.Validate()
			if tt.expectError && err == nil {
				t.Errorf("Expected error for size %d but got none", tt.size)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for size %d but got: %v", tt.size, err)
			}
		})
	}
}

// TestTroubleReportWithAttachments tests trouble report creation with attachments
func TestTroubleReportWithAttachments(t *testing.T) {
	// Create test attachments
	attachments := []*pgvis.Attachment{
		{
			ID:       "screenshot_123456789_0",
			MimeType: "image/png",
			Data:     testFiles["image"].content,
		},
		{
			ID:       "log_file_123456789_1",
			MimeType: "text/plain",
			Data:     testFiles["text"].content,
		},
	}

	// Validate all attachments
	for _, att := range attachments {
		if err := att.Validate(); err != nil {
			t.Fatalf("Attachment validation failed: %v", err)
		}
	}

	// Create a trouble report with attachments
	user := &pgvis.User{
		ID:         1,
		UserName:   "testuser",
		TelegramID: 123456789,
	}

	mod := pgvis.NewModified(user, pgvis.TroubleReportMod{
		Title:             "Test Report with Attachments",
		Content:           "This is a test report with multiple attachments",
		LinkedAttachments: attachments,
	})

	tr := pgvis.NewTroubleReport("Test Report with Attachments", "This is a test report with multiple attachments", mod)
	tr.LinkedAttachments = attachments

	// Validate the trouble report
	if err := tr.Validate(); err != nil {
		t.Fatalf("Trouble report validation failed: %v", err)
	}

	// Test attachment operations
	if !tr.HasAttachments() {
		t.Error("Expected trouble report to have attachments")
	}

	if tr.AttachmentCount() != 2 {
		t.Errorf("Expected 2 attachments, got %d", tr.AttachmentCount())
	}

	// Test adding another attachment
	newAttachment := &pgvis.Attachment{
		ID:       "additional_file_123456789_2",
		MimeType: "application/pdf",
		Data:     testFiles["pdf"].content,
	}

	err := tr.AddAttachment(newAttachment)
	if err != nil {
		t.Fatalf("Failed to add attachment: %v", err)
	}

	if tr.AttachmentCount() != 3 {
		t.Errorf("Expected 3 attachments after adding one, got %d", tr.AttachmentCount())
	}

	// Test removing an attachment
	err = tr.RemoveAttachment(1) // Remove the second attachment
	if err != nil {
		t.Fatalf("Failed to remove attachment: %v", err)
	}

	if tr.AttachmentCount() != 2 {
		t.Errorf("Expected 2 attachments after removing one, got %d", tr.AttachmentCount())
	}
}

// TestAttachmentUtilityMethods tests utility methods on attachments
func TestAttachmentUtilityMethods(t *testing.T) {
	tests := []struct {
		name       string
		attachment *pgvis.Attachment
		isImage    bool
		isDocument bool
		isArchive  bool
	}{
		{
			name: "PNG image",
			attachment: &pgvis.Attachment{
				ID:       "image_test",
				MimeType: "image/png",
				Data:     []byte("fake png data"),
			},
			isImage:    true,
			isDocument: false,
			isArchive:  false,
		},
		{
			name: "PDF document",
			attachment: &pgvis.Attachment{
				ID:       "pdf_test",
				MimeType: "application/pdf",
				Data:     []byte("fake pdf data"),
			},
			isImage:    false,
			isDocument: true,
			isArchive:  false,
		},
		{
			name: "ZIP archive",
			attachment: &pgvis.Attachment{
				ID:       "zip_test",
				MimeType: "application/zip",
				Data:     []byte("fake zip data"),
			},
			isImage:    false,
			isDocument: false,
			isArchive:  true,
		},
		{
			name: "Text file",
			attachment: &pgvis.Attachment{
				ID:       "txt_test",
				MimeType: "text/plain",
				Data:     []byte("fake text data"),
			},
			isImage:    false,
			isDocument: true,
			isArchive:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.attachment.IsImage() != tt.isImage {
				t.Errorf("IsImage() = %v, expected %v", tt.attachment.IsImage(), tt.isImage)
			}
			if tt.attachment.IsDocument() != tt.isDocument {
				t.Errorf("IsDocument() = %v, expected %v", tt.attachment.IsDocument(), tt.isDocument)
			}
			if tt.attachment.IsArchive() != tt.isArchive {
				t.Errorf("IsArchive() = %v, expected %v", tt.attachment.IsArchive(), tt.isArchive)
			}
		})
	}
}

// TestAttachmentClone tests the Clone method
func TestAttachmentClone(t *testing.T) {
	original := &pgvis.Attachment{
		ID:       "original_file",
		MimeType: "text/plain",
		Data:     []byte("original data"),
	}

	cloned := original.Clone()

	// Test that clone is equal but separate
	if !original.Equals(cloned) {
		t.Error("Cloned attachment should be equal to original")
	}

	// Test that modifying clone doesn't affect original
	cloned.Data[0] = 'X'
	if original.Equals(cloned) {
		t.Error("Original and clone should be different after modification")
	}
}

// Benchmark tests for performance validation
func BenchmarkAttachmentValidation(b *testing.B) {
	attachment := &pgvis.Attachment{
		ID:       "benchmark_file_123456789",
		MimeType: "application/octet-stream",
		Data:     make([]byte, 1024*1024), // 1MB
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = attachment.Validate()
	}
}

func BenchmarkMimeTypeDetection(b *testing.B) {
	data := testFiles["image"].content

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = http.DetectContentType(data)
	}
}

// Example usage demonstration
func ExampleAttachmentUsage() {
	// Create a new attachment
	attachment := &pgvis.Attachment{
		ID:       "example_screenshot_123456789_0",
		MimeType: "image/png",
		Data:     []byte("fake image data"),
	}

	// Validate the attachment
	if err := attachment.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	// Check file type
	if attachment.IsImage() {
		fmt.Println("This is an image file")
	}

	// Create trouble report with attachment
	user := &pgvis.User{ID: 1, UserName: "example_user"}
	mod := pgvis.NewModified(user, pgvis.TroubleReportMod{
		Title:             "Example Report",
		Content:           "Example content",
		LinkedAttachments: []*pgvis.Attachment{attachment},
	})

	tr := pgvis.NewTroubleReport("Example Report", "Example content", mod)
	tr.LinkedAttachments = []*pgvis.Attachment{attachment}

	fmt.Printf("Created trouble report with %d attachments\n", tr.AttachmentCount())
	// Output: Created trouble report with 1 attachments
}
