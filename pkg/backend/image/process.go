// Image (de)coding and processing magic package.
package image

import (
	"fmt"
	"image"
	"net/http"
	"os"
	"strings"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"
)

type ImageProcessPayload struct {
	ImageByteData *[]byte
	ImageFileName string
	ImageBaseName string
}

func ProcessPost(post *models.Post, postContent *string) (error, int) {
	data := &ImageProcessPayload{
		ImageByteData: &post.Data,
		ImageFileName: post.Figure,
		ImageBaseName: post.ID,
	}

	baseImageURL, err := ProcessImageBytes(data)
	if err != nil {
		return err, common.DecideStatusFromError(err)
	}

	postContent = baseImageURL

	return nil, http.StatusOK
}

func ProcessImageBytes(data *ImageProcessPayload) (*string, error) {
	var (
		newBytes *[]byte
		err      error
		img      *image.Image
		format   string
	)

	// Ensure the data presence.
	if data.ImageByteData == nil || data.ImageFileName == "" {
		return nil, fmt.Errorf(common.ERR_INPUT_DATA_FAIL)
	}

	// Parse the filename.
	fileExplode := strings.Split(data.ImageFileName, ".")
	extension := strings.ToLower(fileExplode[len(fileExplode)-1])

	// decode image from []byte stream
	img, format, err = DecodeImage(data.ImageByteData, extension)
	if err != nil {
		//l.Msg(common.ERR_IMG_DECODE_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return nil, fmt.Errorf(common.ERR_IMG_DECODE_FAIL + err.Error())
	}

	// Decide the action according to the extension.
	switch extension {
	case "png", "jpg", "jpeg":
		// fix the image orientation for decoded image
		img, err = FixOrientation(img, data.ImageByteData)
		if err != nil {
			//l.Msg(common.ERR_IMG_ORIENTATION_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return nil, fmt.Errorf(common.ERR_IMG_ORIENTATION_FAIL + err.Error())
		}

		// re-encode the image to flush EXIF metadata header
		newBytes, err = EncodeImage(img, format)
		if err != nil {
			//l.Msg(common.ERR_IMG_ENCODE_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return nil, fmt.Errorf(common.ERR_IMG_ENCODE_FAIL + err.Error())
		}

	case "gif":
		// to be converted to WebP
		format = "webp"

		newBytes, err = ConvertGifToWebp(data.ImageByteData)
		if err != nil {
			//l.Msg(common.ERR_IMG_GIF_TO_WEBP_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return nil, fmt.Errorf(common.ERR_IMG_GIF_TO_WEBP_FAIL + err.Error())
		}

	default:
		//l.Msg(common.ERR_IMG_UNKNOWN_TYPE).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return nil, fmt.Errorf(common.ERR_IMG_UNKNOWN_TYPE)
	}

	// prepare the novel image's filename
	imageBaseName := data.ImageBaseName + "." + format

	// upload the novel image to local storage
	if err := os.WriteFile("/opt/pix/"+imageBaseName, *newBytes, 0600); err != nil {
		//l.Msg(common.ERR_IMG_SAVE_FILE_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
		return nil, fmt.Errorf(common.ERR_IMG_SAVE_FILE_FAIL + err.Error())
	}

	// generate thumbnails --- keep aspect ratio in px
	//thumbImg := ResizeImage(img, 450)
	thumbImg := CropToSquare(img)

	// encode the thumbnail back to []byte
	thumbImgData, err := EncodeImage(thumbImg, format)
	if err != nil {
		//l.Msg(common.ERR_IMG_THUMBNAIL_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
		return nil, fmt.Errorf(common.ERR_IMG_THUMBNAIL_FAIL + err.Error())
	}

	// write the thumbnail byte stream to a file
	if err := os.WriteFile("/opt/pix/thumb_"+imageBaseName, *thumbImgData, 0600); err != nil {
		//l.Msg(common.ERR_IMG_SAVE_FILE_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
		return nil, fmt.Errorf(common.ERR_IMG_SAVE_FILE_FAIL + err.Error())
	}

	return &imageBaseName, nil
}
