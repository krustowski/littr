package users

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"golang.org/x/image/draw"
)

// ResizeImage resizes an image to a target width and height while maintaining aspect ratio
func resizeImage(img image.Image, targetWidth, targetHeight int) image.Image {
	// Create a new image with the target size
	thumb := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	// Perform high-quality image resizing
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), img, img.Bounds(), draw.Over, nil)

	return thumb
}

// CropToSquare crops an image to a 1:1 aspect ratio (square)
func cropToSquare(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Determine the size of the square
	var cropSize int
	if width < height {
		cropSize = width
	} else {
		cropSize = height
	}

	// Calculate cropping rectangle, centering the crop
	x0 := (width - cropSize) / 2
	y0 := (height - cropSize) / 2
	x1 := x0 + cropSize
	y1 := y0 + cropSize

	// Crop the image to the calculated square
	cropRect := image.Rect(x0, y0, x1, y1)
	croppedImg := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(cropRect)

	return croppedImg
}

// DecodeImage decodes a byte stream to an image
func decodeImage(imgData []byte) (image.Image, string, error) {
	img, format, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, "", err
	}
	return img, format, nil
}

// EncodeImage encodes an image back to byte stream (JPEG or PNG)
func encodeImage(img image.Image, format string) ([]byte, error) {
	var buf bytes.Buffer

	// Encode depending on the format
	switch format {
	case "jpeg":
		err := jpeg.Encode(&buf, img, nil)
		if err != nil {
			return nil, err
		}
	case "png":
		err := png.Encode(&buf, img)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown format")
	}

	return buf.Bytes(), nil
}
