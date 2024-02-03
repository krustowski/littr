package backend

import (
	"os"

	// imported for init purposes
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	thumb "github.com/prplecake/go-thumbnail"
)

func genThumbnails(src, dest string) error {
	var config = thumb.Generator{
		DestinationPath:   "",
		DestinationPrefix: "thumb_",
		Scaler:            "CatmullRom",
	}

	gen := thumb.NewGenerator(config)

	// load the image from the source stream
	img, err := gen.NewImageFromFile(src)
	if err != nil {
		return err
	}

	// generate a thumbnail
	thumbBytes, err := gen.CreateThumbnail(img)
	if err != nil {
		return err
	}

	// write the thumbnail to a file
	err = os.WriteFile(dest, thumbBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
