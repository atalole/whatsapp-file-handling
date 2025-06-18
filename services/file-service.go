package FileServices

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
)

var (
	bucketName = "quick-hub"
	region     = "us-east-1" // e.g., "ap-south-1"
)

func FileUpload(file multipart.File, fileHeader *multipart.FileHeader, mimeType string, encoding string) (string, error) {

	ctx := context.TODO()
	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv("AWS_KEY"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "")))

	if err != nil {
		log.Printf("unable to load AWS SDK config: %v", err)
		return "", err
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)
	var keys = SpittedName(fileHeader.Filename)
	fmt.Println("keys", keys)
	// Generate unique file name
	key := fmt.Sprintf("whatsapp-media-library/%d_%s", time.Now().Unix(), keys)

	// Upload input
	uploadInput := &s3.PutObjectInput{
		Bucket:          aws.String(bucketName),
		Key:             aws.String(key),
		Body:            file,
		ACL:             "public-read", // Optional: makes the file publicly accessible
		ContentType:     aws.String(mimeType),
		ContentEncoding: aws.String(encoding),
	}

	// Upload
	_, err = uploader.Upload(ctx, uploadInput)
	if err != nil {
		log.Printf("failed to upload file: %v", err)
		return "", err
	}

	// Return the file URL

	fileURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, region, key)
	return fileURL, nil

}

func DetectMimeAndEncoding(buffer []byte, filename string) (string, string) {
	// 1. First try mimetype for better detection
	mimeDetector := mimetype.Detect(buffer)

	// 2. If mimetype gives us encoding info, use that
	if mimeDetector != nil {
		mimeParts := strings.SplitN(mimeDetector.String(), ";", 2)
		if len(mimeParts) > 1 {
			// Example: "text/plain; charset=utf-8"
			return mimeParts[0], strings.TrimSpace(mimeParts[1])
		}
		return mimeDetector.String(), "binary" // Default for non-text
	}

	// 3. Fallback to http.DetectContentType
	mimeType := http.DetectContentType(buffer)
	mimeParts := strings.SplitN(mimeType, ";", 2)

	if len(mimeParts) > 1 {
		return mimeParts[0], strings.TrimSpace(mimeParts[1])
	}

	// 4. Guess encoding from file extension for text files
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".txt" || ext == ".csv" || ext == ".json" {
		return mimeType, "utf-8" // Common default for text
	}

	return mimeType, "binary"
}

func SpittedName(Filename string) string {
	var spittedName = strings.Split(Filename, ".")
	var filename = time.Now().Format("20060102150405")
	return filename + "." + spittedName[len(spittedName)-1]

}
