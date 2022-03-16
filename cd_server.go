package gocd

import (
	"context"
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
	jenkins        *gojenkins.Jenkins
	env            string
	defCdNodeParam *CdNodeParam
	s3Info         *CdS3Info
}

type DeployResult struct {
	Status        int
	Result        string
	ConsoleOutput string
}

func NewCdServer(ctx context.Context, url, username, token string, options ...CdServerOption) *CdServer {
	jenkins := gojenkins.CreateJenkins(nil, url, username, token)
	_, err := jenkins.Init(ctx)
	if err != nil {
		log.Error(ctx, "jenkins init failed, err: %v", err)
	}
	cdServer := &CdServer{
		jenkins: jenkins,
		env:     "dev",
	}

	if len(options) > 0 {
		for _, option := range options {
			option(cdServer)
		}
	}

	if cdServer.defCdNodeParam == nil {
		cdServer.defCdNodeParam = NewCdNodeParam()
	}

	return cdServer
}

//最好使用内网IP
func (j *CdServer) CreateNode(ctx context.Context, ip, remark string, options ...CdNodeOption) error {
	cdNodeInfo := j.defCdNodeParam
	if len(options) > 0 {
		cdNodeInfo = &CdNodeParam{}
		*cdNodeInfo = *j.defCdNodeParam

		for _, option := range options {
			option(cdNodeInfo)
		}
	}

	desc := fmt.Sprintf("%v:(%v)%v", j.env, ip, remark)
	node, err := j.jenkins.CreateNode(ctx, ip, cdNodeInfo.numExecutors, desc, cdNodeInfo.remoteFs, ip,
		map[string]string{
			"method":        "SSHLauncher",
			"host":          ip,
			"port":          cdNodeInfo.sshPort,
			"credentialsId": cdNodeInfo.credentialsId,
			"jvmOptions":    cdNodeInfo.jvmOptions,
		})
	if err != nil {
		log.Error(ctx, "CreateNode failed, err: %v", err)
		return err
	}
	log.Info(ctx, "CreateNode: %v", node)
	return nil
}

func (j *CdServer) DeleteNode(ctx context.Context, ip string) (bool, error) {
	node, err := j.jenkins.GetNode(ctx, ip)
	if err != nil {
		return false, err
	}

	return node.Delete(ctx)
}

func (j *CdServer) GetAllNodes(ctx context.Context) ([]*gojenkins.Node, error) {
	nodes, err := j.jenkins.GetAllNodes(ctx)

	envNodes := make([]*gojenkins.Node, 0, len(nodes))
	for _, node := range nodes {
		if !strings.HasPrefix(node.Raw.Description, j.env) {
			continue
		}

		envNodes = append(envNodes, node)
	}

	return envNodes, err
}

func (j *CdServer) GetNode(ctx context.Context, ip string) (*gojenkins.Node, error) {
	return j.jenkins.GetNode(ctx, ip)
}

func (j *CdServer) getOrCreateJob(ctx context.Context, service *CdService, ip string) (*gojenkins.Job, error) {
	jobName := fmt.Sprintf("%v-%v-%v-%v", service.cdScript.scriptVersion, j.env, service.Name(), ip)

	job, err := j.jenkins.GetJob(ctx, jobName)
	if err != nil {
		taskConfig, err := service.GetCdTaskScriptConfig(ip)
		//fmt.Println(taskConfig)
		if err != nil {
			return nil, err
		}

		_, err = j.jenkins.CreateJob(ctx, taskConfig, jobName)
		if err != nil {
			log.Error(ctx, "CreateJob failed: %v, err: %v", jobName, err)
			return nil, err
		}
		time.Sleep(time.Second * 3)

		job, err = j.jenkins.GetJob(ctx, jobName)
		if err != nil {
			log.Error(ctx, "GetJob failed: %v, err: %v", jobName, err)
			return nil, err
		}
	}
	return job, nil
}

//程序运行配置中，抽提db信息放到环境变量中运行时传递
//不同环境的配置文件直接写入程序包,动态内容使用环境变量设置
func (j *CdServer) Deploy(ctx context.Context, service *CdService, ip string) (int64, error) {
	job, err := j.getOrCreateJob(ctx, service, ip)
	if err != nil {
		return 0, err
	}

	//s3get env
	var s3EnvsStr strings.Builder
	for key, value := range j.s3Info.envVar() {
		s3EnvsStr.WriteString(fmt.Sprintf(" %v=%v", key, value))
	}

	//动态参数
	envVar := service.EnvVar()
	var envsStr strings.Builder
	for key, value := range envVar {
		envsStr.WriteString(fmt.Sprintf(" %v=%v", key, value))
	}

	params := map[string]string{
		"RUN_ENV":     j.env,
		"HOST_IP":     ip,
		"S3GET_URL":   j.s3Info.s3GetToolUrl,
		"S3ENV_VAR":   s3EnvsStr.String(),
		"PKG_URL":     service.PkgUrl(), //s3get download package
		"TARGET_PATH": service.TargetPath(),
		"RUN_CMD":     service.RunCmd(),
		"ENV_VAR":     envsStr.String(),
	}

	taskId, err := job.InvokeSimple(ctx, params)
	if err != nil {
		log.Error(ctx, "job build failed: %v", err)
		return 0, err
	}

	return taskId, nil
}

func (j *CdServer) GetDeployResult(ctx context.Context, name, ip string, taskId int64) (*DeployResult, error) {
	jobName := fmt.Sprintf("%v-%v-%v", j.env, name, ip)

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
