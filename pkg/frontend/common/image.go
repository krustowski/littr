package common

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/dsoprea/go-exif/v3"
)

func ProcessImage(data *[]byte) (*[]byte, error) {
	// Inflate the image to memory.
	img, format, err := image.Decode(bytes.NewReader(*data))
	if err != nil {
		return nil, err
	}

	imgP := &img

	// Fix the image orientation for decoded image.
	imgP, err = FixOrientation(imgP, data)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	// Encode depending on the format.
	switch format {
	case "jpeg":
		err := jpeg.Encode(&buf, *imgP, nil)
		if err != nil {
			return nil, err
		}
	case "png":
		err := png.Encode(&buf, *imgP)
		if err != nil {
			return nil, err
		}
	}

	out := buf.Bytes()

	return &out, nil
}

// FixOrientation checks the EXIF orientation tag and corrects the image's orientation if necessary
func FixOrientation(img *image.Image, imgBytes *[]byte) (*image.Image, error) {
	rawExif, err := exif.SearchAndExtractExif(*imgBytes)
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
				*img = rotate180(img)
			case 6: // 90 degrees clockwise
				*img = rotate90(img)
			case 8: // 90 degrees counterclockwise
				*img = rotate270(img)
			}
		}
	}

	return img, nil
}

// Rotate image 90 degrees clockwise
func rotate90(img *image.Image) image.Image {
	bounds := (*img).Bounds()
	rotated := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))

	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			rotated.Set(bounds.Dy()-y-1, x, (*img).At(x, y))
		}
	}

	return rotated
}

// Rotate image 180 degrees
func rotate180(img *image.Image) image.Image {
	bounds := (*img).Bounds()
	rotated := image.NewRGBA(bounds)

	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			rotated.Set(bounds.Dx()-x-1, bounds.Dy()-y-1, (*img).At(x, y))
		}
	}

	return rotated
}

// Rotate image 270 degrees (90 degrees counter-clockwise)
func rotate270(img *image.Image) image.Image {
	bounds := (*img).Bounds()
	rotated := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))

	for x := 0; x < bounds.Dx(); x++ {
		for y := 0; y < bounds.Dy(); y++ {
			rotated.Set(y, bounds.Dx()-x-1, (*img).At(x, y))
		}
	}

	return rotated
}
