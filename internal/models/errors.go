package models

import "errors"

var (
	ErrBucketNotFound      = errors.New("Bucket not found.")
	ErrBucketAlreadyExists = errors.New("Bucket already exists.")
	ErrBucketNotEmpty      = errors.New("Bucket is not empty.")

	ErrFileNotFound      = errors.New("File not found.")
	ErrFileNotSaved      = errors.New("File not saved.")
	ErrFileAlreadyExists = errors.New("File already exists.")
	ErrNotEnoughSpace    = errors.New("Not enough space.")
	ErrFileTooLarge      = errors.New("File is too large.")

	ErrPresetNotFound    = errors.New("Preset not found.")
	ErrUnsupportedFormat = errors.New("Unsupported format.")

	ErrDatabaseAlreadyExists = errors.New("Database already exists.")
)
