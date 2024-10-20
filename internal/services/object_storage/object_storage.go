package object_storage

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
)

type ObjectStorage struct {
	AwsSession *session.Session
	Ctx        context.Context
}

func New() *ObjectStorage {
	return &ObjectStorage{
		AwsSession: session.Must(session.NewSession()),
		Ctx:        context.Background(),
	}
}

func (o *ObjectStorage) AddText(id, text string) error {
	// TODO: Creating a new text file where the text is put
	// After that the path to the .txt file is given to Upload function
	// .txt file is deleted
	f, err := os.CreateTemp("D:/pastebin/internal/services/object_storage/texts/", id+".txt")
	if err != nil {
		return fmt.Errorf("couldn't create a file for provided text:%w", err)
	}
	defer f.Close()

	fileName := f.Name()[51 : len(f.Name())-10]
	if string(fileName[len(fileName)-1]) == "x" {
		fileName += "t"
	}

	err = os.WriteFile(f.Name(), []byte(text), 0666)
	if err != nil {
		return fmt.Errorf("couldn't write text to the file:%w", err)
	}

	uploader := s3manager.NewUploader(o.AwsSession)

	if _, err := uploader.UploadWithContext(o.Ctx, &s3manager.UploadInput{
		Bucket: aws.String("pastebin-storage"),
		Key:    aws.String(fileName),
		Body:   f,
	}); err != nil {
		return fmt.Errorf("couldn't upload the file to storage:%w", err)
	}

	f.Close()
	err = os.Remove(f.Name())
	if err != nil {
		return fmt.Errorf("couldn't remove created text file:%w", err)
	}

	return nil
}

func (o *ObjectStorage) Text(textUrl string) (string, error) {
	downloader := s3manager.NewDownloader(o.AwsSession)

	f, err := os.CreateTemp("D:/pastebin/internal/services/object_storage/texts/", textUrl)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := downloader.DownloadWithContext(o.Ctx, f, &s3.GetObjectInput{
		Bucket: aws.String("pastebin-storage"),
		Key:    aws.String(textUrl + ".txt"),
	}); err != nil {
		return "", fmt.Errorf("couldn't download the file from object storage:%w", err)
	}

	fileInfo, err := f.Stat()
	if err != nil {
		return "", fmt.Errorf("couldn't download the file from object storage:%w", err)
	}

	text := make([]byte, fileInfo.Size())

	if _, err := f.Read(text); err != nil {
		return "", fmt.Errorf("couldn't read the text from a file:%w", err)
	}

	f.Close()
	if err := os.Remove(f.Name()); err != nil {
		return "", fmt.Errorf("couldn't delete the file:%w", err)
	}

	return string(text), nil
}
