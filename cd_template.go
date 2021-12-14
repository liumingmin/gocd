package gocd

const DEFAULT_JOB_TPL = `<?xml version='1.1' encoding='UTF-8'?>
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
        <hudson.model.StringParameterDefinition>
          <name>RUN_ENV</name>
          <description></description>
          <defaultValue>dev</defaultValue>
          <trim>false</trim>
        </hudson.model.StringParameterDefinition>
        <hudson.model.StringParameterDefinition>
          <name>HOST_IP</name>
          <description></description>
          <defaultValue>nil</defaultValue>
          <trim>false</trim>
        </hudson.model.StringParameterDefinition>
        <hudson.model.StringParameterDefinition>
          <name>PKG_URL</name>
          <description></description>
          <defaultValue></defaultValue>
          <trim>false</trim>
        </hudson.model.StringParameterDefinition>
		<hudson.model.StringParameterDefinition>
          <name>TARGET_PATH</name>
          <description></description>
          <defaultValue></defaultValue>
          <trim>false</trim>
        </hudson.model.StringParameterDefinition>
		<hudson.model.StringParameterDefinition>
          <name>RUN_CMD</name>
          <description></description>
          <defaultValue></defaultValue>
          <trim>false</trim>
        </hudson.model.StringParameterDefinition>
		<hudson.model.StringParameterDefinition>
          <name>ENV_VAR</name>
          <description></description>
          <defaultValue></defaultValue>
          <trim>false</trim>
        </hudson.model.StringParameterDefinition>
      </parameterDefinitions>
    </hudson.model.ParametersDefinitionProperty>
  </properties>
  <scm class="hudson.scm.NullSCM"/>
  <assignedNode>$$HOST_IP$$</assignedNode>
  <canRoam>false</canRoam>
  <disabled>false</disabled>
  <blockBuildWhenDownstreamBuilding>false</blockBuildWhenDownstreamBuilding>
  <blockBuildWhenUpstreamBuilding>false</blockBuildWhenUpstreamBuilding>
  <triggers/>
  <concurrentBuild>false</concurrentBuild>
  <builders>
    <hudson.tasks.Shell>
      <command><![CDATA[$$SCRIPT_CONTENT$$]]></command>
    </hudson.tasks.Shell>
  </builders>
  <publishers/>
  <buildWrappers/>
</project>`

const JOB_BASE_SCRIPT = `#!/bin/bash -il
#jenkins内置参数
#NODE_NAME
#WORKSPACE

#固定参数
#RUN_ENV 运行环境
#HOST_IP 目标机器IP
#PKG_URL 程序包远程地址
#TARGET_PATH 节点程序目录
#RUN_CMD 运行脚本
#ENV_VAR 环境变量

if [[ "${HOST_IP}" != "${NODE_NAME}" ]];then
   echo "jenkins: ${HOST_IP} not match agent node name ${NODE_NAME}..."
   exit 1
fi

DATENAME=$(date +%Y%m%d%H%M%S-%N)
PKG_PATH="${WORKSPACE}/${RANDOM}-${DATENAME}.tgz"
RUN_PATH="${TARGET_PATH}/run.sh"

mkdir -p ${WORKSPACE}
mkdir -p ${TARGET_PATH}

curl -s ${PKG_URL} -o ${PKG_PATH}

if [[ ! -f "${PKG_PATH}" ]]; then
   echo "jenkins: not found pkg ${PKG_PATH}..."
   exit 1
fi

tar -zxf  ${PKG_PATH}  -C ${TARGET_PATH}
EXIT_CODE=$?
if [[ EXIT_CODE -ne 0 ]]; then
  echo "jenkins: tar xzf failed ${PKG_PATH} ${TARGET_PATH}..."
  echo "jenkins: clear pkg ${PKG_PATH}..."
  rm -f ${PKG_PATH}
  exit 1
fi

rm -f ${PKG_PATH}

export ${ENV_VAR}

cd ${TARGET_PATH}

/bin/bash ${RUN_CMD}
`
