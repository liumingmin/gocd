package gocd

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func getJServer() *CdServer {
	//http://127.0.0.1:8091/user/admin/configure 获取apitoken
	return NewCdServer(context.Background(), "http://127.0.0.1:8091/", "admin", "116e908012f0e76a71c788619470681f83",
		NewCdNodeInfo(), CdServerEnvOption("dev"))
}

func TestDeploy(t *testing.T) {
	taskId, _ := getJServer().Deploy(context.Background(), "test", "master", &DeployParam{
		PkgUrl:     "http://10.11.244.107/pkg.tgz",
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
		NewCdNodeInfo(), CdServerEnvOption("dev"))
	job, _ := server.getOrCreateJob(context.Background(), "test", "master")
	fmt.Println(job.GetConfig(context.Background()))
}

func TestGetTaskBuild(t *testing.T) {
	build, _ := getJServer().GetDeployResult(context.Background(), "test", "master", 8)
	bs, _ := json.Marshal(build)
	t.Log(string(bs))
}
