package gocd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/liumingmin/gojenkins"
	"github.com/liumingmin/goutils/log"
)

const (
	RUN_STATUS_RUNNING = 1
	RUN_STATUS_FINISH  = 2
	RUN_STATUS_ERR     = 3
)

type CdServer struct {
	jenkins *gojenkins.Jenkins
	env     string
	s3Info  *CdS3Info

	nodeBroker *CdNodeBroker
}

type DeployResult struct {
	Status        int
	Result        string
	ConsoleOutput string
}

func NewCdServer(ctx context.Context, url, username, token, env string, options ...CdServerOption) *CdServer {
	jenkins := gojenkins.CreateJenkins(nil, url, username, token)
	_, err := jenkins.Init(ctx)
	if err != nil {
		log.Error(ctx, "jenkins init failed, err: %v", err)
	}
	cdServer := &CdServer{
		jenkins:    jenkins,
		env:        env,
		nodeBroker: NewCdNodeBroker(jenkins, env, nil),
	}

	if len(options) > 0 {
		for _, option := range options {
			option(cdServer)
		}
	}

	return cdServer
}

func (j *CdServer) GetNodeBroker() *CdNodeBroker {
	return j.nodeBroker
}

func (j *CdServer) getOrCreateJob(ctx context.Context, service CdService, node *gojenkins.Node) (string, *gojenkins.Job, error) {
	idx := int64(service.IncDeployCounter()) % node.Raw.NumExecutors
	jobName := fmt.Sprintf("%v-%v-%v-%v-%v", service.GetCdScript().scriptVersion, j.env, service.GetName(), node.GetName(), idx)

	job, err := j.jenkins.GetJob(ctx, jobName)
	if err != nil || job == nil {
		taskConfig, err := service.GetCdScript().GetCdTaskScriptConfig(node.GetName())
		//fmt.Println(taskConfig)
		if err != nil {
			return jobName, nil, err
		}

		_, err = j.jenkins.CreateJob(ctx, taskConfig, jobName)
		if err != nil {
			log.Error(ctx, "CreateJob failed: %v, err: %v", jobName, err)
			return jobName, nil, err
		}

		for i := 0; i < 3; i++ {
			job, err = j.jenkins.GetJob(ctx, jobName)
			if err != nil || job == nil {
				log.Debug(ctx, "GetJob failed: %v, err: %v", jobName, err)
				time.Sleep(time.Second)
				continue
			}

			log.Info(ctx, "GetJob ok: %v", jobName)
			break
		}
	}
	return jobName, job, nil
}

func (j *CdServer) DeploySimple(ctx context.Context, service CdService, nodeName string) (string, int64, error) {
	node := j.nodeBroker.GetNodeByName(nodeName)
	if node == nil {
		return "", 0, errors.New("not found node")
	}

	return j.deploy(ctx, service, node)
}

func (j *CdServer) deploy(ctx context.Context, service CdService, node *gojenkins.Node) (string, int64, error) {
	jobName, job, err := j.getOrCreateJob(ctx, service, node)
	if err != nil {
		return jobName, 0, err
	}

	//s3get env
	var s3EnvsStr strings.Builder
	for key, value := range j.s3Info.envVar() {
		s3EnvsStr.WriteString(fmt.Sprintf(" %v=%v", key, value))
	}

	params := map[string]string{
		"RUN_ENV":   j.env,
		"S3GET_URL": j.s3Info.s3GetToolUrl,
		"S3ENV_VAR": s3EnvsStr.String(),
	}

	// service generate svc params
	svcParams := service.GetParams()
	for k, v := range svcParams {
		params[k] = v
	}

	taskId, err := job.InvokeSimple(ctx, params)
	if err != nil {
		log.Error(ctx, "job build failed: %v", err)
		return jobName, 0, err
	}

	return jobName, taskId, nil
}

func (j *CdServer) GetDeployResult(ctx context.Context, jobName string, taskId int64) (*DeployResult, error) {
	job, err := j.jenkins.GetJob(ctx, jobName)
	if err != nil {
		log.Error(ctx, "get job from jenkins failed: %v, err: %v", jobName, err)
		return nil, err
	}

	build, err := j.jenkins.GetBuildFromQueueID(ctx, job, taskId)
	if err != nil || build == nil {
		log.Error(ctx, "get build from jenkins failed: %v, err: %v", taskId, err)
		return nil, err
	}

	status := RUN_STATUS_RUNNING
	if !build.IsRunning(ctx) {
		if build.IsGood(ctx) {
			status = RUN_STATUS_FINISH
		} else {
			status = RUN_STATUS_ERR
		}
	}

	taskBuild := &DeployResult{
		Status:        status,
		Result:        build.GetResult(),
		ConsoleOutput: build.GetConsoleOutput(ctx),
	}

	log.Info(ctx, "get build result from jenkins  %v", taskBuild)
	return taskBuild, nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type CdServerOption func(*CdServer)

func CdServerNodeOption(options ...CdNodeOption) CdServerOption {
	return func(server *CdServer) {
		server.nodeBroker.SetDefCdNodeParam(NewCdNodeParam(options...))
	}
}

func CdServerS3Option(s3AK, s3SK, s3Endpoint, s3Bucket, s3Region, s3getToolUrl string) CdServerOption {
	return func(server *CdServer) {
		server.s3Info = NewCdS3Info(s3AK, s3SK, s3Endpoint, s3Bucket, s3Region, s3getToolUrl)
	}
}
