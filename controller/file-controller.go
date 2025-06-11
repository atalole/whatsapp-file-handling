package fileController

import (
	"fmt"
	fileServices "whatsapp_file_handling/services"
	structs "whatsapp_file_handling/structs"

	"github.com/gin-gonic/gin"
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
		mimeType, encoding := fileServices.DetectMimeAndEncoding(buffer, header.Filename)
		detectFileChan <- structs.DetectFile{
			MIMETYPE: mimeType,
			ENCODING: encoding,
		}
	}()
	detectResult := <-detectFileChan

	go func() {
		url, err := fileServices.FileUpload(file, header, detectResult.MIMETYPE, detectResult.ENCODING)
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
