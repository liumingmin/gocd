package gocd

import (
	"bytes"
	"context"
	"text/template"

	"github.com/liumingmin/goutils/log"
)

var (
	taskDef      *TaskDef
	taskTemplate *template.Template
)

type ParameterDef struct {
	Name         string
	Description  string
	DefaultValue string
}

type TaskDef struct {
	ParameterDefs []*ParameterDef
	ScriptContent string
}

type NodeTaskDef struct {
	*TaskDef
	HostIp string
}

func getCdTaskConfig(hostIp string) (string, error) {
	nodeTaskDef := &NodeTaskDef{HostIp: hostIp, TaskDef: taskDef}
	buf := new(bytes.Buffer)
	err := taskTemplate.Execute(buf, nodeTaskDef)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func initTaskDef() {
	taskDef = &TaskDef{
		ParameterDefs: []*ParameterDef{
			{
				Name:         "RUN_ENV",
				DefaultValue: "dev",
			},
			{
				Name: "S3GET_URL",
			},
			{
				Name: "S3ENV_VAR",
			},
			{
				Name: "PKG_URL",
			},
			{
				Name: "TARGET_PATH",
			},
			{
				Name: "RUN_CMD",
			},
			{
				Name: "ENV_VAR",
			},
		},
		ScriptContent: defaultTaskScript,
	}
}

func initTaskTpl() {
	tmpl, err := template.New("defaultTaskTpl").Parse(TASK_TPL)
	if err != nil {
		log.Error(context.Background(), "parse tpl failed, err: %v", err)
		return
	}
	taskTemplate = tmpl
}

func init() {
	initTaskTpl()
	initTaskDef()
}

const TASK_TPL = `<?xml version='1.1' encoding='UTF-8'?>
<project>
  <actions/>
  <description></description>
  <keepDependencies>false</keepDependencies>
  <properties>
    <com.sonyericsson.rebuild.RebuildSettings plugin="rebuild@1.31">
      <autoRebuild>false</autoRebuild>
      <rebuildDisabled>false</rebuildDisabled>
    </com.sonyericsson.rebuild.RebuildSettings>
    <hudson.model.ParametersDefinitionProperty>
      <parameterDefinitions>
        {{range .ParameterDefs}}
			<hudson.model.StringParameterDefinition>
			  <name>{{.Name}}</name>
			  <description>{{.Description}}</description>
			  <defaultValue>{{.DefaultValue}}</defaultValue>
			  <trim>true</trim>
			</hudson.model.StringParameterDefinition>
		{{end}}
      </parameterDefinitions>
    </hudson.model.ParametersDefinitionProperty>
  </properties>
  <scm class="hudson.scm.NullSCM"/>
  <assignedNode>{{.HostIp}}</assignedNode>
  <canRoam>false</canRoam>
  <disabled>false</disabled>
  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
  <triggers/>
  <concurrentBuild>false</concurrentBuild>
  <builders>
    <hudson.tasks.Shell>
      <command><![CDATA[{{.ScriptContent}}]]></command>
    </hudson.tasks.Shell>
  </builders>
  <publishers/>
  <buildWrappers/>
</project>`

const defaultTaskScript = `#!/bin/bash -il
#jenkins内置参数
#NODE_NAME

#固定参数
#RUN_ENV 运行环境
#S3GET_URL s3get工具下载地址
#S3ENV_VAR s3get环境变量
#PKG_URL 程序包s3 key
#TARGET_PATH 程序目录
#RUN_CMD 运行脚本
#ENV_VAR 环境变量

#变量
S3GET_PATH="/tmp/s3get"
mkdir -p /tmp

#下载s3工具
if [[ ! -f ${S3GET_PATH} ]]; then
    echo "gocd: downloading s3get..."
     ( flock -x 42;
      if [[ ! -f ${S3GET_PATH} ]]; then
        curl -s --insecure ${S3GET_URL} -o ${S3GET_PATH}.tgz
        tar -xzf ${S3GET_PATH}.tgz -C $(dirname ${S3GET_PATH}.tgz)
        EXIT_CODE=$?

        if [[ EXIT_CODE -ne 0 ]]; then
            echo "gocd: download s3get tgz failed ${S3GET_URL}..."
            rm -f ${S3GET_PATH}.tgz
            rm -f ${S3GET_PATH}
            exit 1
        fi
        chmod +x ${S3GET_PATH}
      fi
     ) 42>"${S3GET_PATH}.lock"
fi


mkdir -p ${TARGET_PATH}
DATENAME=$(date +%Y%m%d%H%M%S-%N)
TMP_PKG_DIR=${TARGET_PATH}/tmppkg${DATENAME}
mkdir ${TMP_PKG_DIR}

#下载程序包
export ${S3ENV_VAR}
${S3GET_PATH} ${PKG_URL} ${TMP_PKG_DIR}.tgz
tar -xzf ${TMP_PKG_DIR}.tgz -C ${TMP_PKG_DIR}
EXIT_CODE=$?
if [[ EXIT_CODE -ne 0 ]]; then
	echo "gocd: download program tgz failed ${PROG_URL}..."
	exit 1
fi
rm -f ${TMP_PKG_DIR}.tgz

#同步程序包
rsync -av ${TMP_PKG_DIR}/  ${TARGET_PATH}
rm -rf ${TMP_PKG_DIR}

cd ${TARGET_PATH}

export ${ENV_VAR}
/bin/bash ${RUN_CMD}
`
