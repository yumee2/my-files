package models

type File struct {
	ID           string `json:"id"`
	OriginalName string `json:"name"`
	Size         int64  `json:"size"`
	MimeType     string `json:"type"`
}
