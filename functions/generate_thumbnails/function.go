package generate_thumbnails

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"regexp"

	"cloud.google.com/go/storage"
	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
)

type GCSEvent struct {
	Bucket      string `json:"bucket"`
	ObjectName  string `json:"name"`
	ContentType string `json:"contentType"`
}

func GenerateThumbnails(ctx context.Context, e GCSEvent) error {
	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to init storage client: %v", err)
	}

	if !shouldGenerateThumbnails(e.ObjectName, e.ContentType) {
		return nil
	}

	bucket := gcsClient.Bucket(e.Bucket)
	obj := bucket.Object(e.ObjectName)
	thumbnailSizes := []int{100, 500, 1000}
	for _, size := range thumbnailSizes {
		if err := generateResizedImage(ctx, bucket, obj, size, size); err != nil {
			return errors.Wrapf(err, "failed to generate thumbnails for %q", e.ObjectName)
		}
	}

	log.Printf("thumbnails for %q have been generated", e.ObjectName)

	return nil
}

func generateResizedImage(ctx context.Context, bucket *storage.BucketHandle, originalImage *storage.ObjectHandle, width, height int) error {
	reader, err := originalImage.NewReader(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to read image")
	}

	objName := fmt.Sprintf("%dx%d@%s", width, height, originalImage.ObjectName())
	writer := bucket.Object(objName).NewWriter(ctx)
	defer writer.Close()

	src, err := imaging.Decode(reader, imaging.AutoOrientation(true))
	if err != nil {
		return errors.Wrap(err, "failed to decode original image")
	}

	img := imaging.Thumbnail(src, width, height, imaging.Lanczos)
	buff, err := encodeToJpeg(img)
	if err != nil {
		return errors.Wrapf(err, "failed to encode %dx%d thumbnail image", width, height)
	}

	if _, err = writer.Write(buff.Bytes()); err != nil {
		return errors.Wrap(err, "failed to write resized image")
	}

	return nil
}

func shouldGenerateThumbnails(fileName string, contentType string) bool {
	if !(contentType == "image/jpeg" || contentType == "image/png") {
		log.Printf("%q is not image", fileName)
		return false
	}

	resizedImage, _ := regexp.MatchString("^[0-9]+x[0-9]+@", fileName)
	if resizedImage {
		log.Printf("%q is already resized", fileName)
		return false
	}
	return true
}

func encodeToJpeg(img *image.NRGBA) (*bytes.Buffer, error) {
	buff := &bytes.Buffer{}
	err := jpeg.Encode(buff, img, nil)
	return buff, err
}
