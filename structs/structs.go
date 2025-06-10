package structs

type UploadResult struct {
	URL string
	Err error
}

type DetectFile struct {
	MIMETYPE string
	ENCODING string
}
