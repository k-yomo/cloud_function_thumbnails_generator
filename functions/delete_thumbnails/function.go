package delete_thumbnails

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"regexp"
)

type GCSEvent struct {
	Bucket string `json:"bucket"`
	ObjectName   string `json:"name"`
	ContentType   string `json:"contentType"`
}

func DeleteThumbnails(ctx context.Context, e GCSEvent) error {
	storageClient, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("failed to init storage client: %v", err)
	}

	bucket := storageClient.Bucket(e.Bucket)
	if !shouldDeleteThumbnails(e.ObjectName, e.ContentType) {
		return nil
	}

	thumbnailSizes := []int{100, 500, 1000}
	for _, size := range thumbnailSizes {
		if err := deleteThumbnail(ctx, bucket, e.ObjectName, size, size); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to generate thumbnails for %q", e.ObjectName))
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
