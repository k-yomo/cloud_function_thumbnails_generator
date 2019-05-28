# Cloud Function Thumbnails Generator
Generate 100x100, 500x500, 1000x1000 thumbnails triggered by uploading image to the target bucket.

## Deploy
```
gcloud functions deploy GenerateThumbnails --runtime go111 --trigger-resource BUCKET_NAME --trigger-event google.storage.object.finalize
```
