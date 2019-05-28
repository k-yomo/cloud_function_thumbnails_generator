# Cloud Function Thumbnails Generator
Generate 100x100, 500x500, 1000x1000 thumbnails triggered by uploading image to the target bucket.

## Deploy
```
gcloud functions deploy GenerateThumbnails --runtime go111 --trigger-resource BUCKET_NAME --trigger-event google.storage.object.finalize
```

## Example
### Uploaded Image
![Uploaded Image](/example/cat.jpg?raw=true)

### Generated Thumbnails
- 1000x1000

![1000x1000](/example/1000x1000@cat.jpg?raw=true)

- 500x500

![500x500](/example/500x500@cat.jpg?raw=true)
- 100x100

![100x100](/example/100x100@cat.jpg?raw=true)
