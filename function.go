package function

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/jpeg"
	"log"
	"regexp"
)

var (
	storageClient *storage.Client
)

type GCSEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

func init() {
	var err error
	storageClient, err = storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("failed to init storage client: %v", err)
	}
}

func GenerateThumbnails(ctx context.Context, e GCSEvent) error {
	if !shouldGenerateThumbnails(e.Name) {
		return nil
	}

	bucket := storageClient.Bucket(e.Bucket)

	thumbnailSizes := []int{100, 500, 1000}
	for _, size := range thumbnailSizes {
		if err := generateResizedImage(ctx, bucket, e.Name, size, size); err != nil {
			return err
		}
	}

	log.Printf("thumbnails have been generated")

	return nil
}

func generateResizedImage(ctx context.Context, bucket *storage.BucketHandle, originalImageName string, width, height int) error {
	reader, err := bucket.Object(originalImageName).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("failed to read object: %v", err)
	}

	writer := bucket.Object(fmt.Sprintf("%dx%d@", width, height) + originalImageName).NewWriter(ctx)
	defer writer.Close()

	src, _, err := image.Decode(reader)
	if err != nil {
		return fmt.Errorf("failed to decode image: %v", err)
	}

	img := imaging.Thumbnail(src, width, height, imaging.Lanczos)
	buff, err := encodeToJpeg(img)
	if err != nil {
		return fmt.Errorf("failed to encode image: %v", err)
	}

	_, err = writer.Write(buff.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write resized image %v", err)
	}

	return nil
}

func shouldGenerateThumbnails(fileName string) bool {
	resizable, _ := regexp.MatchString("(jpg|jpeg|png)$", fileName)
	if !resizable {
		log.Printf("%q is not image", fileName)
		return false
	}

	resized, _ := regexp.MatchString("^[0-9]+x[0-9]+@", fileName)
	if resized {
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
