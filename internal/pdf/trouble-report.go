package pdf

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/knackwurstking/pg-press/internal/shared"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

func GenerateTroubleReportPDF(tr *shared.TroubleReport) (*bytes.Buffer, error) {
	htmlContent, err := generateTroubleReportHTML(tr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HTML: %w", err)
	}

	pdfBuf, err := generatePDFFromHTML(htmlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return pdfBuf, nil
}

func generateTroubleReportHTML(tr *shared.TroubleReport) (template.HTML, error) {
	var contentHTML string

	if tr.UseMarkdown {
		markdown := goldmark.New(
			goldmark.WithExtensions(extension.GFM),
		)
		var buf bytes.Buffer
		err := markdown.Convert([]byte(tr.Content), &buf)
		if err != nil {
			return "", fmt.Errorf("failed to convert markdown: %w", err)
		}
		contentHTML = buf.String()
	} else {
		contentHTML = fmt.Sprintf("<pre>%s</pre>", escapeHTML(tr.Content))
	}

	imagesHTML, err := generateImagesHTML(tr.LinkedAttachments)
	if err != nil {
		return "", fmt.Errorf("failed to generate images HTML: %w", err)
	}

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            font-size: 12px;
            line-height: 1.6;
            color: #1a1a1a;
            padding: 20px;
            max-width: 800px;
            margin: 0 auto;
        }
        
        .header {
            margin-bottom: 20px;
        }
        
        .header h1 {
            font-size: 24px;
            color: #003366;
            margin-bottom: 5px;
        }
        
        .header .report-id {
            font-size: 12px;
            color: #808080;
        }
        
        .section {
            margin-bottom: 20px;
        }
        
        .section-title {
            font-size: 14px;
            font-weight: bold;
            background-color: #f0f8ff;
            padding: 8px 12px;
            border: 1px solid #ddd;
            margin-bottom: 10px;
        }
        
        .content {
            padding: 0 5px;
        }
        
        .content h1, .content h2, .content h3 {
            margin: 0.8em 0 0.4em 0;
            font-weight: bold;
            line-height: 1.3;
        }
        
        .content h1 { font-size: 1.4em; }
        .content h2 { font-size: 1.2em; }
        .content h3 { font-size: 1.1em; }
        
        .content p {
            margin: 0.5em 0 1em 0;
        }
        
        .content ul, .content ol {
            margin: 0.5em 0;
            padding-left: 1.5em;
            list-style: inherit;
        }
        
        .content ul { list-style-type: disc; }
        .content ol { list-style-type: decimal; }
        
        .content li {
            margin: 0.25em 0;
        }
        
        .content code {
            font-size: 0.9em;
            padding: 0.125em 0.25em;
            background-color: #f5f5f5;
            border-radius: 2px;
        }
        
        .content pre {
            margin: 1em 0;
            padding: 1em;
            background-color: #f5f5f5;
            border-radius: 4px;
            overflow-x: auto;
        }
        
        .content strong {
            font-weight: 600;
        }
        
        .content em {
            font-style: italic;
        }
        
        .content u {
            text-decoration: underline;
        }
        
        .content blockquote {
            margin: 1em 0;
            padding: 0.5em 1em;
            border-left: 3px solid #ccc;
            background-color: #f9f9f9;
        }
        
        .images-section {
            margin-top: 30px;
            page-break-before: always;
        }
        
        .images-section .section-title {
            background-color: #f0f8ff;
        }
        
        .images-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 15px;
            margin-top: 15px;
        }
        
        .image-container {
            break-inside: avoid;
        }
        
        .image-container img {
            max-width: 100%;
            height: auto;
            border: 1px solid #ddd;
            border-radius: 4px;
        }
        
        .image-container .image-caption {
            font-size: 10px;
            color: #666;
            margin-top: 5px;
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Fehlerbericht</h1>
        <div class="report-id">Report-ID: #{{ .ReportID }}</div>
    </div>
    
    <div class="section">
        <div class="section-title">TITEL</div>
        <div class="content">
            <p>{{ .Title }}</p>
        </div>
    </div>
    
    <div class="section">
        <div class="section-title">INHALT</div>
        <div class="content">
            {{ .ContentHTML }}
        </div>
    </div>
    
    {{ if .ImagesHTML }}
    <div class="section images-section">
        <div class="section-title">BILDER ({{ .ImageCount }})</div>
        <div class="images-grid">
            {{ .ImagesHTML }}
        </div>
    </div>
    {{ end }}
</body>
</html>`

	data := struct {
		ReportID    int
		Title       string
		ContentHTML template.HTML
		ImagesHTML  template.HTML
		ImageCount  int
	}{
		ReportID:    int(tr.ID),
		Title:       tr.Title,
		ContentHTML: template.HTML(contentHTML),
		ImagesHTML:  template.HTML(imagesHTML),
		ImageCount:  len(tr.LinkedAttachments),
	}

	tmpl, err := template.New("trouble-report").Parse(htmlTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return template.HTML(buf.String()), nil
}

func generateImagesHTML(attachments []string) (string, error) {
	if len(attachments) == 0 {
		return "", nil
	}

	var imageTags []string
	for _, a := range attachments {
		img := shared.NewImage(a, nil)
		if err := img.ReadFile(); err != nil {
			continue
		}

		base64Data := base64.StdEncoding.EncodeToString(img.Data)
		dataURI := fmt.Sprintf("data:%s;base64,%s", img.MimeType(), base64Data)

		imageTag := fmt.Sprintf(
			`<div class="image-container"><img loading="lazy" src="%s" alt="%s"><div class="image-caption">%s</div></div>`,
			dataURI,
			escapeHTML(a),
			escapeHTML(a),
		)
		imageTags = append(imageTags, imageTag)
	}

	return strings.Join(imageTags, "\n"), nil
}

func generatePDFFromHTML(htmlContent template.HTML) (*bytes.Buffer, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "trouble-report-pdf.html")

	err := os.WriteFile(tmpFile, []byte(htmlContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write temp HTML file: %w", err)
	}
	defer os.Remove(tmpFile)

	fileURL := "file://" + tmpFile

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("allow-file-access-from-files", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	var pdfBuf []byte

	err = chromedp.Run(taskCtx,
		chromedp.Navigate(fileURL),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(8.27).
				WithPaperHeight(11.69).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return bytes.NewBuffer(pdfBuf), nil
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, `'`, "&#39;")
	s = strings.ReplaceAll(s, `/`, "&#x2F;")
	return s
}
