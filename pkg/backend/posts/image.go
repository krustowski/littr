package posts

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/dsoprea/go-exif/v3"
	//"github.com/dsoprea/go-exif/v3/common"
	"golang.org/x/image/draw"
	_ "image/gif"
	_ "image/png"
)

func removeExif(imgBytes []byte, format string) (image.Image, []byte, error) {
	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, nil, err
	}

	// Create a buffer to hold the new image data (without EXIF metadata)
	var buf bytes.Buffer

	// Encode the image without EXIF metadata
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, img, nil)
	case "png":
		err = png.Encode(&buf, img)
	default:
		return nil, nil, err
	}

	if err != nil {
		return nil, nil, err
	}

	return img, buf.Bytes(), nil
}

// ResizeImage resizes an image to a target width and height while maintaining aspect ratio
/*func resizeImage(img image.Image, targetWidth, targetHeight int) image.Image {
	thumb := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), img, img.Bounds(), draw.Over, nil)

	return thumb
}*/

// ResizeImage resizes an image to a target width and height while maintaining aspect ratio
func resizeImage(img image.Image, targetWidth int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate aspect ratio and determine target height
	aspectRatio := float64(height) / float64(width)
	targetHeight := int(float64(targetWidth) * aspectRatio)

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

// FixOrientation checks the EXIF orientation tag and corrects the image's orientation if necessary
func fixOrientation(img image.Image, imgBytes []byte) (image.Image, error) {
	rawExif, err := exif.SearchAndExtractExif(imgBytes)
	if err != nil {
		if err == exif.ErrNoExif {
			return img, nil // If there's no EXIF data, return the original image
		}
		return nil, err
	}

	// Parse the EXIF data
	entries, _, err := exif.GetFlatExifData(rawExif, nil)
	if err != nil {
		return nil, err
	}

	// Find the Orientation tag
	for _, entry := range entries {
		if entry.TagName == "Orientation" {
			//fmt.Printf("orientation: entry.Value: %v\n", entry.Value)
			orientationRaw := entry.Value.([]uint16) // Orientation should be a uint16 value
			orientation := orientationRaw[0]

			//orientation := entry.Formatted
			//fmt.Println("Orientation tag found:", orientation)

			switch orientation {
			case 3: // 180 degrees
				img = rotate180(img)
			case 6: // 90 degrees clockwise
				img = rotate90(img)
			case 8: // 90 degrees counterclockwise
				img = rotate270(img)
			}
		}
	}

	return img, nil
}

// Rotate image 90 degrees clockwise
func rotate90(img image.Image) image.Image {
	bounds := img.Bounds()
	rotated := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))

	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			rotated.Set(bounds.Dy()-y-1, x, img.At(x, y))
		}
	}

	return rotated
}

// Rotate image 180 degrees
func rotate180(img image.Image) image.Image {
	bounds := img.Bounds()
	rotated := image.NewRGBA(bounds)

	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			rotated.Set(bounds.Dx()-x-1, bounds.Dy()-y-1, img.At(x, y))
		}
	}

	return rotated
}

// Rotate image 270 degrees (90 degrees counter-clockwise)
func rotate270(img image.Image) image.Image {
	bounds := img.Bounds()
	rotated := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))

	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			rotated.Set(y, bounds.Dx()-x-1, img.At(x, y))
		}
	}

	return rotated
}
