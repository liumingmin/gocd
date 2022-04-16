package gocd

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const testLocalIp = "192.168.0.112"
const testUsername = "admin"
const testToken = "116e908012f0e76a71c788619470681f83" //admin

func getJServer() *CdServer {
	//http://127.0.0.1:8091/user/admin/configure 获取apitoken
	return NewCdServer(context.Background(), "http://"+testLocalIp+":8091/", testUsername, testToken,
		CdServerEnvOption("prod"), CdServerS3Option(
			"Vg6p9p/WM55ZbiZkE8Vyzw==",
			"r0yRc7Yxc0fB7yWRoaWJrvLlC3hShtqBFfqj13PKTLo=",
			"http://"+testLocalIp+":9005",
			"test",
			"zh-south-1",
			"http://"+testLocalIp+"/s3get.tgz", //s3get工具http下载地址
		))
}

func TestCreateNode(t *testing.T) {
	err := getJServer().CreateNode(context.Background(), "172.17.0.3", "172.17.0.3_2", CdNodeCredIdOption("defssh"))
	t.Log(err)
}

func TestDeploy(t *testing.T) {
	service := NewDefaultCdService("test", "pkg.tgz", "/tmp/test", "run.sh", map[string]string{
		"A": "1",
		"B": "2",
		"C": "3",
	})
	taskId, _ := getJServer().Deploy(context.Background(), service, "172.17.0.3")

	t.Log(taskId)

}

func TestGetTaskBuild(t *testing.T) {
	build, _ := getJServer().GetDeployResult(context.Background(), "test", "master", 25)
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

func TestNewDefaultCdScript(t *testing.T) {
	cdScript := NewDefaultCdScript()
	scriptConfig, _ := cdScript.GetCdTaskScriptConfig("127.0.0.1")
	t.Log(scriptConfig)
}
