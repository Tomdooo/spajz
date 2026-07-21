package models

const (
	CodeBucketNotFound      = "BUCKET_NOT_FOUND"
	CodeBucketAlreadyExists = "BUCKET_ALREADY_EXISTS"
	CodeBucketNotEmpty      = "BUCKET_NOT_EMPTY"

	CodeFileNotFound       = "FILE_NOT_FOUND"
	CodeFileNotSaved       = "FILE_NOT_SAVED"
	CodeFileAlreadyExists  = "FILE_ALREADY_EXISTS"
	CodeNotEnoughSpace     = "NOT_ENOUGHT_SPACE"
	CodeFileTooLarge       = "FILE_TOO_LARGE"
	CodeFileNotProcessable = "FILE_NOT_PROCESSABLE"

	CodePresetNotFound      = "PRESET_NOT_FOUND"
	CodePresetFormatInvalid = "PRESET_FORMAT_INVALID"
	CodeUnsupportedFormat   = "UNSUPPORTED_FORMAT"
	CodeImageNotProcessable = "IMAGE_NOT_PROCESSABLE"

	CodeDatabaseAlreadyExists = "DATABASE_ALREADY_EXISTS"
	CodeInvalidURLFormat      = "INVALID_URL_FORMAT"
)
