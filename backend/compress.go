package backend

import (
	"os"

	// impoerted for init purposes
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

	//buff := bytes.NewReader(buffer)

	//imagePath := "/image.jpg"
	//dest := "path/to/thumb_image.jpg"
	gen := thumb.NewGenerator(config)

	i, err := gen.NewImageFromFile(src)
	if err != nil {
		return err
	}

	thumbBytes, err := gen.CreateThumbnail(i)
	if err != nil {
		return err
	}

	err = os.WriteFile(dest, thumbBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
