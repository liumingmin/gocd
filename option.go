package gocd

type CdServerOption func(*CdServer)

func CdServerEnvOption(env string) CdServerOption {
	return func(server *CdServer) {
		server.env = env
	}
}

func CdServerNodeOption(options ...CdNodeOption) CdServerOption {
	return func(server *CdServer) {
		server.defCdNodeInfo = NewCdNodeInfo(options...)
	}
}

func CdServerS3Option(s3AK, s3SK, s3Endpoint, s3Bucket, s3Region, s3getToolUrl string) CdServerOption {
	return func(server *CdServer) {
		server.s3Info = NewCdS3Info(s3AK, s3SK, s3Endpoint, s3Bucket, s3Region, s3getToolUrl)
	}
}

func NewCdS3Info(s3AK, s3SK, s3Endpoint, s3Bucket, s3Region, s3getToolUrl string) *CdS3Info {
	return &CdS3Info{
		s3AK:         s3AK,
		s3SK:         s3SK,
		s3Endpoint:   s3Endpoint,
		s3Bucket:     s3Bucket,
		s3Region:     s3Region,
		s3GetToolUrl: s3getToolUrl,
	}
}

func NewCdNodeInfo(options ...CdNodeOption) *CdNodeInfo {
	cdNodeInfo := &CdNodeInfo{
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

type CdNodeOption func(*CdNodeInfo)

func CdNodeCredIdOption(credId string) CdNodeOption {
	return func(nodeInfo *CdNodeInfo) {
		nodeInfo.credentialsId = credId
	}
}

func CdNodeJvmOption(jvm string) CdNodeOption {
	return func(nodeInfo *CdNodeInfo) {
		nodeInfo.jvmOptions = jvm
	}
}

func CdNodeNumExecutorsOption(numExecutors int) CdNodeOption {
	return func(nodeInfo *CdNodeInfo) {
		nodeInfo.numExecutors = numExecutors
	}
}

func CdNodeRemoteFsOption(remoteFs string) CdNodeOption {
	return func(nodeInfo *CdNodeInfo) {
		nodeInfo.remoteFs = remoteFs
	}
}

func CdNodeSshPortOption(sshPort string) CdNodeOption {
	return func(nodeInfo *CdNodeInfo) {
		nodeInfo.sshPort = sshPort
	}
}
