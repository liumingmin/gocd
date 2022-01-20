package gocd

type CdNodeParam struct {
	credentialsId string
	jvmOptions    string
	numExecutors  int
	remoteFs      string
	sshPort       string
}

func NewCdNodeParam(options ...CdNodeOption) *CdNodeParam {
	cdNodeInfo := &CdNodeParam{
		numExecutors: 1,
		jvmOptions:   "-Xms16m -Xmx64m",
		remoteFs:     "/var/lib/jenkins",
		sshPort:      "22",
	}
	if len(options) > 0 {
		for _, option := range options {
			option(cdNodeInfo)
		}
	}
	return cdNodeInfo
}

type CdNodeOption func(*CdNodeParam)

func CdNodeCredIdOption(credId string) CdNodeOption {
	return func(nodeInfo *CdNodeParam) {
		nodeInfo.credentialsId = credId
	}
}

func CdNodeJvmOption(jvm string) CdNodeOption {
	return func(nodeInfo *CdNodeParam) {
		nodeInfo.jvmOptions = jvm
	}
}

func CdNodeNumExecutorsOption(numExecutors int) CdNodeOption {
	return func(nodeInfo *CdNodeParam) {
		nodeInfo.numExecutors = numExecutors
	}
}

func CdNodeRemoteFsOption(remoteFs string) CdNodeOption {
	return func(nodeInfo *CdNodeParam) {
		nodeInfo.remoteFs = remoteFs
	}
}

func CdNodeSshPortOption(sshPort string) CdNodeOption {
	return func(nodeInfo *CdNodeParam) {
		nodeInfo.sshPort = sshPort
	}
}
