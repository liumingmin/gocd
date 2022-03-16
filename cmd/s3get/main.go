package main

//GOOS=linux GOARCH=amd64 go build -v --tags netgo -ldflags '-s -w -extldflags "-static"' -o s3get main.go
//tar -czf s3get.tgz s3get
import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	s3AK       = os.Getenv("GOCD_S3_AK")
	s3SK       = os.Getenv("GOCD_S3_SK")
	s3Endpoint = os.Getenv("GOCD_S3_ENDPOINT")
	s3Bucket   = os.Getenv("GOCD_S3_BUCKET")
	s3Region   = os.Getenv("GOCD_S3_REGION")
)

func usage() {
	fmt.Println("Usage: s3get S3_FILENAME LOCAL_FILENAME")
}

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}
	s3Filename := os.Args[1]
	localFilename := os.Args[2]

	if len(s3Filename) == 0 || len(localFilename) == 0 {
		usage()
		os.Exit(1)
	}

	sess, _ := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(s3AK, s3SK, ""),
		Region:           aws.String(s3Region),
		Endpoint:         aws.String(s3Endpoint),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	},
	)

	downloader := s3manager.NewDownloader(sess)
	file, err := os.Create(localFilename)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(s3Bucket),
			Key:    aws.String(s3Filename),
		})
	if err != nil {
		fmt.Println(err)
		return
	}
}
