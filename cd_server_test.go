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

const testLocalIp = "192.168.0.104" //
const testUsername = "admin"
const testToken = "116e908012f0e76a71c788619470681f83" //admin

func getCdServer() *CdServer {
	//http://127.0.0.1:8091/user/admin/configure 获取apitoken
	cdServer := NewCdServer(context.Background(), "http://"+testLocalIp+":8091/", testUsername, testToken, "prod",
		CdServerS3Option(
			"Vg6p9p/WM55ZbiZkE8Vyzw==",
			"r0yRc7Yxc0fB7yWRoaWJrvLlC3hShtqBFfqj13PKTLo=",
			"http://"+testLocalIp+":9005",
			"test",
			"zh-south-1",
			"http://"+testLocalIp+"/s3get.tgz", //s3get工具http下载地址
		))

	simpleSvc := NewDefaultCdService("runit", "pkg.tgz", "/tmp/test", "run.sh", map[string]string{
		"A": "1",
		"B": "2",
		"C": "3",
	})
	cdServer.AddService(simpleSvc)
	return cdServer
}

func TestGetNodes(t *testing.T) {
	nodes, err := getCdServer().GetNodeBroker().getAllNodes(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, node := range nodes {
		t.Log(node)
	}
}

func TestCreateNode(t *testing.T) {
	err := getCdServer().GetNodeBroker().CreateNode(context.Background(), "172.17.0.4", "172.17.0.4",
		CdNodeCredIdOption("defssh"), CdNodeNumExecutorsOption(5))
	t.Log(err)
}

func TestDeleteNode(t *testing.T) {
	ok, err := getCdServer().GetNodeBroker().DeleteNode(context.Background(), "172.17.0.4")
	t.Log(ok, err)
}

func TestDeploy(t *testing.T) {
	jserver := getCdServer()

	for i := 0; i < 4; i++ {
		jobName, taskId, _ := jserver.DeploySimple(context.Background(), "runit", "172.17.0.4") //172.17.0.3

		fmt.Println(jobName, taskId)

		//time.Sleep(time.Second)
	}
}

func TestGetTaskBuild(t *testing.T) {
	build, _ := getCdServer().GetDeployResult(context.Background(), "1-prod-runit-172.17.0.4-0", 131)
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
