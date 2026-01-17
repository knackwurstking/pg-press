package pdf

import (
	"fmt"
	"os"

	"github.com/jung-kurt/gofpdf/v2"
)

type imageOptions struct {
	PDF        *gofpdf.Fpdf
	Translator func(string) string
}

type imageLayoutProps struct {
	PageWidth   float64
	LeftMargin  float64
	RightMargin float64
	UsableWidth float64
	ImageWidth  float64
}

type imagePositionOptions struct {
	StartIndex  int
	TotalImages int
	ImageWidth  float64
	LeftMargin  float64
	RightX      float64
	CurrentY    *float64
}

func processImageRow(
	o *imageOptions,
	layout *imageLayoutProps,
	position *imagePositionOptions,
	images []*models.Attachment,
) {
	leftHeight, rightHeight := calculateImageHeights(
		o, images, position.StartIndex, layout.ImageWidth)

	actualRowHeight := max(leftHeight, rightHeight)
	if actualRowHeight == 0 {
		actualRowHeight = 60.0
	}

	captionY := *position.CurrentY
	imageY := captionY + 6

	// Check if we need a new page
	if imageY+actualRowHeight+25 > 270 {
		o.PDF.AddPage()
		_, *position.CurrentY = o.PDF.GetXY()
		captionY = *position.CurrentY
		imageY = captionY + 6
	}

	addImageCaptions(o, position, captionY)
	addImages(o, images, position, imageY)

	// Update currentY to the bottom of the images
	*position.CurrentY = imageY + actualRowHeight
}

func calculateImageHeights(
	o *imageOptions,
	images []*models.Attachment,
	startIndex int,
	imageWidth float64,
) (leftHeight, rightHeight float64) {
	// Calculate left image height
	if startIndex < len(images) {
		leftHeight = calculateSingleImageHeight(o, images[startIndex], imageWidth)
	}

	// Calculate right image height if it exists
	if startIndex+1 < len(images) {
		rightHeight = calculateSingleImageHeight(o, images[startIndex+1], imageWidth)
	}

	return leftHeight, rightHeight
}

func calculateSingleImageHeight(
	o *imageOptions,
	image *models.Attachment,
	imageWidth float64,
) (height float64) {
	tmpFile, err := createTempImageFile(image)
	if err != nil {
		return 60.0
	}
	defer os.Remove(tmpFile)

	imageType := getImageType(image.MimeType)
	info := o.PDF.RegisterImage(tmpFile, imageType)
	if info != nil {
		return (imageWidth * info.Height()) / info.Width()
	}

	return 60.0
}

func addImageCaptions(
	o *imageOptions,
	position *imagePositionOptions,
	captionY float64,
) {
	o.PDF.SetFont("Arial", "", 9)

	// Left image caption
	o.PDF.SetXY(position.LeftMargin, captionY)
	o.PDF.CellFormat(position.ImageWidth, 4,
		o.Translator(fmt.Sprintf("Anhang %d", position.StartIndex+1)),
		"0", 0, "C", false, 0, "")

	// Right image caption (if exists)
	if position.StartIndex+1 < position.TotalImages {
		o.PDF.SetXY(position.RightX, captionY)
		o.PDF.CellFormat(
			position.ImageWidth, 4,
			o.Translator(fmt.Sprintf("Anhang %d", position.StartIndex+2)),
			"0", 0, "C", false, 0, "",
		)
	}
}

func addImages(
	o *imageOptions,
	images []*models.Attachment,
	position *imagePositionOptions,
	imageY float64,
) {
	// Add left image
	if position.StartIndex < len(images) {
		addSingleImage(
			o,
			images[position.StartIndex],
			position.LeftMargin, imageY, position.ImageWidth,
		)
	}

	// Add right image (if it exists)
	if position.StartIndex+1 < len(images) {
		addSingleImage(
			o,
			images[position.StartIndex+1],
			position.RightX, imageY, position.ImageWidth,
		)
	}
}

func addSingleImage(
	o *imageOptions,
	image *models.Attachment,
	x, y, width float64,
) {
	tmpFile, err := createTempImageFile(image)
	if err != nil {
		return
	}
	defer os.Remove(tmpFile)

	imageType := getImageType(image.MimeType)
	o.PDF.Image(tmpFile, x, y, width, 0, false, imageType, 0, "")
}

func createTempImageFile(image *models.Attachment) (string, error) {
	tmpFile, err := os.CreateTemp("",
		fmt.Sprintf("attachment_%s_*.jpg", image.ID))
	if err != nil {
		return "", err
	}

	_, err = tmpFile.Write(image.Data)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

func getImageType(mimeType string) string {
	switch mimeType {
	case "image/jpeg", "image/jpg":
		return "JPG"
	case "image/png":
		return "PNG"
	case "image/gif":
		return "GIF"
	default:
		return "JPG"
	}
}
