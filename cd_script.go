package gocd

import (
	"bytes"
	"context"
	"text/template"

	"github.com/liumingmin/goutils/log"
)

type CdScriptParamDef struct {
	Name         string
	Description  string
	DefaultValue string
}

type CdScript struct {
	scriptParamDefs []*CdScriptParamDef
	scriptTemplate  *template.Template

	scriptContent string
	scriptVersion int
}

type cdScriptInstance struct {
	ParameterDefs []*CdScriptParamDef
	ScriptContent string
	HostIp        string
}

func (t *CdScript) GetCdTaskScriptConfig(hostIp string) (string, error) {
	nodeTaskDef := &cdScriptInstance{HostIp: hostIp, ParameterDefs: t.scriptParamDefs, ScriptContent: t.scriptContent}
	buf := new(bytes.Buffer)
	err := t.scriptTemplate.Execute(buf, nodeTaskDef)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func NewCdScript(scriptParamDefs []*CdScriptParamDef, scriptXmlTpl, scriptContent string, scriptVersion int) *CdScript {
	tmpl, err := template.New("defaultTaskTpl").Parse(scriptXmlTpl)
	if err != nil {
		log.Error(context.Background(), "parse tpl failed, err: %v", err)
		return nil
	}

	baseScriptParamDefs := make([]*CdScriptParamDef, 0)
	baseScriptParamDefs = append(baseScriptParamDefs, &CdScriptParamDef{
		Name:         "RUN_ENV",
		DefaultValue: "dev",
	})
	baseScriptParamDefs = append(baseScriptParamDefs, &CdScriptParamDef{
		Name: "S3GET_URL",
	})
	baseScriptParamDefs = append(baseScriptParamDefs, &CdScriptParamDef{
		Name: "S3ENV_VAR",
	})

	return &CdScript{
		scriptParamDefs: append(baseScriptParamDefs, scriptParamDefs...),
		scriptTemplate:  tmpl,
		scriptContent:   scriptContent,
		scriptVersion:   scriptVersion,
	}
}

func NewDefaultCdScript() *CdScript {
	scriptParamDefs := make([]*CdScriptParamDef, 0)
	paramNames := []string{"PKG_URL", "TARGET_PATH", "RUN_CMD", "ENV_VAR"}
	for _, paramName := range paramNames {
		scriptParamDefs = append(scriptParamDefs, &CdScriptParamDef{
			Name: paramName,
		})
	}
	return NewCdScript(scriptParamDefs, DefaultXmlTpl, DefaultTaskScript, defaultTaskScriptVer)
}

const DefaultXmlTpl = `<?xml version='1.1' encoding='UTF-8'?>
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

const defaultTaskScriptVer = 1

const DefaultTaskScript = `#!/bin/bash -il
#jenkins????????????
#NODE_NAME

#????????????
#RUN_ENV ????????????
#S3GET_URL s3get??????????????????
#S3ENV_VAR s3get????????????

#????????????
#PKG_URL ?????????s3 key
#TARGET_PATH ????????????
#RUN_CMD ????????????
#ENV_VAR ????????????

#??????
S3GET_PATH="/tmp/s3get"
mkdir -p /tmp

#??????s3??????
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

#???????????????
export ${S3ENV_VAR}
${S3GET_PATH} ${PKG_URL} ${TMP_PKG_DIR}.tgz
tar -xzf ${TMP_PKG_DIR}.tgz -C ${TMP_PKG_DIR}
EXIT_CODE=$?
if [[ EXIT_CODE -ne 0 ]]; then
	echo "gocd: download program tgz failed ${PROG_URL}..."
	exit 1
fi
rm -f ${TMP_PKG_DIR}.tgz

#???????????????
rsync -av ${TMP_PKG_DIR}/  ${TARGET_PATH}
rm -rf ${TMP_PKG_DIR}

cd ${TARGET_PATH}

export ${ENV_VAR}
/bin/bash ${RUN_CMD}
`
