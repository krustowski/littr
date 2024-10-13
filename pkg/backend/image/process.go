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

// ProcessBytes takes pointer to a post with image data as []byte stream,
// and pointer to content string to return as a reference to the novel image
func ProcessBytes(post *models.Post, postContent *string) (error, int) {
	var (
		newBytes *[]byte
		err      error
		img      *image.Image
		format   string
	)

	// recheck the data presence
	if post.Data == nil || post.Figure == "" {
		return fmt.Errorf(common.ERR_INPUT_DATA_FAIL), http.StatusBadRequest
	}

	// parse the filename
	fileExplode := strings.Split(post.Figure, ".")
	extension := fileExplode[len(fileExplode)-1]

	// decode image from []byte stream
	img, format, err = DecodeImage(&post.Data, extension)
	if err != nil {
		//l.Msg(common.ERR_IMG_DECODE_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return fmt.Errorf(common.ERR_IMG_DECODE_FAIL + err.Error()), http.StatusBadRequest
	}

	switch extension {
	case "png", "jpg", "jpeg":
		// fix the image orientation for decoded image
		img, err = FixOrientation(img, &post.Data)
		if err != nil {
			//l.Msg(common.ERR_IMG_ORIENTATION_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return fmt.Errorf(common.ERR_IMG_ORIENTATION_FAIL + err.Error()), http.StatusInternalServerError
		}

		// re-encode the image to flush EXIF metadata header
		newBytes, err = EncodeImage(img, format)
		if err != nil {
			//l.Msg(common.ERR_IMG_ENCODE_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return fmt.Errorf(common.ERR_IMG_ENCODE_FAIL + err.Error()), http.StatusInternalServerError
		}

	case "gif":
		// to be converted to WebP
		format = "webp"

		newBytes, err = ConvertGifToWebp(&post.Data)
		if err != nil {
			//l.Msg(common.ERR_IMG_GIF_TO_WEBP_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return fmt.Errorf(common.ERR_IMG_GIF_TO_WEBP_FAIL + err.Error()), http.StatusInternalServerError
		}

	default:
		//l.Msg(common.ERR_IMG_UNKNOWN_TYPE).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return fmt.Errorf(common.ERR_IMG_UNKNOWN_TYPE), http.StatusBadRequest
	}

	// prepare the novel image's filename
	content := post.ID + "." + format

	// upload the novel image to local storage
	//if err := os.WriteFile("/opt/pix/"+content, post.Data, 0600); err != nil {
	if err := os.WriteFile("/opt/pix/"+content, *newBytes, 0600); err != nil {
		//l.Msg(common.ERR_IMG_SAVE_FILE_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
		return fmt.Errorf(common.ERR_IMG_SAVE_FILE_FAIL + err.Error()), http.StatusInternalServerError
	}

	// generate thumbnails --- keep aspect ratio in px
	thumbImg := ResizeImage(img, 450)

	// encode the thumbnail back to []byte
	thumbImgData, err := EncodeImage(&thumbImg, format)
	if err != nil {
		//l.Msg(common.ERR_IMG_THUMBNAIL_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
		return fmt.Errorf(common.ERR_IMG_THUMBNAIL_FAIL + err.Error()), http.StatusInternalServerError
	}

	// write the thumbnail byte stream to a file
	if err := os.WriteFile("/opt/pix/thumb_"+content, *thumbImgData, 0600); err != nil {
		//l.Msg(common.ERR_IMG_SAVE_FILE_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
		return fmt.Errorf(common.ERR_IMG_SAVE_FILE_FAIL + err.Error()), http.StatusInternalServerError
	}

	// flush post's image-related fields
	*postContent = content
	post.Figure = content
	post.Data = []byte{}

	return nil, 0
}
