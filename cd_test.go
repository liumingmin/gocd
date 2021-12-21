package gocd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func getJServer() *CdServer {
	//http://127.0.0.1:8091/user/admin/configure 获取apitoken
	return NewCdServer(context.Background(), "http://127.0.0.1:8091/", "admin", "116e908012f0e76a71c788619470681f83",
		CdServerEnvOption("prod"), CdServerS3Option(
			"Vg6p9p/WM55ZbiZkE8Vyzw==",
			"r0yRc7Yxc0fB7yWRoaWJrvLlC3hShtqBFfqj13PKTLo=",
			"http://localhost:9005",
			"test",
			"zh-south-1",
		))
}

func TestDeploy(t *testing.T) {
	//todo download s3get
	taskId, _ := getJServer().Deploy(context.Background(), "test", "master", &DeployParam{
		PkgUrl:     "pkg.tgz", //http://10.11.244.107/pkg.tgz
		TargetPath: "/tmp/test",
		RunCmd:     "run.sh",
		EnvVar: map[string]string{
			"A": "1",
			"B": "2",
			"C": "3",
		},
	})

	t.Log(taskId)
	//server.getOrCreateJob(context.Background(), "test", "master")
}

func TestGetJobRaw(t *testing.T) {
	server := NewCdServer(context.Background(), "http://127.0.0.1:8091/", "admin", "admin",
		CdServerEnvOption("prod"))
	job, _ := server.getOrCreateJob(context.Background(), "test", "master")
	fmt.Println(job.GetConfig(context.Background()))
}

func TestGetTaskBuild(t *testing.T) {
	build, _ := getJServer().GetDeployResult(context.Background(), "test", "master", 8)
	bs, _ := json.Marshal(build)
	t.Log(string(bs))
}

func TestS3Get(t *testing.T) {
	sess, _ := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials("Vg6p9p/WM55ZbiZkE8Vyzw==",
			"r0yRc7Yxc0fB7yWRoaWJrvLlC3hShtqBFfqj13PKTLo=", ""),
		Region:           aws.String("zh-south-1"),
		Endpoint:         aws.String("http://localhost:9005"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	},
	)

	downloader := s3manager.NewDownloader(sess)
	file, err := os.Create("./pkg.tgz")
	if err != nil {
		t.Log(err)
		return
	}
	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String("test"),
			Key:    aws.String("pkg.tgz"),
		})
	if err != nil {
		t.Log(err)
		return
	}

	//ioutil.WriteFile("./pkg.tgz", buffer.Bytes(), 0666)
}
