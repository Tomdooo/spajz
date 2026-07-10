package models

import "path/filepath"

type FileRequestContext struct {
	Bucket     string
	ObjectKey  string
	ObjectHash string // hash spočítaný jednou v handleru
	Filename   string
}

func NewFileRequestContext(bucket, objectKey, objectHash string) *FileRequestContext {
	return &FileRequestContext{
		Bucket:     bucket,
		ObjectKey:  objectKey,
		ObjectHash: objectHash,
		Filename:   filepath.Base(objectKey),
	}
}
