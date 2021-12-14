package gocd

type CdServerOption func(*CdServer)

func CdServerEnvOption(env string) CdServerOption {
	return func(server *CdServer) {
		server.env = env
	}
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
