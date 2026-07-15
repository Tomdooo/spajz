package models

type S3LikeDto struct {
	Bucket    string `param:"bucket" validate:"required,min=3,max=63,spajz_bucket"`
	ObjectKey string `param:"*" validate:"required,min=1,max=1024,spajz_objectkey"`
}
