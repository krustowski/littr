package posts

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/dsoprea/go-exif/v3"
	//"github.com/dsoprea/go-exif/v3/common"
	"golang.org/x/image/draw"
	//"golang.org/x/image/webp" --- only implements a decoder, not an encoder (Sep 2024)
	"github.com/chai2010/webp"
)

//
//  []byte input handling
//

func convertGifToWebp(gifData []byte) ([]byte, error) {
	// Decode the GIF image from the byte stream
	gifImg, err := gif.Decode(bytes.NewReader(gifData))
	if err != nil {
		return nil, fmt.Errorf("error decoding GIF: %w", err)
	}

	// Create a buffer to store the WebP data
	var webpBuffer bytes.Buffer

	// Encode the image to WebP format and write to the buffer
	err = webp.Encode(&webpBuffer, gifImg, &webp.Options{Lossless: true})
	if err != nil {
		return nil, fmt.Errorf("error encoding to WebP: %w", err)
	}

	// Return the encoded WebP byte slice
	return webpBuffer.Bytes(), nil
}

// DecodeImage decodes a byte stream to an image
func decodeImage(imgData []byte, extInput string) (image.Image, string, error) {
	var img image.Image
	var format string
	var err error

	extension := strings.ToLower(extInput)

	switch extension {
	case "png", "jpg", "jpeg":
		img, format, err = image.Decode(bytes.NewReader(imgData))
		if err != nil {
			return nil, "", err
		}
		return img, format, nil

	case "gif":
		// this is to be used for the novel image's/thumbnail's extension, and for the encoder
		// we want to decrease the file's size (to convert it to WebP), and GIFs be thicc...
		format = "webp"
		img, err = gif.Decode(bytes.NewReader(imgData))
		if err != nil {
			return nil, "", err
		}

	case "webp":
		format = "webp"
		img, err = webp.Decode(bytes.NewReader(imgData))
		if err != nil {
			return nil, "", err
		}
	default:
		return nil, "", fmt.Errorf("unsupported format: %s", extension)
	}

	return img, format, nil
}

//
//  image.Image input handling
//

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
	case "gif", "webp":
		err := webp.Encode(&buf, img, &webp.Options{Lossless: true})
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return buf.Bytes(), nil
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
