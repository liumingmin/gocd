package gocd

type CdServerOption func(*CdServer)

func CdServerEnvOption(env string) CdServerOption {
	return func(server *CdServer) {
		server.env = env
	}
}

func CdServerNodeOption(options ...CdNodeOption) CdServerOption {
	return func(server *CdServer) {
		server.defCdNodeParam = NewCdNodeParam(options...)
	}
}
