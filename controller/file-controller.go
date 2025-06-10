package fileController

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"whatsapp_file_handling/structs"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
)

var (
	bucketName = "quick-review"
	region     = "us-east-1" // e.g., "ap-south-1"
)

func UploadFileHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		fmt.Println("Error while reading file:", err)

		c.JSON(400, gin.H{"message": "file not found", "status": 400})
		return
	}
	defer file.Close()

	// Read the first 512 bytes to detect MIME type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "failed to read file",
			"status":  500,
			"error":   err.Error(),
		})
		return
	}

	// Reset the file pointer after reading
	if _, err := file.Seek(0, 0); err != nil {
		c.JSON(500, gin.H{
			"message": "failed to reset file pointer",
			"status":  500,
			"error":   err.Error(),
		})
		return
	}

	// here handle goroutine async operation
	fileUploadChan := make(chan structs.UploadResult)
	detectFileChan := make(chan structs.DetectFile)
	// Detect MIME type

	go func() {
		mimeType, encoding := detectMimeAndEncoding(buffer, header.Filename)
		detectFileChan <- structs.DetectFile{
			MIMETYPE: mimeType,
			ENCODING: encoding,
		}
	}()
	detectResult := <-detectFileChan

	go func() {
		url, err := FileUpload(file, header, detectResult.MIMETYPE, detectResult.ENCODING)
		fileUploadChan <- structs.UploadResult{
			URL: url,
			Err: err,
		}
	}()

	uploadResult := <-fileUploadChan

	if uploadResult.Err != nil {
		c.JSON(500, gin.H{"error": "upload failed"})
		return
	}

	c.JSON(200, gin.H{"file_url": uploadResult.URL})
}

func FileUpload(file multipart.File, fileHeader *multipart.FileHeader, mimeType string, encoding string) (string, error) {

	ctx := context.TODO()

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("", "", "")))
	fmt.Println("cfg:", cfg)

	if err != nil {
		log.Printf("unable to load AWS SDK config: %v", err)
		return "", err
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	// Generate unique file name
	key := fmt.Sprintf("uploads/%d_%s", time.Now().Unix(), fileHeader.Filename)

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

func detectMimeAndEncoding(buffer []byte, filename string) (string, string) {
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
