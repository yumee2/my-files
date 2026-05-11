package dto

import "io"

type UploadDTO struct {
	Name string
	Body io.Reader
}
