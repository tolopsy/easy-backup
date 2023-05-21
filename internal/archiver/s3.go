package archiver

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Archiver struct {
	Bucket  string
	Session *session.Session
}

func NewS3Archiver(bucket string, session *session.Session) *S3Archiver {
	return &S3Archiver{
		Bucket: bucket,
		Session: session,
	}
}

func (z *S3Archiver) Archive(source, destination string) error {
	zipReader, err := zipFile(source)
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploader(z.Session, func(u *s3manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // Default File Size is 10MB
		u.Concurrency = 2
		u.MaxUploadParts = 10
	})

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(z.Bucket),
		Body:   zipReader,
		Key:    aws.String(destination),
	})

	return err
}
