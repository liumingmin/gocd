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

type DeployServer struct {
	jenkins        *gojenkins.Jenkins
	env            string
	credentialsId  string
	jvmOptions     string
	numExecutors   int
	remoteFs       string
	fetchBuildTime time.Duration
}

type DeployParam struct {
	PkgUrl     string
	TargetPath string
	RunCmd     string
	EnvVar     map[string]string
}

type DeployResult struct {
	Status        int
	Result        string
	ConsoleOutput string
}

func NewDeployServer(ctx context.Context, env, url, username, token string) *DeployServer {
	jenkins := gojenkins.CreateJenkins(nil, url, username, token)
	_, err := jenkins.Init(ctx)
	if err != nil {
		log.Error(ctx, "jenkins init failed, err: %v", err)
	}
	return &DeployServer{
		jenkins:        jenkins,
		env:            env,
		numExecutors:   2,
		jvmOptions:     "-Xms16m -Xmx64m",
		remoteFs:       "/var/lib/jenkins",
		fetchBuildTime: time.Second,
	}
}

func (j *DeployServer) SetCredentialsId(credentialsId string) {
	j.credentialsId = credentialsId
}

func (j *DeployServer) SetJvmOptions(jvmOptions string) {
	j.jvmOptions = jvmOptions
}

func (j *DeployServer) SetNumExecutors(numExecutors int) {
	j.numExecutors = numExecutors
}

//最好使用内网IP
func (j *DeployServer) CreateNode(ctx context.Context, ip, sshPort, remark string) error {
	desc := fmt.Sprintf("%v:(%v)%v", j.env, ip, remark)
	node, err := j.jenkins.CreateNode(ctx, ip, j.numExecutors, desc, j.remoteFs, ip,
		map[string]string{
			"method":        "SSHLauncher",
			"host":          ip,
			"port":          sshPort,
			"credentialsId": j.credentialsId,
			"jvmOptions":    j.jvmOptions,
		})
	if err != nil {
		log.Error(ctx, "CreateNode failed, err: %v", err)
	}
	log.Info(ctx, "CreateNode: %v", node)

	return err
}

//todo
func (j *DeployServer) GetNode(ctx context.Context, ip string) (*gojenkins.Node, error) {
	return j.jenkins.GetNode(ctx, ip)
}

func (j *DeployServer) getOrCreateJob(ctx context.Context, name, ip string) (*gojenkins.Job, error) {
	jobName := fmt.Sprintf("%v-%v-%v", j.env, name, ip)

	job, err := j.jenkins.GetJob(ctx, jobName)
	if err != nil {
		jobConfig := strings.Replace(DEFAULT_JOB_TPL, "$$SCRIPT_CONTENT$$", JOB_BASE_SCRIPT, -1)
		jobConfig = strings.Replace(jobConfig, "$$HOST_IP$$", ip, -1)
		_, err = j.jenkins.CreateJob(ctx, jobConfig, jobName)
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
func (j *DeployServer) Deploy(ctx context.Context, name, ip string, param *DeployParam) (int64, error) {
	job, err := j.getOrCreateJob(ctx, name, ip)
	if err != nil {
		return 0, err
	}

	var envsStr strings.Builder
	for key, value := range param.EnvVar {
		envsStr.WriteString(fmt.Sprintf(" %v=%v", key, value))
	}

	params := map[string]string{
		"RUN_ENV":     j.env,
		"HOST_IP":     ip,
		"PKG_URL":     param.PkgUrl,
		"TARGET_PATH": param.TargetPath,
		"RUN_CMD":     param.RunCmd,
		"ENV_VAR":     envsStr.String(),
	}

	taskId, err := job.InvokeSimple(ctx, params)
	if err != nil {
		log.Error(ctx, "job build failed: %v", err)
		return 0, err
	}

	return taskId, nil
}

func (j *DeployServer) GetDeployResult(ctx context.Context, name, ip string, taskId int64) (*DeployResult, error) {
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
