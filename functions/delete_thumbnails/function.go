package delete_thumbnails

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
)

type GCSEvent struct {
	Bucket      string `json:"bucket"`
	ObjectName  string `json:"name"`
	ContentType string `json:"contentType"`
}

func DeleteThumbnails(ctx context.Context, e GCSEvent) error {
	gcsClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to init storage client: %v", err)
	}

	bucket := gcsClient.Bucket(e.Bucket)
	if !shouldDeleteThumbnails(e.ObjectName, e.ContentType) {
		return nil
	}

	thumbnailSizes := []int{100, 500, 1000}
	for _, size := range thumbnailSizes {
		if err := deleteThumbnail(ctx, bucket, e.ObjectName, size, size); err != nil {
			return errors.Wrapf(err, "failed to generate thumbnails for %q", e.ObjectName)
		}
	}

	log.Printf("thumbnails for %q have been generated", e.ObjectName)

	return nil
}

func deleteThumbnail(ctx context.Context, bucket *storage.BucketHandle, originalImageName string, width, height int) error {
	return bucket.Object(fmt.Sprintf("%dx%d@", width, height) + originalImageName).Delete(ctx)
}

func shouldDeleteThumbnails(fileName string, contentType string) bool {
	if !(contentType == "image/jpeg" || contentType == "image/png") {
		log.Printf("%q is not image", fileName)
		return false
	}

	resizedImage, _ := regexp.MatchString("^[0-9]+x[0-9]+@", fileName)
	if resizedImage {
		log.Printf("%q is being deleted", fileName)
		return false
	}
	return true
}
